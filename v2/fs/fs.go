package fs

import (
	"io"
	"os"
)

type OsUser struct {
	UID int
	GID int
}

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
