package interfaces

import (
	"io"

	"github.com/pkg/sftp"

	"github.com/oarkflow/ftp-server/log"
)

type Filesystem interface {
	Fileread(request *sftp.Request) (io.ReaderAt, error)
	Filewrite(request *sftp.Request) (io.WriterAt, error)
	Filecmd(request *sftp.Request) error
	Filelist(request *sftp.Request) (sftp.ListerAt, error)
	SetLogger(logger log.Logger)
	SetPermissions(p []string)
	SetID(p string)
	Type() string
}
