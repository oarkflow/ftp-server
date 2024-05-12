package v2

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"

	"github.com/fclairamb/go-log"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/oarkflow/ftp-server/log/oarklog"
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/fs/afos"
	"github.com/oarkflow/ftp-server/v2/interfaces"
	"github.com/oarkflow/ftp-server/v2/models"
	"github.com/oarkflow/ftp-server/v2/providers"
	"github.com/oarkflow/ftp-server/v2/utils"
)

type Server struct {
	basePath            string
	sshPath             string
	dataPath            string
	privateKey          string
	publicKey           string
	readOnly            bool
	port                int
	address             string
	user                fs.OsUser
	userProvider        interfaces.UserProvider
	logger              log.Logger
	pathValidator       func(fs interfaces.Filesystem, p string) (string, error)
	diskSpaceValidator  func(fs interfaces.Filesystem) bool
	credentialValidator func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error)
}

func defaultServer() *Server {
	basePath := utils.AbsPath("")
	dataPath := "data"
	userProvider := providers.NewJsonFileProvider("sha256", "")
	return &Server{
		basePath:   basePath,
		dataPath:   dataPath,
		sshPath:    ".ssh",
		port:       2022,
		privateKey: "id_rsa",
		address:    "0.0.0.0",
		user: fs.OsUser{
			UID: 10,
			GID: 10,
		},
		logger:       oarklog.Default(),
		userProvider: userProvider,
		pathValidator: func(fs interfaces.Filesystem, p string) (string, error) {
			join := path.Join(basePath, dataPath, p)
			clean := path.Clean(path.Join(basePath, dataPath))
			if strings.HasPrefix(join, clean) {
				return join, nil
			}
			return "", errors.New("invalid path outside the configured directory was provided")
		},
		diskSpaceValidator: func(fs interfaces.Filesystem) bool {
			return true // TODO
		},
		credentialValidator: func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error) {
			return server.userProvider.Login(r.User, r.Pass)
		},
	}
}

func New(opts ...func(*Server)) *Server {
	svr := defaultServer()
	for _, o := range opts {
		o(svr)
	}
	return svr
}

func (c *Server) AddUser(user models.User) {
	c.userProvider.Register(user)
}

func (c *Server) Validate(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	resp, err := c.credentialValidator(c, fs.AuthenticationRequest{
		User:          conn.User(),
		Pass:          string(pass),
		IP:            conn.RemoteAddr().String(),
		SessionID:     conn.SessionID(),
		ClientVersion: conn.ClientVersion(),
	})

	if err != nil {
		return nil, err
	}

	sshPerm := &ssh.Permissions{
		Extensions: map[string]string{
			"uuid":        resp.Server,
			"user":        conn.User(),
			"permissions": strings.Join(resp.Permissions, ","),
		},
	}

	return sshPerm, nil
}

// Initialize the SFTP server and add a persistent listener to handle inbound SFTP connections.
func (c *Server) Initialize() error {
	serverConfig := &ssh.ServerConfig{
		NoClientAuth:     false,
		MaxAuthTries:     6,
		PasswordCallback: c.Validate,
	}

	if _, err := os.Stat(path.Join(c.basePath, c.sshPath, c.privateKey)); os.IsNotExist(err) {
		if err := c.generatePrivateKey(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	privateBytes, err := os.ReadFile(path.Join(c.basePath, c.sshPath, c.privateKey))
	if err != nil {
		return err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return err
	}

	// Add our private key to the server configuration.
	serverConfig.AddHostKey(private)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.address, c.port))
	if err != nil {
		return err
	}

	c.logger.Info("sftp subsystem listening for connections", "host", c.address, "port", c.port)

	for {
		conn, _ := listener.Accept()
		if conn != nil {
			go c.AcceptInboundConnection(conn, serverConfig)
		}
	}
}

// AcceptInboundConnection ... Handles an inbound connection to the instance and determines if
// we should serve the request  or not.
func (c *Server) AcceptInboundConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// Before beginning a handshake must be performed on the incoming net.Conn
	sconn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sconn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		// If its not a session channel we just move on because its not something we
		// know how to handle at this point.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		// Channels have a type that is dependent on the protocol. For SFTP this is "subsystem"
		// with a payload that (should) be "sftp". Discard anything else we receive ("pty", "shell", etc)
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false

				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}

				req.Reply(ok, nil)
			}
		}(requests)

		// Configure the user's home folder for the rest of the request cycle.
		if sconn.Permissions.Extensions["uuid"] == "" {
			continue
		}

		// Create a new handler for the currently logged in user's server.
		fs := c.createHandler(sconn.Permissions)

		// Create the server instance for the channel using the filesystem we created above.
		server := sftp.NewRequestServer(channel, fs)

		if err := server.Serve(); err == io.EOF {
			server.Close()
		}
	}
}

// Creates a new SFTP handler for a given server. The directory argument should
// be the base directory for a server. All actions done on the server will be
// relative to that directory, and the user will not be able to escape out of it.
func (c *Server) createHandler(perm *ssh.Permissions) sftp.Handlers {
	p := afos.Afos{
		UUID:          perm.Extensions["uuid"],
		Permissions:   strings.Split(perm.Extensions["permissions"], ","),
		ReadOnly:      c.readOnly,
		User:          c.user,
		HasDiskSpace:  c.diskSpaceValidator,
		PathValidator: c.pathValidator,
	}
	p.SetLogger(c.logger)
	return sftp.Handlers{
		FileGet:  &p,
		FilePut:  &p,
		FileCmd:  &p,
		FileList: &p,
	}
}

// Generates a private key that will be used by the SFTP server.
func (c *Server) generatePrivateKey() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(c.basePath, c.sshPath), 0755); err != nil {
		return err
	}

	o, err := os.OpenFile(path.Join(c.basePath, c.sshPath, c.privateKey), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer o.Close()

	pkey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(o, pkey); err != nil {
		return err
	}

	return nil
}
