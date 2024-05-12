package ftpserverlib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/oarkflow/ftp-server/log"

	"github.com/oarkflow/ftp-server/fs"
	"github.com/oarkflow/ftp-server/fs/afos"
	interfaces2 "github.com/oarkflow/ftp-server/interfaces"
	"github.com/oarkflow/ftp-server/log/oarklog"
	"github.com/oarkflow/ftp-server/models"
	"github.com/oarkflow/ftp-server/providers"
	"github.com/oarkflow/ftp-server/utils"
)

type Server struct {
	userProvider        interfaces2.UserProvider
	fs                  interfaces2.Filesystem
	logger              log.Logger
	credentialValidator func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error)
	basePath            string
	sshPath             string
	privateKey          string
	publicKey           string
	address             string
	port                int
}

func defaultServer(filesystem interfaces2.Filesystem) *Server {
	basePath := utils.AbsPath("")
	userProvider := providers.NewJsonFileProvider("sha256", "")
	return &Server{
		basePath:     basePath,
		sshPath:      ".ssh",
		port:         2022,
		privateKey:   "id_rsa",
		address:      "0.0.0.0",
		logger:       oarklog.Default(),
		fs:           filesystem,
		userProvider: userProvider,
		credentialValidator: func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error) {
			return server.userProvider.Login(r.User, r.Pass)
		},
	}
}

func New(filesystem interfaces2.Filesystem, opts ...func(*Server)) *Server {
	svr := defaultServer(filesystem)
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
	config, err := c.setupSSH()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.address, c.port))
	if err != nil {
		return err
	}

	c.logger.Info("sftp subsystem listening for connections", "host", c.address, "port", c.port, "fs_type", c.fs.Type())

	for {
		conn, _ := listener.Accept()
		if conn != nil {
			go c.AcceptInboundConnection(conn, config)
		}
	}
}

// AcceptInboundConnection ... Handles an inbound connection to the instance and determines if
// we should serve the request or not.
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
		handlers := c.createHandler(sconn.Permissions)

		// Create the server instance for the channel using the filesystem we created above.
		server := sftp.NewRequestServer(channel, handlers)

		if err := server.Serve(); err == io.EOF {
			server.Close()
		}
	}
}

// Creates a new SFTP handler for a given server. The directory argument should
// be the base directory for a server. All actions done on the server will be
// relative to that directory, and the user will not be able to escape out of it.
func (c *Server) createHandler(perm *ssh.Permissions) sftp.Handlers {
	if c.fs == nil {
		c.fs = afos.New(c.basePath)
	}
	c.fs.SetPermissions(strings.Split(perm.Extensions["permissions"], ","))
	c.fs.SetID(perm.Extensions["uuid"])
	c.fs.SetLogger(c.logger)
	return sftp.Handlers{
		FileGet:  c.fs,
		FilePut:  c.fs,
		FileCmd:  c.fs,
		FileList: c.fs,
	}
}

func (c *Server) getSSHPath() string {
	return path.Join(c.basePath, c.sshPath, c.privateKey)
}

func (c *Server) setupSSH() (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		NoClientAuth:     false,
		MaxAuthTries:     6,
		PasswordCallback: c.Validate,
	}
	if _, err := os.Stat(c.getSSHPath()); os.IsNotExist(err) {
		if err := c.generatePrivateKey(); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	privateBytes, err := os.ReadFile(c.getSSHPath())
	if err != nil {
		return nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	// Add our private key to the server configuration.
	config.AddHostKey(private)
	return config, nil
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

	o, err := os.OpenFile(c.getSSHPath(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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
