package storage

import (
	"io"

	pkgerrors "backend/pkg/errors"
)

type FileInput struct {
	Name   string    `json:"name" validate:"required,filepath"`
	Reader io.Reader `json:"-"`
	Size   int64     `json:"size" validate:"required,gt=0,max=1048576"`
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

// ExecutionStorage manages Terraform execution directories where environments
// run their terraform init/apply/destroy lifecycle.
type ExecutionStorage interface {
	// CopyTemplateToExecution copies all files from a template directory into
	// a new execution directory for an environment.
	CopyTemplateToExecution(templatePath, executionPath string) *pkgerrors.Error

	// WriteVarsFile writes a terraform.tfvars file into the execution directory.
	WriteVarsFile(executionPath string, content []byte) *pkgerrors.Error

	// DeleteDir removes an execution directory and all its contents.
	DeleteDir(executionPath string) *pkgerrors.Error

	// Exists checks whether the execution directory exists.
	Exists(executionPath string) bool
}
