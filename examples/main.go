package main

import (
	"encoding/json"
	"os"

	v2 "github.com/oarkflow/ftp-server"
	"github.com/oarkflow/ftp-server/fs/s3"
	"github.com/oarkflow/ftp-server/models"
)

type config struct {
	Users       map[string]models.User `json:"users"`
	BindAddress string                 `json:"bind"`
	Filepath    string                 `json:"files"`
	Port        int                    `json:"port"`
	ReadOnly    bool                   `json:"readOnly"`
}

func main() {
	// Read the config.json.
	var conf config
	var s3Opt s3.Option
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(configFile, &conf)
	if err != nil {
		panic(err)
	}
	s3Config, err := os.ReadFile("creds.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(s3Config, &s3Opt)
	if err != nil {
		panic(err)
	}
	// filesystem := afos.New(utils.AbsPath(""))
	filesystem, err := s3.New(s3Opt)
	if err != nil {
		panic(err)
	}
	server := v2.NewWithNotify(filesystem)
	for _, user := range conf.Users {
		server.AddUser(user)
	}
	panic(server.Initialize())
}
