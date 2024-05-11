package sftp

import (
	"io"
	"os"
)

// AuthenticationRequest ... An authentication request to the SFTP server.
type AuthenticationRequest struct {
	User          string `json:"username"`
	Pass          string `json:"password"`
	IP            string `json:"ip"`
	SessionID     []byte `json:"session_id"`
	ClientVersion []byte `json:"client_version"`
}

// AuthenticationResponse ... An authentication response from the SFTP server.
type AuthenticationResponse struct {
	Server      string   `json:"server"`
	Token       string   `json:"token"`
	Permissions []string `json:"permissions"`
}

// InvalidCredentialsError ... An error emitted when credentials are invalid.
type InvalidCredentialsError struct {
}

func (ice InvalidCredentialsError) Error() string {
	return "the credentials provided were invalid"
}

// IsInvalidCredentialsError ... Checks if an error is a InvalidCredentialsError.
func IsInvalidCredentialsError(err error) bool {
	_, ok := err.(*InvalidCredentialsError)

	return ok
}

type fxerr uint32

const (
	// ErrSSHQuotaExceeded ...
	// Extends the default SFTP server to return a quota exceeded error to the client.
	//
	// @see https://tools.ietf.org/id/draft-ietf-secsh-filexfer-13.txt
	ErrSSHQuotaExceeded = fxerr(15)
)

func (e fxerr) Error() string {
	switch e {
	case ErrSSHQuotaExceeded:
		return "Quota Exceeded"
	default:
		return "Failure"
	}
}

// ListerAt ... A list of files.
type ListerAt []os.FileInfo

// ListAt ...
// Returns the number of entries copied and an io.EOF error if we made it to the end of the file list.
// Take a look at the pkg/sftp godoc for more information about how this function should work.
func (l ListerAt) ListAt(f []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}

	n := copy(f, l[offset:])
	if n < len(f) {
		return n, io.EOF
	}
	return n, nil
}
