package directory

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	directory string
}
type IHandler interface {
	Find() ([]string, error)
}

func NewDirectoryHandler(directory string) IHandler {
	return &Handler{
		directory: directory,
	}
}
func (h Handler) validate() error {
	if _, err := os.Stat(h.directory); os.IsNotExist(err) {
		return err
	}
	return nil
}
func (h Handler) Find() ([]string, error) {
	if err := h.validate(); err != nil {
		return nil, err
	}
	composeFilePaths := []string{}
	err := filepath.Walk(h.directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// don't panic on inaccessible paths
			if errors.Is(err, fs.ErrPermission) {
				return nil
			}
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".yml" {
			if strings.HasPrefix(info.Name(), "docker-compose") {
				composeFilePaths = append(composeFilePaths, path)
			}
		}
		return nil
	})
	return composeFilePaths, err
}
