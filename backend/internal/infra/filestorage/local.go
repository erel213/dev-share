package filestorage

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain/storage"
	pkgerrors "backend/pkg/errors"
)

type LocalFileStorage struct {
	basePath string
}

func NewLocalFileStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{basePath: basePath}
}

func (s *LocalFileStorage) SaveFiles(dirPath string, files []storage.FileInput) *pkgerrors.Error {
	fullPath := filepath.Join(s.basePath, dirPath)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return apperrors.ReturnInternalError("failed to create template directory")
	}

	for _, f := range files {
		filePath := filepath.Join(fullPath, f.Name)

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return apperrors.ReturnInternalError("failed to create directory for file: " + f.Name)
		}

		dst, err := os.Create(filePath)
		if err != nil {
			os.RemoveAll(fullPath)
			return apperrors.ReturnInternalError("failed to create file: " + f.Name)
		}

		if _, err := io.Copy(dst, f.Reader); err != nil {
			dst.Close()
			os.RemoveAll(fullPath)
			return apperrors.ReturnInternalError("failed to write file: " + f.Name)
		}

		dst.Close()
	}

	return nil
}

func (s *LocalFileStorage) DeleteDir(dirPath string) *pkgerrors.Error {
	fullPath := filepath.Join(s.basePath, dirPath)

	if err := os.RemoveAll(fullPath); err != nil {
		return apperrors.ReturnInternalError("failed to delete template directory")
	}

	return nil
}

func (s *LocalFileStorage) ListFiles(dirPath string) ([]storage.FileInfo, *pkgerrors.Error) {
	fullPath := filepath.Join(s.basePath, dirPath)

	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, apperrors.ReturnInternalError("failed to list template files")
	}

	var files []storage.FileInfo
	err := filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		relPath, relErr := filepath.Rel(fullPath, path)
		if relErr != nil {
			return nil
		}
		files = append(files, storage.FileInfo{
			Name: filepath.ToSlash(relPath),
			Size: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, apperrors.ReturnInternalError("failed to list template files")
	}

	return files, nil
}

func (s *LocalFileStorage) ReadFile(filePath string) ([]byte, *pkgerrors.Error) {
	fullPath := filepath.Join(s.basePath, filePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apperrors.ReturnNotFound("file not found")
		}
		return nil, apperrors.ReturnInternalError("failed to read file")
	}

	return data, nil
}
