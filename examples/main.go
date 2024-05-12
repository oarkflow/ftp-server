package main

import (
	"encoding/json"
	"os"

	v2 "github.com/oarkflow/ftp-server/v2"
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/models"
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

	server := v2.New()
	for _, user := range conf.Users {
		server.AddUser(user)
	}
	panic(server.Initialize())
}
