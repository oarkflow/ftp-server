package afos

import (
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/interfaces"
)

func WithOsUser(val fs.OsUser) func(server *Afos) {
	return func(o *Afos) {
		o.osUser = val
	}
}

func WithDataPath(val string) func(server *Afos) {
	return func(o *Afos) {
		o.dataPath = val
	}
}

func WithPermissions(val []string) func(server *Afos) {
	return func(o *Afos) {
		o.permissions = val
	}
}

func WithReadOnly(val bool) func(server *Afos) {
	return func(o *Afos) {
		o.readOnly = val
	}
}

func WithDiskSpaceValidator(val func(fs interfaces.Filesystem) bool) func(server *Afos) {
	return func(o *Afos) {
		o.hasDiskSpace = val
	}
}

func WithPathValidator(val func(fs interfaces.Filesystem, p string) (string, error)) func(server *Afos) {
	return func(o *Afos) {
		o.pathValidator = val
	}
}
