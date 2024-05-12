package main

import (
	"encoding/json"
	"os"

	v2 "github.com/oarkflow/ftp-server/v2"
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/fs/afos"
	"github.com/oarkflow/ftp-server/v2/models"
	"github.com/oarkflow/ftp-server/v2/utils"
)

type config struct {
	ReadOnly    bool                   `json:"readOnly"`
	Port        int                    `json:"port"`
	BindAddress string                 `json:"bind"`
	Filepath    string                 `json:"files"`
	User        fs.OsUser              `json:"osUser"`
	Users       map[string]models.User `json:"users"`
}

func main() {
	// Read the config.json.
	var conf config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(configFile, &conf)
	fs := afos.New(utils.AbsPath(""))
	server := v2.New(fs)
	for _, user := range conf.Users {
		server.AddUser(user)
	}
	panic(server.Initialize())
}
