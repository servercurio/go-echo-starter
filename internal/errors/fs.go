package errors

import e "github.com/joomcode/errorx"

var (
	FileSystemErrors = e.NewNamespace("filesystem")

	FileAccessDenied  = FileSystemErrors.NewType("file_access_denied")
	FileNotFound      = FileSystemErrors.NewType("file_not_found", e.NotFound())
	InvalidFilePath   = FileSystemErrors.NewType("invalid_file_path")
	IllegalFileFormat = FileSystemErrors.NewType("illegal_file_format")
)
