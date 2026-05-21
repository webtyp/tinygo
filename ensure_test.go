package tinygo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureInstalled(t *testing.T) {
	// Setup mock archive
	archiveFile, _ := os.CreateTemp("", "tinygo-*.tar.gz")
	defer os.Remove(archiveFile.Name())
	gw := gzip.NewWriter(archiveFile)
	tw := tar.NewWriter(gw)
	mockBody := "#!/bin/bash\necho \"tinygo version " + DefaultVersion + " linux/amd64\"\n"
	tw.WriteHeader(&tar.Header{Name: "tinygo/bin/tinygo", Mode: 0755, Size: int64(len(mockBody))})
	tw.Write([]byte(mockBody))
	tw.Close()
	gw.Close()
	archiveFile.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archiveFile.Name())
	}))
	defer ts.Close()

	tmpDir, _ := os.MkdirTemp("", "tinygo-ensure-*")
	defer os.RemoveAll(tmpDir)

	lookPathFails := func(string) (string, error) { return "", fmt.Errorf("not found") }

	// Test 1: Install when missing
	binPath, err := EnsureInstalled(
		WithInstallDir(tmpDir),
		withHTTPClient(ts.Client()),
		withDownloadURLFunc(func() string { return ts.URL }),
		withLookPath(lookPathFails),
		withGOOS("linux"),
	)
	if err != nil {
		t.Fatalf("EnsureInstalled failed: %v", err)
	}

	expectedBinPath := filepath.Join(tmpDir, "tinygo/bin/tinygo")
	if binPath != expectedBinPath {
		t.Errorf("expected %s, got %s", expectedBinPath, binPath)
	}

	// Test 2: Return existing when present
	binPath2, err := EnsureInstalled(
		WithInstallDir(tmpDir),
		withLookPath(lookPathFails),
		withGOOS("linux"),
	)
	if err != nil {
		t.Fatalf("Second EnsureInstalled failed: %v", err)
	}
	if binPath2 != binPath {
		t.Errorf("expected same path, got %s", binPath2)
	}
}
