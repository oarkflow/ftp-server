package v2

import (
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/interfaces"
)

func WithUserProvider(provider interfaces.UserProvider) func(*Server) {
	return func(o *Server) {
		o.userProvider = provider
	}
}

func WithBasePath(path string) func(server *Server) {
	return func(o *Server) {
		o.basePath = path
	}
}

func WithReadOnly(val bool) func(server *Server) {
	return func(o *Server) {
		o.readOnly = val
	}
}

func WithPort(val int) func(server *Server) {
	return func(o *Server) {
		o.port = val
	}
}

func WithAddress(val string) func(server *Server) {
	return func(o *Server) {
		o.address = val
	}
}

func WithSSHPath(val string) func(server *Server) {
	return func(o *Server) {
		o.sshPath = val
	}
}

func WithDataPath(val string) func(server *Server) {
	return func(o *Server) {
		o.dataPath = val
	}
}

func WithPrivateKey(val string) func(server *Server) {
	return func(o *Server) {
		o.privateKey = val
	}
}

func WithPublicKey(val string) func(server *Server) {
	return func(o *Server) {
		o.publicKey = val
	}
}

func WithOsUser(val fs.OsUser) func(server *Server) {
	return func(o *Server) {
		o.user = val
	}
}

func WithDiskSpaceValidator(val func(fs interfaces.Filesystem) bool) func(server *Server) {
	return func(o *Server) {
		o.diskSpaceValidator = val
	}
}

func WithPathValidator(val func(fs interfaces.Filesystem, p string) (string, error)) func(server *Server) {
	return func(o *Server) {
		o.pathValidator = val
	}
}

func WithCredentialValidator(val func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error)) func(server *Server) {
	return func(o *Server) {
		o.credentialValidator = val
	}
}
