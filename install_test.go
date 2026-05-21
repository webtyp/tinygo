package tinygo

import (
	"archive/tar"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInstall(t *testing.T) {
	// 1. Create a dummy tar.gz archive
	archiveFile, err := os.CreateTemp("", "tinygo-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(archiveFile.Name())
	defer archiveFile.Close()

	gw := gzip.NewWriter(archiveFile)
	tw := tar.NewWriter(gw)
	mockBody := "#!/bin/bash\necho \"tinygo version " + DefaultVersion + " linux/amd64\"\n"
	tw.WriteHeader(&tar.Header{
		Name: "tinygo/bin/tinygo",
		Mode: 0755,
		Size: int64(len(mockBody)),
	})
	tw.Write([]byte(mockBody))
	tw.Close()
	gw.Close()

	// 2. Mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archiveFile.Name())
	}))
	defer ts.Close()

	// 3. Configure and run Install
	tmpInstallDir, err := os.MkdirTemp("", "tinygo-install-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpInstallDir)

	err = Install(
		WithInstallDir(tmpInstallDir),
		withHTTPClient(ts.Client()),
		withDownloadURLFunc(func() string { return ts.URL }),
		withGOOS("linux"),
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// 4. Verify binary exists
	binPath := filepath.Join(tmpInstallDir, "tinygo/bin/tinygo")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("binary not found at %s", binPath)
	}

	// 5. Test idempotency
	var logged []string
	logger := func(s string) {
		logged = append(logged, s)
	}
	err = Install(
		WithInstallDir(tmpInstallDir),
		WithLogger(logger),
		withGOOS("linux"),
	)
	if err != nil {
		t.Fatalf("Second Install failed: %v", err)
	}
	found := false
	for _, l := range logged {
		if l == "TinyGo already installed at "+binPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected idempotency log message")
	}
}
