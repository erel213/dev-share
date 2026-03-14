package filestorage

import (
	"io"
	"os"
	"path/filepath"

	apperrors "backend/internal/application/errors"
	"backend/internal/domain/storage"
	pkgerrors "backend/pkg/errors"
)

type LocalExecutionStorage struct {
	executionBasePath string
	templateBasePath  string
}

func NewLocalExecutionStorage(executionBasePath, templateBasePath string) storage.ExecutionStorage {
	return &LocalExecutionStorage{
		executionBasePath: executionBasePath,
		templateBasePath:  templateBasePath,
	}
}

func (s *LocalExecutionStorage) CopyTemplateToExecution(templatePath, executionPath string) *pkgerrors.Error {
	srcDir := filepath.Join(s.templateBasePath, templatePath)
	dstDir := filepath.Join(s.executionBasePath, executionPath)

	if err := os.MkdirAll(dstDir, 0700); err != nil {
		return apperrors.ReturnInternalError("failed to create execution directory")
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		os.RemoveAll(dstDir)
		return apperrors.ReturnInternalError("failed to read template directory: " + templatePath)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		src, err := os.Open(filepath.Join(srcDir, entry.Name()))
		if err != nil {
			os.RemoveAll(dstDir)
			return apperrors.ReturnInternalError("failed to open template file: " + entry.Name())
		}

		dst, err := os.OpenFile(filepath.Join(dstDir, entry.Name()), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			src.Close()
			os.RemoveAll(dstDir)
			return apperrors.ReturnInternalError("failed to create execution file: " + entry.Name())
		}

		_, copyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if copyErr != nil {
			os.RemoveAll(dstDir)
			return apperrors.ReturnInternalError("failed to copy file: " + entry.Name())
		}
	}

	return nil
}

func (s *LocalExecutionStorage) WriteVarsFile(executionPath string, content []byte) *pkgerrors.Error {
	fullPath := filepath.Join(s.executionBasePath, executionPath, "terraform.tfvars")

	if err := os.WriteFile(fullPath, content, 0600); err != nil {
		return apperrors.ReturnInternalError("failed to write terraform.tfvars")
	}

	return nil
}

func (s *LocalExecutionStorage) DeleteDir(executionPath string) *pkgerrors.Error {
	fullPath := filepath.Join(s.executionBasePath, executionPath)

	if err := os.RemoveAll(fullPath); err != nil {
		return apperrors.ReturnInternalError("failed to delete execution directory")
	}

	return nil
}

func (s *LocalExecutionStorage) Exists(executionPath string) bool {
	fullPath := filepath.Join(s.executionBasePath, executionPath)
	_, err := os.Stat(fullPath)
	return err == nil
}
