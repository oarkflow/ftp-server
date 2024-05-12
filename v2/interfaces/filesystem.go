package interfaces

import (
	"io"

	"github.com/fclairamb/go-log"
	"github.com/pkg/sftp"
)

type Filesystem interface {
	Fileread(request *sftp.Request) (io.ReaderAt, error)
	Filewrite(request *sftp.Request) (io.WriterAt, error)
	Filecmd(request *sftp.Request) error
	Filelist(request *sftp.Request) (sftp.ListerAt, error)
	SetLogger(logger log.Logger)
}
