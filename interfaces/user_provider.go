package interfaces

import (
	"github.com/oarkflow/ftp-server/fs"
	"github.com/oarkflow/ftp-server/models"
)

type UserProvider interface {
	Login(user, pass string) (*fs.AuthenticationResponse, error)
	Register(user models.User)
}