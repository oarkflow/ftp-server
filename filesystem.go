package ftpserverlib

import (
	"io"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/oarkflow/ftp-server/log"

	"github.com/oarkflow/ftp-server/interfaces"
)

type FS struct {
	fs interfaces.Filesystem
}

func NewFS(fs interfaces.Filesystem) interfaces.Filesystem {
	return &FS{fs: fs}
}

func (f *FS) SetContext(ctx map[string]string) {
	f.fs.SetContext(ctx)
}

func (f *FS) Context() map[string]string {
	return f.fs.Context()
}

func (f *FS) Notify(request *sftp.Request, err error) {
	method := request.Method
	f.fs.Logger().Info("Triggered", method, "file", request.Filepath, "target", request.Target, "error", err)
}

func (f *FS) Fileread(request *sftp.Request) (io.ReaderAt, error) {
	var err error
	defer func() {
		f.Notify(request, err)
	}()
	rs, e := f.fs.Fileread(request)
	err = e
	return rs, e
}

func (f *FS) Filewrite(request *sftp.Request) (io.WriterAt, error) {
	var err error
	defer func() {
		f.Notify(request, err)
	}()
	rs, e := f.fs.Filewrite(request)
	err = e
	return rs, e
}

func (f *FS) Filecmd(request *sftp.Request) error {
	var err error
	defer func() {
		f.Notify(request, err)
	}()
	e := f.fs.Filecmd(request)
	err = e
	return e
}

func (f *FS) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	var err error
	defer func() {
		f.Notify(request, err)
	}()
	rs, e := f.fs.Filelist(request)
	err = e
	return rs, e
}

func (f *FS) SetLogger(logger log.Logger) {
	f.fs.SetLogger(logger)
}

func (f *FS) Logger() log.Logger {
	return f.fs.Logger()
}

func (f *FS) SetPermissions(p []string) {
	f.fs.SetPermissions(p)
}

func (f *FS) Permissions() []string {
	return f.fs.Permissions()
}

func (f *FS) SetConn(sconn *ssh.ServerConn) {
	f.fs.SetConn(sconn)
}

func (f *FS) Conn() *ssh.ServerConn {
	return f.fs.Conn()
}

func (f *FS) SetID(p string) {
	f.fs.SetID(p)
}

func (f *FS) Type() string {
	return f.fs.Type()
}
