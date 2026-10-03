package testingx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestProjectRoot_Same_Dir(t *testing.T) {
	dir := t.TempDir()

	fileDir := filepath.Join(dir, "go.mod")
	file, err := os.Create(fileDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.Close())

	actualDir, err := ProjectRoot(dir)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, dir, actualDir)
}

func TestProjectRoot_Up(t *testing.T) {
	dir := t.TempDir()

	start := filepath.Join(dir, "dir1/dir2/dir3")
	if err := os.MkdirAll(start, 0700); err != nil {
		t.Fatal(err)
	}

	file, err := os.Create(filepath.Join(dir, "dir1/go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.Close())

	actualDir, err := ProjectRoot(start)
	if err != nil {
		t.Error(err)
	}

	expectedDir := filepath.Join(dir, "dir1")
	assert.Equal(t, expectedDir, actualDir)
}

func TestProjectRoot_Not_Found(t *testing.T) {
	dir := t.TempDir()

	start := filepath.Join(dir, "dir1/dir2/dir3")
	if err := os.MkdirAll(start, 0700); err != nil {
		t.Fatal(err)
	}

	actualDir, err := ProjectRoot(start)
	if err == nil {
		t.Fatal("expected ProjectRoot to error but didn't")
	}
	assert.Equal(t, ErrFailedToFindProjectRoot, err)
	assert.Equal(t, "", actualDir)
}
