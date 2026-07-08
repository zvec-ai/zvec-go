// download-libs downloads the pre-built zvec C-API libraries for the current platform
// from the upstream zvec-ai/zvec-go GitHub Releases.
//
// Usage:
//
//	go run ./cmd/download-libs [-version v0.5.1] [-dest ./lib]
//
// If -version is not specified, it queries the GitHub Releases API to fetch the
// latest published release tag.
// If -dest is not specified, it defaults to ./lib relative to the module root.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	upstreamRepo = "zvec-ai/zvec-go"
	baseURL      = "https://github.com/" + upstreamRepo + "/releases/download"
	latestAPIURL = "https://api.github.com/repos/" + upstreamRepo + "/releases/latest"
)

// platformArtifact maps GOOS/GOARCH to the artifact name and whether it's a zip.
type platformArtifact struct {
	name  string
	isZip bool
}

var platformMap = map[string]platformArtifact{
	"darwin/arm64":  {name: "zvec-libs-darwin-arm64.tar.gz", isZip: false},
	"linux/amd64":   {name: "zvec-libs-linux-x64.tar.gz", isZip: false},
	"linux/arm64":   {name: "zvec-libs-linux-arm64.tar.gz", isZip: false},
	"windows/amd64": {name: "zvec-libs-windows-x64.zip", isZip: true},
}

func main() {
	var version string
	var dest string

	flag.StringVar(&version, "version", "", "Library version to download (e.g. v0.5.1). Defaults to the latest GitHub release.")
	flag.StringVar(&dest, "dest", "", "Destination directory for lib/. Defaults to ./lib relative to module root.")
	flag.Parse()

	// Locate module root (directory containing this go.mod)
	moduleRoot, err := findModuleRoot()
	if err != nil {
		fatalf("Cannot find module root: %v", err)
	}

	// Resolve version: when -version is not specified, query GitHub Releases API
	// for the latest published tag. This avoids the need for an in-tree VERSION file
	// that has to be bumped on every release.
	if version == "" {
		fmt.Println("No -version flag provided; querying latest release from GitHub...")
		version, err = fetchLatestReleaseTag()
		if err != nil {
			fatalf("Cannot determine latest release: %v\nUsage: go run ./cmd/download-libs -version v0.5.1", err)
		}
		fmt.Printf("  Latest release: %s\n", version)
	}
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Resolve destination
	if dest == "" {
		dest = filepath.Join(moduleRoot, "lib")
	}

	// Detect platform
	key := runtime.GOOS + "/" + runtime.GOARCH
	artifact, ok := platformMap[key]
	if !ok {
		fatalf("Unsupported platform: %s\nSupported platforms: darwin/arm64, linux/amd64, linux/arm64, windows/amd64", key)
	}

	downloadURL := fmt.Sprintf("%s/%s/%s", baseURL, version, artifact.name)

	fmt.Printf("Downloading pre-built libraries for %s (%s)...\n", key, version)
	fmt.Printf("  URL: %s\n", downloadURL)
	fmt.Printf("  Destination: %s\n", dest)

	// Download to temp file
	tmpFile, err := os.CreateTemp("", "zvec-libs-*")
	if err != nil {
		fatalf("Cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := download(downloadURL, tmpFile); err != nil {
		fatalf("Download failed: %v", err)
	}
	tmpFile.Close()

	// Extract
	fmt.Println("Extracting libraries...")
	if artifact.isZip {
		err = extractZip(tmpFile.Name(), dest)
	} else {
		err = extractTarGz(tmpFile.Name(), dest)
	}
	if err != nil {
		fatalf("Extraction failed: %v", err)
	}

	fmt.Println("Done! Pre-built libraries installed to:", dest)
	fmt.Println("You can now build with: CGO_ENABLED=1 go build .")
}

// findModuleRoot walks up from the current directory to find go.mod.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

// fetchLatestReleaseTag queries the GitHub Releases API and returns the
// tag name of the latest published release (e.g. "v0.5.1").
func fetchLatestReleaseTag() (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, latestAPIURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	// Honor GITHUB_TOKEN if available to lift the unauthenticated rate limit.
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, latestAPIURL, strings.TrimSpace(string(body)))
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response from %s", latestAPIURL)
	}
	return payload.TagName, nil
}

// download fetches a URL and writes to dst.
func download(url string, dst *os.File) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, writeErr := dst.Write(buf[:n])
			written += int64(wn)
			if writeErr != nil {
				return writeErr
			}
			if total > 0 {
				fmt.Printf("\r  Progress: %.1f MB / %.1f MB", float64(written)/1e6, float64(total)/1e6)
			} else {
				fmt.Printf("\r  Downloaded: %.1f MB", float64(written)/1e6)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	fmt.Println()
	return nil
}

// extractTarGz extracts a .tar.gz archive into destDir.
func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Security: prevent path traversal
		// Strip leading "./" if present
		cleanName := strings.TrimPrefix(hdr.Name, "./")
		cleaned := filepath.Clean("/" + cleanName)
		if cleaned == "/" {
			continue // Skip root directory entry
		}
		target := filepath.Join(destDir, cleaned[1:])
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec
				out.Close()
				return err
			}
			out.Close()
			fmt.Println("  +", strings.TrimPrefix(target, destDir+string(os.PathSeparator)))
		}
	}
	return nil
}

// extractZip extracts a .zip archive into destDir.
func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Strip leading "./" if present
		cleanName := strings.TrimPrefix(f.Name, "./")
		cleaned := filepath.Clean("/" + cleanName)
		if cleaned == "/" {
			continue // Skip root directory entry
		}
		target := filepath.Join(destDir, cleaned[1:])
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil { //nolint:gosec
			rc.Close()
			out.Close()
			return err
		}
		rc.Close()
		out.Close()
		fmt.Println("  +", strings.TrimPrefix(target, destDir+string(os.PathSeparator)))
	}
	return nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
