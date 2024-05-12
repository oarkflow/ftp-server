package utils

import (
	"path/filepath"
)

const (
	// PermissionFileRead ... Permission to read a file.
	PermissionFileRead = "file.read"
	// PermissionFileReadContent ... Permission to read the contents of a file.
	PermissionFileReadContent = "file.read-content"
	// PermissionFileCreate ... Permission to create a file.
	PermissionFileCreate = "file.create"
	// PermissionFileUpdate ... Permission to update a file.
	PermissionFileUpdate = "file.update"
	// PermissionFileDelete ... Permission to delete a file.
	PermissionFileDelete = "file.delete"
)

func AbsPath(path string) string {
	if !filepath.IsAbs(path) {
		b, err := filepath.Abs(path)
		if err == nil {
			path = b
		}
	}
	return path
}
