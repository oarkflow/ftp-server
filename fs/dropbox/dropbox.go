// Package dropbox provides a Dropbox layer
package dropbox

import (
	"errors"
	"os"
	
	dropbox "github.com/fclairamb/afero-dropbox"
	"github.com/spf13/afero"
	
	"github.com/oarkflow/ftp-server/models"
)

// ErrMissingToken is returned if a dropbox token wasn't specified.
var ErrMissingToken = errors.New("missing token")

// LoadFs loads a file system from an access description
func LoadFs(access *models.Access) (afero.Fs, error) {
	token := access.Params["token"]
	
	if token == "" {
		token = os.Getenv("DROPBOX_TOKEN")
	}
	
	if token == "" {
		return nil, ErrMissingToken
	}
	
	return dropbox.NewFs(token), nil
}
