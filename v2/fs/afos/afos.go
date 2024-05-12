package afos

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/fclairamb/go-log"
	"github.com/pkg/sftp"
	"go.uber.org/zap"

	"github.com/oarkflow/ftp-server/v2/errs"
	"github.com/oarkflow/ftp-server/v2/fs"
	"github.com/oarkflow/ftp-server/v2/interfaces"
	"github.com/oarkflow/ftp-server/v2/utils"
)

// Afos ... A file system exposed to a user.
type Afos struct {
	UUID          string
	Permissions   []string
	ReadOnly      bool
	User          fs.OsUser
	PathValidator func(fs interfaces.Filesystem, p string) (string, error)
	HasDiskSpace  func(fs interfaces.Filesystem) bool
	lock          sync.Mutex
	logger        log.Logger
}

func (f *Afos) buildPath(p string) (string, error) {
	if f.PathValidator == nil {
		return "", nil
	}
	return f.PathValidator(f, p)
}

func (f *Afos) SetLogger(logger log.Logger) {
	f.logger = logger
}

// Fileread creates a reader for a file on the system and returns the reader back.
func (f *Afos) Fileread(request *sftp.Request) (io.ReaderAt, error) {
	// Check first if the user can actually open and view a file. This permission is named
	// really poorly, but it is checking if they can read. There is an addition permission,
	// "save-files" which determines if they can write that file.
	if !f.can(utils.PermissionFileReadContent) {
		return nil, sftp.ErrSshFxPermissionDenied
	}

	p, err := f.buildPath(request.Filepath)
	if err != nil {
		return nil, sftp.ErrSshFxNoSuchFile
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, sftp.ErrSshFxNoSuchFile
	}

	file, err := os.Open(p)
	if err != nil {
		f.logger.Error("could not open file for reading", "source", p, "err", err)
		return nil, sftp.ErrSshFxFailure
	}

	return file, nil
}

// Filewrite handles the write actions for a file on the system.
func (f *Afos) Filewrite(request *sftp.Request) (io.WriterAt, error) {
	if f.ReadOnly {
		return nil, sftp.ErrSshFxOpUnsupported
	}

	p, err := f.buildPath(request.Filepath)
	if err != nil {
		return nil, sftp.ErrSshFxNoSuchFile
	}

	// If the user doesn't have enough space left on the server it should respond with an
	// error since we won't be letting them write this file to the disk.
	if !f.HasDiskSpace(f) {
		return nil, errs.ErrSSHQuotaExceeded
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	stat, statErr := os.Stat(p)
	// If the file doesn't exist we need to create it, as well as the directory pathway
	// leading up to where that file will be created.
	if os.IsNotExist(statErr) {
		// This is a different pathway than just editing an existing file. If it doesn't exist already
		// we need to determine if this user has permission to create files.
		if !f.can(utils.PermissionFileCreate) {
			return nil, sftp.ErrSshFxPermissionDenied
		}

		// Create all of the directories leading up to the location where this file is being created.
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			f.logger.Error("error making path for file",
				"source", p,
				"path", filepath.Dir(p),
				"err", err,
			)
			return nil, sftp.ErrSshFxFailure
		}

		file, err := os.Create(p)
		if err != nil {
			f.logger.Error("error creating file", "source", p, "err", err)
			return nil, sftp.ErrSshFxFailure
		}

		// Not failing here is intentional. We still made the file, it is just owned incorrectly
		// and will likely cause some issues.
		if err := os.Chown(p, f.User.UID, f.User.GID); err != nil {
			f.logger.Warn("error chowning file", "file", p, "err", err)
		}

		return file, nil
	}

	// If the stat error isn't about the file not existing, there is some other issue
	// at play and we need to go ahead and bail out of the process.
	if statErr != nil {
		f.logger.Error("error performing file stat", "source", p, "err", err)
		return nil, sftp.ErrSshFxFailure
	}

	// If we've made it here it means the file already exists and we don't need to do anything
	// fancy to handle it. Just pass over the request flags so the system knows what the end
	// goal with the file is going to be.
	//
	// But first, check that the user has permission to save modified files.
	if !f.can(utils.PermissionFileUpdate) {
		return nil, sftp.ErrSshFxPermissionDenied
	}

	// Not sure this would ever happen, but lets not find out.
	if stat.IsDir() {
		f.logger.Warn("attempted to open a directory for writing to", "source", p)
		return nil, sftp.ErrSshFxOpUnsupported
	}

	file, err := os.Create(p)
	if err != nil {
		f.logger.Error("error opening existing file",
			zap.Uint32("flags", request.Flags),
			zap.String("source", p),
			zap.Error(err),
		)
		return nil, sftp.ErrSshFxFailure
	}

	// Not failing here is intentional. We still made the file, it is just owned incorrectly
	// and will likely cause some issues.
	if err := os.Chown(p, f.User.UID, f.User.GID); err != nil {
		f.logger.Warn("error chowning file", "file", p, zap.Error(err))
	}

	return file, nil
}

// Filecmd hander for basic SFTP system calls related to files, but not anything to do with reading
// or writing to those files.
func (f *Afos) Filecmd(request *sftp.Request) error {
	if f.ReadOnly {
		return sftp.ErrSshFxOpUnsupported
	}

	p, err := f.buildPath(request.Filepath)
	if err != nil {
		return sftp.ErrSshFxNoSuchFile
	}

	var target string
	// If a target is provided in this request validate that it is going to the correct
	// location for the server. If it is not, return an operation unsupported error. This
	// is maybe not the best error response, but its not wrong either.
	if request.Target != "" {
		target, err = f.buildPath(request.Target)
		if err != nil {
			return sftp.ErrSshFxOpUnsupported
		}
	}

	switch request.Method {
	case "Setstat":
		if !f.can(utils.PermissionFileUpdate) {
			return sftp.ErrSshFxPermissionDenied
		}

		var mode os.FileMode = 0644
		// If the client passed a valid file permission use that, otherwise use the
		// default of 0644 set above.
		if request.Attributes().FileMode().Perm() != 0000 {
			mode = request.Attributes().FileMode().Perm()
		}

		// Force directories to be 0755
		if request.Attributes().FileMode().IsDir() {
			mode = 0755
		}

		if err := os.Chmod(p, mode); err != nil {
			f.logger.Error("failed to perform setstat", "err", err)
			return sftp.ErrSshFxFailure
		}
		return nil
	case "Rename":
		if !f.can(utils.PermissionFileUpdate) {
			return sftp.ErrSshFxPermissionDenied
		}

		if err := os.Rename(p, target); err != nil {
			f.logger.Error("failed to rename file",
				"source", p,
				"target", target,
				"err", err,
			)
			return sftp.ErrSshFxFailure
		}

		break
	case "Rmdir":
		if !f.can(utils.PermissionFileDelete) {
			return sftp.ErrSshFxPermissionDenied
		}

		if err := os.RemoveAll(p); err != nil {
			f.logger.Error("failed to remove directory", "source", p, "err", err)
			return sftp.ErrSshFxFailure
		}

		return sftp.ErrSshFxOk
	case "Mkdir":
		if !f.can(utils.PermissionFileCreate) {
			return sftp.ErrSshFxPermissionDenied
		}

		if err := os.MkdirAll(p, 0755); err != nil {
			f.logger.Error("failed to create directory", "source", p, "err", err)
			return sftp.ErrSshFxFailure
		}

		break
	case "Symlink":
		if !f.can(utils.PermissionFileCreate) {
			return sftp.ErrSshFxPermissionDenied
		}

		if err := os.Symlink(p, target); err != nil {
			f.logger.Error("failed to create symlink",
				"source", p, "err", err,
				"target", target,
			)
			return sftp.ErrSshFxFailure
		}

		break
	case "Remove":
		if !f.can(utils.PermissionFileDelete) {
			return sftp.ErrSshFxPermissionDenied
		}

		if err := os.Remove(p); err != nil {
			if !os.IsNotExist(err) {
				f.logger.Error("failed to remove a file", "source", p, "err", err)
			}
			return sftp.ErrSshFxFailure
		}

		return sftp.ErrSshFxOk
	default:
		return sftp.ErrSshFxOpUnsupported
	}

	var fileLocation = p
	if target != "" {
		fileLocation = target
	}

	// Not failing here is intentional. We still made the file, it is just owned incorrectly
	// and will likely cause some issues. There is no logical check for if the file was removed
	// because both of those cases (Rmdir, Remove) have an explicit return rather than break.
	if err := os.Chown(fileLocation, f.User.UID, f.User.GID); err != nil {
		f.logger.Warn("error chowning file", "file", fileLocation, "err", err)
	}

	return sftp.ErrSshFxOk
}

// Filelist is the handler for SFTP filesystem list calls. This will handle calls to list the contents of
// a directory as well as perform file/folder stat calls.
func (f *Afos) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	p, err := f.buildPath(request.Filepath)
	if err != nil {
		return nil, sftp.ErrSshFxNoSuchFile
	}

	switch request.Method {
	case "List":
		if !f.can(utils.PermissionFileRead) {
			return nil, sftp.ErrSshFxPermissionDenied
		}

		files, err := ioutil.ReadDir(p)
		if err != nil {
			f.logger.Error("error listing directory", zap.Error(err))
			return nil, sftp.ErrSshFxFailure
		}

		return fs.ListerAt(files), nil
	case "Stat":
		if !f.can(utils.PermissionFileRead) {
			return nil, sftp.ErrSshFxPermissionDenied
		}

		s, err := os.Stat(p)
		if os.IsNotExist(err) {
			return nil, sftp.ErrSshFxNoSuchFile
		} else if err != nil {
			f.logger.Error("error running STAT on file", zap.Error(err))
			return nil, sftp.ErrSshFxFailure
		}

		return fs.ListerAt([]os.FileInfo{s}), nil
	default:
		// Before adding readlink support we need to evaluate any potential security risks
		// as a result of navigating around to a location that is outside the home directory
		// for the logged in user. I don't forsee it being much of a problem, but I do want to
		// check it out before slapping some code here. Until then, we'll just return an
		// unsupported response code.
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

// Determines if a user has permission to perform a specific action on the SFTP server. These
// permissions are defined and returned by the Panel API.
func (f *Afos) can(permission string) bool {
	// Server owners and super admins have their permissions returned as '[*]' via the Panel
	// API, so for the sake of speed do an initial check for that before iterating over the
	// entire array of permissions.
	if len(f.Permissions) == 1 && f.Permissions[0] == "*" {
		return true
	}
	return slices.Contains(f.Permissions, permission)
}
