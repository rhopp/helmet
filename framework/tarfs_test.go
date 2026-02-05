package framework

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"testing"

	o "github.com/onsi/gomega"
)

// createTarballFromMap creates a simple tarball from a map of files
func createTarballFromMap(files map[string]string) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			panic(err)
		}
	}

	if err := tw.Close(); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// TestNewTarFS tests creating an fs.FS from a tarball
func TestNewTarFS(t *testing.T) {
	g := o.NewWithT(t)

	tarball := createTarballFromMap(map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})

	tfs, err := NewTarFS(tarball)
	g.Expect(err).To(o.Succeed())
	g.Expect(tfs).ToNot(o.BeNil())

	// Verify we can read files from the tarfs
	file1, err := fs.ReadFile(tfs, "file1.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(string(file1)).To(o.Equal("content1"))

	file2, err := fs.ReadFile(tfs, "file2.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(string(file2)).To(o.Equal("content2"))
}

// TestNewTarFSWithDirectory tests tarball containing directories
func TestNewTarFSWithDirectory(t *testing.T) {
	g := o.NewWithT(t)

	tarball := createTarballFromMap(map[string]string{
		"dir/file.txt":        "nested content",
		"dir/subdir/deep.txt": "deeply nested",
	})

	tfs, err := NewTarFS(tarball)
	g.Expect(err).To(o.Succeed())
	g.Expect(tfs).ToNot(o.BeNil())

	// Verify nested files can be read
	content, err := fs.ReadFile(tfs, "dir/file.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(string(content)).To(o.Equal("nested content"))

	deepContent, err := fs.ReadFile(tfs, "dir/subdir/deep.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(string(deepContent)).To(o.Equal("deeply nested"))
}

// TestNewTarFSEmpty tests creating fs.FS from empty tarball
func TestNewTarFSEmpty(t *testing.T) {
	g := o.NewWithT(t)

	tarball := createTarballFromMap(map[string]string{})

	tfs, err := NewTarFS(tarball)
	g.Expect(err).To(o.Succeed())
	g.Expect(tfs).ToNot(o.BeNil())

	// Should be a valid filesystem, just empty
	_, err = fs.ReadFile(tfs, "nonexistent.txt")
	g.Expect(err).To(o.HaveOccurred()) // File doesn't exist, which is expected
}

// TestNewTarFSInvalidTarball tests error handling with invalid tarball
func TestNewTarFSInvalidTarball(t *testing.T) {
	g := o.NewWithT(t)

	invalidTarball := []byte("this is not a valid tarball")

	tfs, err := NewTarFS(invalidTarball)
	g.Expect(err).To(o.HaveOccurred())
	g.Expect(tfs).To(o.BeNil())
}

// TestNewTarFSNilTarball tests error handling with nil tarball
func TestNewTarFSNilTarball(t *testing.T) {
	g := o.NewWithT(t)

	tfs, err := NewTarFS(nil)
	// tarfs.New with empty reader doesn't error - it creates an empty filesystem
	g.Expect(err).To(o.Succeed())
	g.Expect(tfs).ToNot(o.BeNil())
}

// TestNewTarFSWalk tests walking the tarfs filesystem
func TestNewTarFSWalk(t *testing.T) {
	g := o.NewWithT(t)

	tarball := createTarballFromMap(map[string]string{
		"file1.txt":     "content1",
		"dir/file2.txt": "content2",
		"dir/file3.txt": "content3",
	})

	tfs, err := NewTarFS(tarball)
	g.Expect(err).To(o.Succeed())

	// Walk the filesystem and count files
	fileCount := 0
	err = fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileCount++
		}
		return nil
	})

	g.Expect(err).To(o.Succeed())
	g.Expect(fileCount).To(o.Equal(3))
}

// TestNewTarFSOpen tests opening files from tarfs
func TestNewTarFSOpen(t *testing.T) {
	g := o.NewWithT(t)

	tarball := createTarballFromMap(map[string]string{
		"test.txt": "test content",
	})

	tfs, err := NewTarFS(tarball)
	g.Expect(err).To(o.Succeed())

	file, err := tfs.Open("test.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(file).ToNot(o.BeNil())
	defer file.Close()

	content, err := io.ReadAll(file)
	g.Expect(err).To(o.Succeed())
	g.Expect(string(content)).To(o.Equal("test content"))
}
