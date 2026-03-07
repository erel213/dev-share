package filestorage

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"backend/internal/domain/storage"
)

func TestSaveFiles_NestedPaths(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	files := []storage.FileInput{
		{Name: "modules/vpc/main.tf", Reader: strings.NewReader("vpc content"), Size: 11},
		{Name: "modules/ec2/main.tf", Reader: strings.NewReader("ec2 content"), Size: 11},
		{Name: "main.tf", Reader: strings.NewReader("root content"), Size: 12},
	}

	if err := s.SaveFiles("tmpl1", files); err != nil {
		t.Fatalf("SaveFiles: %v", err)
	}

	// Verify files exist on disk
	for _, f := range files {
		path := filepath.Join(base, "tmpl1", f.Name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", f.Name, err)
		}
	}
}

func TestSaveFiles_FlatFiles(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	files := []storage.FileInput{
		{Name: "main.tf", Reader: strings.NewReader("hello"), Size: 5},
		{Name: "vars.tf", Reader: strings.NewReader("world"), Size: 5},
	}

	if err := s.SaveFiles("tmpl2", files); err != nil {
		t.Fatalf("SaveFiles: %v", err)
	}

	for _, f := range files {
		path := filepath.Join(base, "tmpl2", f.Name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", f.Name, err)
		}
	}
}

func TestListFiles_Recursive(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	files := []storage.FileInput{
		{Name: "main.tf", Reader: strings.NewReader("root"), Size: 4},
		{Name: "modules/vpc/main.tf", Reader: strings.NewReader("vpc"), Size: 3},
		{Name: "modules/ec2/main.tf", Reader: strings.NewReader("ec2"), Size: 3},
	}

	if err := s.SaveFiles("tmpl3", files); err != nil {
		t.Fatalf("SaveFiles: %v", err)
	}

	listed, listErr := s.ListFiles("tmpl3")
	if listErr != nil {
		t.Fatalf("ListFiles: %v", listErr)
	}

	names := make([]string, len(listed))
	for i, f := range listed {
		names[i] = f.Name
	}
	sort.Strings(names)

	expected := []string{"main.tf", "modules/ec2/main.tf", "modules/vpc/main.tf"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(names), names)
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], name)
		}
	}
}

func TestListFiles_NonExistentDir(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	files, err := s.ListFiles("nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files, got %v", files)
	}
}

func TestReadFile_NestedPath(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	files := []storage.FileInput{
		{Name: "modules/vpc/main.tf", Reader: strings.NewReader("vpc content"), Size: 11},
	}

	if err := s.SaveFiles("tmpl4", files); err != nil {
		t.Fatalf("SaveFiles: %v", err)
	}

	data, readErr := s.ReadFile("tmpl4/modules/vpc/main.tf")
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(data) != "vpc content" {
		t.Errorf("expected 'vpc content', got '%s'", string(data))
	}
}

func TestReadFile_NotFound(t *testing.T) {
	base := t.TempDir()
	s := NewLocalFileStorage(base)

	_, err := s.ReadFile("tmpl4/nonexistent.tf")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}
