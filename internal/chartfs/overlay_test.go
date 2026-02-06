package chartfs

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	o "github.com/onsi/gomega"
)

// TestNewOverlayFS tests the NewOverlayFS constructor.
func TestNewOverlayFS(t *testing.T) {
	g := o.NewWithT(t)

	// Create test filesystems
	embedded := fstest.MapFS{
		"embedded-file.txt": {Data: []byte("embedded content")},
		"shared-file.txt":   {Data: []byte("embedded version")},
	}

	local := fstest.MapFS{
		"local-file.txt":  {Data: []byte("local content")},
		"shared-file.txt": {Data: []byte("local version")},
	}

	overlay := NewOverlayFS(embedded, local)
	g.Expect(overlay).ToNot(o.BeNil())
	g.Expect(overlay.Embedded).To(o.Equal(fs.FS(embedded)))
	g.Expect(overlay.Local).To(o.Equal(fs.FS(local)))
}

// TestOverlayFSOpen tests the Open method.
func TestOverlayFSOpen(t *testing.T) {
	// Create test filesystems
	embedded := fstest.MapFS{
		"embedded-file.txt": {Data: []byte("embedded content")},
		"shared-file.txt":   {Data: []byte("embedded version")},
	}

	local := fstest.MapFS{
		"local-file.txt":  {Data: []byte("local content")},
		"shared-file.txt": {Data: []byte("local version")},
	}

	overlay := NewOverlayFS(embedded, local)

	t.Run("open_from_embedded", func(t *testing.T) {
		g := o.NewWithT(t)

		file, err := overlay.Open("embedded-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(file).ToNot(o.BeNil())
		defer file.Close()

		// Read content to verify it's from embedded
		content, err := fs.ReadFile(overlay, "embedded-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(string(content)).To(o.Equal("embedded content"))
	})

	t.Run("open_from_local", func(t *testing.T) {
		g := o.NewWithT(t)

		file, err := overlay.Open("local-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(file).ToNot(o.BeNil())
		defer file.Close()

		// Read content to verify it's from local
		content, err := fs.ReadFile(overlay, "local-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(string(content)).To(o.Equal("local content"))
	})

	t.Run("prefer_embedded_over_local", func(t *testing.T) {
		g := o.NewWithT(t)

		// shared-file.txt exists in both, should return embedded version
		file, err := overlay.Open("shared-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(file).ToNot(o.BeNil())
		defer file.Close()

		// Read content to verify it's from embedded (not local)
		content, err := fs.ReadFile(overlay, "shared-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(string(content)).To(o.Equal("embedded version"))
	})

	t.Run("file_not_found", func(t *testing.T) {
		g := o.NewWithT(t)

		file, err := overlay.Open("nonexistent-file.txt")
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(file).To(o.BeNil())
	})
}

// TestWithEmbeddedBaseDir tests the WithEmbeddedBaseDir method.
func TestWithEmbeddedBaseDir(t *testing.T) {
	// Create test filesystems with subdirectories
	embedded := fstest.MapFS{
		"subdir/file.txt": {Data: []byte("embedded subdir content")},
		"root-file.txt":   {Data: []byte("root content")},
	}

	local := fstest.MapFS{
		"local-file.txt": {Data: []byte("local content")},
	}

	overlay := NewOverlayFS(embedded, local)

	t.Run("change_embedded_base_dir", func(t *testing.T) {
		g := o.NewWithT(t)

		subOverlay, err := overlay.WithEmbeddedBaseDir("subdir")
		g.Expect(err).To(o.Succeed())
		g.Expect(subOverlay).ToNot(o.BeNil())

		// Should be able to access file.txt directly (was subdir/file.txt)
		file, err := subOverlay.Open("file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(file).ToNot(o.BeNil())
		defer file.Close()

		content, err := fs.ReadFile(subOverlay, "file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(string(content)).To(o.Equal("embedded subdir content"))

		// Local filesystem should remain unchanged
		localFile, err := subOverlay.Open("local-file.txt")
		g.Expect(err).To(o.Succeed())
		g.Expect(localFile).ToNot(o.BeNil())
		defer localFile.Close()
	})

	t.Run("embedded_base_dir_changes_scope", func(t *testing.T) {
		g := o.NewWithT(t)

		// After changing to subdir, root-file.txt should not be accessible
		subOverlay, err := overlay.WithEmbeddedBaseDir("subdir")
		g.Expect(err).To(o.Succeed())

		// root-file.txt should not be accessible from subdir scope
		_, err = subOverlay.Open("root-file.txt")
		g.Expect(err).To(o.HaveOccurred())

		// But file.txt (was subdir/file.txt) should be accessible
		_, err = subOverlay.Open("file.txt")
		g.Expect(err).To(o.Succeed())
	})
}

// TestOverlayFSWithRealFS tests OverlayFS with real os.DirFS.
func TestOverlayFSWithRealFS(t *testing.T) {
	g := o.NewWithT(t)

	// Create embedded test filesystem
	embedded := fstest.MapFS{
		"test-file.txt": {Data: []byte("embedded test content")},
	}

	// Use real test directory as local filesystem
	local := os.DirFS("../../test")

	overlay := NewOverlayFS(embedded, local)
	g.Expect(overlay).ToNot(o.BeNil())

	// Should be able to open embedded file
	embeddedFile, err := overlay.Open("test-file.txt")
	g.Expect(err).To(o.Succeed())
	g.Expect(embeddedFile).ToNot(o.BeNil())
	defer embeddedFile.Close()

	// Should be able to open local file
	localFile, err := overlay.Open("config.yaml")
	g.Expect(err).To(o.Succeed())
	g.Expect(localFile).ToNot(o.BeNil())
	defer localFile.Close()
}
