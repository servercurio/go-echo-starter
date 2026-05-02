package config

import (
	"os"
	"path/filepath"
	"strings"

	ex "github.com/joomcode/errorx"
	"github.com/servercurio/go-echo-starter/internal/errors"
)

type validFilePathFn func(stat os.FileInfo) bool

var directoryCheck = func(stat os.FileInfo) bool {
	return stat.IsDir()
}

var fileCheck = func(stat os.FileInfo) bool {
	return !stat.IsDir()
}

func FileNameVariants(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return []string{}
	}

	name = strings.TrimSuffix(name, filepath.Ext(name))

	return []string{
		name + ".json",
		name + ".yml",
		name + ".yaml",
	}
}

func checkPath(file string, checkFn validFilePathFn) (bool, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return false, ex.IllegalArgument.New("file path is empty")
	}

	stat, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.FileNotFound.Wrap(err, "file not found: %s", file)
		} else if os.IsPermission(err) {
			return false, errors.FileAccessDenied.Wrap(err, "permission denied: %s", file)
		}

		return false, ex.ExternalError.Wrap(err, "failed to stat file: %s", file)
	}

	return checkFn(stat), nil
}

func isYamlFile(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".yaml" || ext == ".yml"
}

func isJsonFile(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".json"
}
