package terraform

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Result holds the output from a Terraform command execution.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Executor runs Terraform CLI commands in execution directories.
type Executor struct {
	basePath        string
	pluginCacheDir  string
}

func NewExecutor(executionBasePath, pluginCacheDir string) *Executor {
	return &Executor{
		basePath:       executionBasePath,
		pluginCacheDir: pluginCacheDir,
	}
}

func (e *Executor) Init(ctx context.Context, executionPath string) (*Result, error) {
	return e.run(ctx, executionPath, "init", "-input=false", "-no-color")
}

func (e *Executor) Plan(ctx context.Context, executionPath string) (*Result, error) {
	return e.run(ctx, executionPath, "plan", "-input=false", "-no-color")
}

func (e *Executor) Apply(ctx context.Context, executionPath string) (*Result, error) {
	return e.run(ctx, executionPath, "apply", "-auto-approve", "-input=false", "-no-color")
}

func (e *Executor) Destroy(ctx context.Context, executionPath string) (*Result, error) {
	return e.run(ctx, executionPath, "destroy", "-auto-approve", "-input=false", "-no-color")
}

func (e *Executor) run(ctx context.Context, executionPath string, args ...string) (*Result, error) {
	workDir := filepath.Join(e.basePath, executionPath)

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = workDir

	if e.pluginCacheDir != "" {
		cmd.Env = append(cmd.Environ(), "TF_PLUGIN_CACHE_DIR="+e.pluginCacheDir)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("terraform %s failed (exit %d): %s", args[0], result.ExitCode, stderr.String())
	} else if err != nil {
		return result, fmt.Errorf("terraform %s exec error: %w", args[0], err)
	}

	return result, nil
}
