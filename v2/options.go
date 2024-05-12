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

func WithCredentialValidator(val func(server *Server, r fs.AuthenticationRequest) (*fs.AuthenticationResponse, error)) func(server *Server) {
	return func(o *Server) {
		o.credentialValidator = val
	}
}

func WithFilesystem(val interfaces.Filesystem) func(server *Server) {
	return func(o *Server) {
		o.fs = val
	}
}
