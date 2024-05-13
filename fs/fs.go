package fs

import (
	"io"
	"os"
	"slices"

	"github.com/oarkflow/ftp-server/models"
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
	Server string      `json:"server"`
	Token  string      `json:"token"`
	User   models.User `json:"user"`
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

// Can - Determines if a user has permission to perform a specific action on the SFTP server. These
// permissions are defined and returned by the Panel API.
func Can(permissions []string, permission string) bool {
	// Server owners and super admins have their permissions returned as '[*]' via the Panel
	// API, so for the sake of speed do an initial check for that before iterating over the
	// entire array of permissions.
	if len(permissions) == 1 && permissions[0] == "*" {
		return true
	}
	return slices.Contains(permissions, permission)
}
