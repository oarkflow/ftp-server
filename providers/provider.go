package providers

import (
	"github.com/oarkflow/ftp-server/fs"
	"github.com/oarkflow/ftp-server/models"
)

var (
	DefaultPermissions = []string{"file.read", "file.read-content", "file.create", "file.update", "file.delete"}
)

type UserProvider interface {
	Login(user, pass string) (*fs.AuthenticationResponse, error)
	Register(user models.User)
}
