package state

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/driftctl/driftctl/internal/model"
)

// Reader loads raw Terraform state bytes from various backends.
type Reader interface {
	Read(ctx context.Context, cfg model.StateConfig) ([]byte, error)
}



// allowedStateRoot is the directory under which all local state files must
// live. Prevents path traversal (e.g. "../../etc/passwd") from an
// API-supplied workspace path escaping into arbitrary filesystem reads.
var allowedStateRoot = getAllowedStateRoot()

func getAllowedStateRoot() string {
        return os.Getenv("DRIFTCTL_STATE_ROOT")
}



func safeStatePath(root, requested string) (string, error) {
        if root == "" {
                return requested, nil
        }
        var cleanFull string
        if filepath.IsAbs(requested) {
                cleanFull = filepath.Clean(requested)
        } else {
                cleanFull = filepath.Clean(filepath.Join(root, requested))
        }
        cleanRoot := filepath.Clean(root)
        if cleanFull != cleanRoot && !strings.HasPrefix(cleanFull, cleanRoot+string(filepath.Separator)) {
                return "", fmt.Errorf("invalid state path: %q escapes allowed root", requested)
        }
        return cleanFull, nil
}







// DefaultReader supports local files, HTTP(S), and S3 backends.
type DefaultReader struct{}

func NewReader() *DefaultReader {
	return &DefaultReader{}
}

func (r *DefaultReader) Read(ctx context.Context, cfg model.StateConfig) ([]byte, error) {
	backend := strings.ToLower(strings.TrimSpace(cfg.Backend))
	if backend == "" {
		backend = "local"
	}

	switch backend {
	case "local", "file":
		return readLocal(cfg.Path)
	case "http", "https":
		return readHTTP(ctx, cfg.Path)
	case "s3":
		return readS3(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported state backend: %s", cfg.Backend)
	}
}



func readLocal(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("state path is required for local backend")
	}
	safePath, err := safeStatePath(allowedStateRoot, path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("read state file %s: %w", safePath, err)
	}
	return data, nil
}





func readHTTP(ctx context.Context, url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("state URL is required for http backend")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch state from %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch state from %s: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
