package interfaces

import (
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/models"
)

type UserProvider interface {
	Login(user, pass string) (*fs.AuthenticationResponse, error)
	Register(user models.User)
}
