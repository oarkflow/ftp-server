package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"strings"

	"github.com/oarkflow/ftp-server/sftp"
)

type config struct {
	ReadOnly    bool   `json:"readOnly"`
	Port        int    `json:"port"`
	BindAddress string `json:"bind"`
	Filepath    string `json:"files"`
	User        struct {
		UID int `json:"uid"`
		GID int `json:"gid"`
	} `json:"osUser"`
	Users map[string]struct {
		Password    string   `json:"password"`
		Permissions []string `json:"permissions"`
	} `json:"users"`
}

func main() {
	// Read the config.json.
	var conf config
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(configFile, &conf)

	server := &sftp.Server{
		Settings: sftp.Settings{
			ReadOnly:    conf.ReadOnly,
			BindAddress: conf.BindAddress,
			BindPort:    conf.Port,
			BasePath:    conf.Filepath,
		},
		User: sftp.User{ // The SFTP server runs `chown` to this user on any uploaded file.
			UID: conf.User.UID,
			GID: conf.User.GID,
		},
		PathValidator: func(fs *sftp.FileSystem, p string) (string, error) {
			join := path.Join(conf.Filepath, p)
			clean := path.Clean(conf.Filepath)
			if strings.HasPrefix(join, clean) {
				return join, nil
			}
			return "", errors.New("invalid path outside the configured directory was provided")
		},
		DiskSpaceValidator: func(fs *sftp.FileSystem) bool {
			return true // TODO
		},
		CredentialValidator: func(r sftp.AuthenticationRequest) (*sftp.AuthenticationResponse, error) {
			user, exists := conf.Users[r.User]
			if !exists || fmt.Sprintf("%x", sha256.Sum256([]byte(r.Pass))) != user.Password {
				return nil, sftp.InvalidCredentialsError{}
			}
			n, _ := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
			return &sftp.AuthenticationResponse{
				Server:      "none",
				Token:       n.String(),
				Permissions: user.Permissions,
			}, nil
		},
	}
	err = sftp.New(server)
	if err != nil {
		panic(err)
	}
	panic(server.Initialize())
}
