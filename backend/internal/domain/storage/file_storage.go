package storage

import (
	"io"

	pkgerrors "backend/pkg/errors"
)

type FileInput struct {
	Name   string
	Reader io.Reader
	Size   int64
}

type FileInfo struct {
	Name string
	Size int64
}

type FileStorage interface {
	SaveFiles(dirPath string, files []FileInput) *pkgerrors.Error
	DeleteDir(dirPath string) *pkgerrors.Error
	ListFiles(dirPath string) ([]FileInfo, *pkgerrors.Error)
	ReadFile(filePath string) ([]byte, *pkgerrors.Error)
}
