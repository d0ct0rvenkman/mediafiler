package paths

import (
	"errors"
	"fmt"
	"os"

	multierr "github.com/hashicorp/go-multierror"
)

const (
	E_VALIDATE_PERMS    string = "permission to access path was denied"
	E_VALIDATE_NO_EXIST string = "path does not exist"
	E_AVAIL_PERMS       string = "permission denied"
	E_AVAIL_PATH_EXISTS string = "path exists"
	E_AVAIL_UNKNOWN     string = "unknown error from validatePath()"
)

/*
GetMediaPaths() returns the working and media destination directories
based on command line arguments and/or environment variables

returns workDir, destRootDir, err
*/
func GetMediaPaths(args []string) (string, string, error) {
	var workDir string
	var destRootDir string
	var err error

	badPath := "UNDEFINED"
	workDir = badPath
	destRootDir = badPath
	err = nil

	// TODO: we'll work up to stuff other than positional arguments later
	argc := len(args)
	switch {
	case argc > 1:
		destRootDir = args[1]
		fallthrough
	case argc > 0:
		workDir = args[0]
	default:

	}

	if workDir == badPath {
		err = multierr.Append(err, fmt.Errorf("working directory was not found in arguments"))
	}

	if destRootDir == badPath {
		err = multierr.Append(err, fmt.Errorf("destination root directory was not found in arguments"))
	}

	return workDir, destRootDir, err
}

/*
	validatePath determines whether the given path exists, and is minimally readable

	input[0]: path, string
	output[0]: nil if directory is okay, error message if not
*/

func validatePath(path string) (os.FileInfo, error) {
	var err error
	var info os.FileInfo

	info, err = os.Stat(path)

	if os.IsPermission(err) {
		err = errors.New("permission to access path was denied")
		return info, err
	}

	if os.IsNotExist(err) {
		err = errors.New("path does not exist")
		return info, err
	}

	return info, nil
}

/*
	ValidateDirectory determines whether the given path exists, is a directory, and is minimally readable

	input[0]: path, string
	output[0]: nil if path is okay, error message if not
*/

func ValidateDirectory(path string) error {
	var err error

	info, err := validatePath(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		err = fmt.Errorf("path '%s' is not a directory", path)
		return err
	}

	return nil
}

/*
	ValidateFileOrDirectory determines whether the given path exists, is a directory OR a file, and is minimally readable

	input[0]: path, string
	output[0]: nil if path is okay, error message if not
*/

func ValidateFileOrDirectory(path string) error {
	var err error

	info, err := validatePath(path)
	if err != nil {
		return err
	}

	if !info.Mode().IsDir() && !info.Mode().IsRegular() {
		err = errors.New("path is not a directory or a file")
		return err
	}

	return nil
}

func IsPathAvailable(path string) (bool, os.FileInfo, error) {
	info, err := validatePath((path))

	if err != nil {
		switch err.Error() {
		case E_VALIDATE_NO_EXIST:
			return true, info, nil
		case E_VALIDATE_PERMS:
			return false, info, errors.New(E_AVAIL_PERMS)
		default:
			return false, info, errors.New(E_AVAIL_UNKNOWN)
		}
	} else {
		// path exists in some shape or form
		return false, info, errors.New(E_AVAIL_PATH_EXISTS)
	}

}
