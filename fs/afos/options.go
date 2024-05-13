package afos

import (
	"github.com/oarkflow/ftp-server/fs"
)

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

func WithDiskSpaceValidator(val func(fs fs.FS) bool) func(server *Afos) {
	return func(o *Afos) {
		o.hasDiskSpace = val
	}
}

func WithPathValidator(val func(fs fs.FS, p string) (string, error)) func(server *Afos) {
	return func(o *Afos) {
		o.pathValidator = val
	}
}
