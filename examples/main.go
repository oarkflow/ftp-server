package main

import (
	"encoding/json"
	"os"
	
	v2 "github.com/oarkflow/ftp-server"
	"github.com/oarkflow/ftp-server/fs/afos"
	"github.com/oarkflow/ftp-server/fs/s3"
	"github.com/oarkflow/ftp-server/models"
	"github.com/oarkflow/ftp-server/utils"
)

type config struct {
	Users    map[string]models.User `json:"users"`
	Address  string                 `json:"address"`
	Filepath string                 `json:"files"`
	Port     int                    `json:"port"`
	ReadOnly bool                   `json:"readOnly"`
	S3       s3.Option              `json:"s3"`
}

func main() {
	// Read the config.json.
	var conf config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(configFile, &conf)
	if err != nil {
		panic(err)
	}
	filesystem := afos.New(utils.AbsPath(conf.Filepath))
	// filesystem, err := s3.New(s3Opt)
	if err != nil {
		panic(err)
	}
	server := v2.NewWithNotify(filesystem)
	for _, user := range conf.Users {
		server.AddUser(user)
	}
	panic(server.Initialize())
}
