package tinygo

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const DefaultVersion = "0.41.1"

// defaultInstallDir returns the platform-specific install directory,
// matching the official TinyGo installation instructions at
// https://tinygo.org/getting-started/install/
//
//   - Linux/macOS: /usr/local  (tar -xf tinygo...tar.gz -C /usr/local)
//   - Windows:     "" (Scoop manages its own directories; no fixed path)
func defaultInstallDir(goos string) string {
	if goos == "windows" {
		return ""
	}
	// Fall back to user-local when /usr/local is not writable (e.g. CI non-root runners).
	if canWriteToDir("/usr/local") {
		return "/usr/local"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local"
	}
	return filepath.Join(home, ".local")
}

func canWriteToDir(dir string) bool {
	tmp, err := os.MkdirTemp(dir, ".tinygo-write-check-*")
	if err != nil {
		return false
	}
	os.Remove(tmp)
	return true
}

type config struct {
	version    string
	installDir string
	logger     func(string)
	lookPath   func(string) (string, error)
	httpClient *http.Client
	goos       string
	goarch     string
	// for testing
	downloadURLFunc func() string
}

type Option func(*config)

func WithVersion(v string) Option {
	return func(c *config) {
		c.version = v
	}
}

func WithInstallDir(dir string) Option {
	return func(c *config) {
		c.installDir = dir
	}
}

func WithLogger(f func(string)) Option {
	return func(c *config) {
		c.logger = f
	}
}

func newConfig(opts ...Option) *config {
	c := &config{
		version:    DefaultVersion,
		installDir: defaultInstallDir(runtime.GOOS),
		logger:     func(string) {},
		lookPath:   exec.LookPath,
		httpClient: http.DefaultClient,
		goos:       runtime.GOOS,
		goarch:     runtime.GOARCH,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// internal options for testing (not exported)
func withLookPath(f func(string) (string, error)) Option {
	return func(c *config) {
		c.lookPath = f
	}
}

func withHTTPClient(cl *http.Client) Option {
	return func(c *config) {
		c.httpClient = cl
	}
}

func withGOOS(goos string) Option {
	return func(c *config) {
		c.goos = goos
	}
}

func withGOARCH(goarch string) Option {
	return func(c *config) {
		c.goarch = goarch
	}
}

func withDownloadURLFunc(f func() string) Option {
	return func(c *config) {
		c.downloadURLFunc = f
	}
}

func (c *config) binPath() string {
	bin := filepath.Join(c.installDir, "tinygo", "bin", "tinygo")
	if c.goos == "windows" {
		bin += ".exe"
	}
	return bin
}

func (c *config) downloadURL() string {
	if c.downloadURLFunc != nil {
		return c.downloadURLFunc()
	}
	ext := "tar.gz"
	if c.goos == "windows" {
		ext = "zip"
	}
	// eg: tinygo0.40.1.linux-amd64.tar.gz
	return "https://github.com/tinygo-org/tinygo/releases/download/v" +
		c.version + "/tinygo" + c.version + "." + c.goos + "-" + c.goarch + "." + ext
}
