package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/oarkflow/ftp-server"
	"github.com/oarkflow/ftp-server/fs/s3"
	"github.com/oarkflow/ftp-server/models"
)

type config struct {
	Address  string    `json:"address"`
	Filepath string    `json:"files"`
	Port     int       `json:"port"`
	ReadOnly bool      `json:"readOnly"`
	S3       s3.Option `json:"s3"`
}

func main() {
	var users map[string]models.User
	var conf config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(configFile, &conf)
	if err != nil {
		panic(err)
	}
	usersFile, err := os.ReadFile("users.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(usersFile, &users)
	if err != nil {
		panic(err)
	}
	callback := ftpserver.WithNotificationCallback(func(notification ftpserver.Notification) error {
		fmt.Println("Notification", notification)
		return nil
	})
	server := ftpserver.NewWithNotify(callback)
	for _, user := range users {
		server.AddUser(user)
	}
	panic(server.Initialize())
}
