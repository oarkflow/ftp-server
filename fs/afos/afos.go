// Package afos provide an afero OS FS access layer
package afos

import (
	"errors"
	
	"github.com/spf13/afero"
	
	"github.com/oarkflow/ftp-server/fs/utils"
	"github.com/oarkflow/ftp-server/models"
)

// ErrMissingBasePath is triggered when the basePath property isn't specified
var ErrMissingBasePath = errors.New("basePath must be specified")

// LoadFs loads a file system from an access description
func LoadFs(access *models.Access) (afero.Fs, error) {
	basePath := access.Params["basePath"]
	if basePath == "" {
		return nil, ErrMissingBasePath
	}
	
	basePath = utils.ReplaceEnvVars(basePath)
	
	return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
}
