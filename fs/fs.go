package fs

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/fclairamb/go-log"

	"github.com/oarkflow/ftp-server/fs/afos"
	"github.com/oarkflow/ftp-server/fs/dropbox"
	"github.com/oarkflow/ftp-server/fs/gdrive"
	"github.com/oarkflow/ftp-server/fs/mail"
	"github.com/oarkflow/ftp-server/fs/s3"
	"github.com/oarkflow/ftp-server/fs/sftp"
	"github.com/oarkflow/ftp-server/fs/snd"
	"github.com/oarkflow/ftp-server/models"
)

// UnsupportedFsError is returned when the described file system is not supported
type UnsupportedFsError struct {
	error
	Type string
}

func (err UnsupportedFsError) Error() string {
	return fmt.Sprintf("Unsupported FS: %s", err.Type)
}

// LoadFs loads a file system from an access description
func LoadFs(access *models.Access, logger log.Logger) (afero.Fs, error) {
	var fs afero.Fs
	var err error
	switch access.Fs {
	case "os":
		fs, err = afos.LoadFs(access)
	case "s3":
		fs, err = s3.LoadFs(access)
	case "sftp":
		fs, err = sftp.LoadFs(access)
	case "mail":
		fs, err = mail.LoadFs(access)
	case "gdrive":
		fs, err = gdrive.LoadFs(access, logger.With("component", "gdrive"))
	case "dropbox":
		fs, err = dropbox.LoadFs(access)
	default:
		fs, err = nil, &UnsupportedFsError{Type: access.Fs}
	}

	if err == nil && access.ReadOnly {
		fs = afero.NewReadOnlyFs(fs)
	}

	// If we're defining a dubious behavior, we can use it
	if err == nil && access != nil && access.SyncAndDelete != nil && access.SyncAndDelete.Enable {
		var temp afero.Fs

		if access.SyncAndDelete.Directory != "" {
			temp = afero.NewBasePathFs(afero.NewOsFs(), access.SyncAndDelete.Directory)
		}

		fs, err = snd.NewFs(&snd.Config{
			Destination: fs,
			Temporary:   temp,
			Logger:      logger.With("component", "snd"),
		})
	}

	return fs, err
}
