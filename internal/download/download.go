package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// DownloadTo downloads a URL to a local file path with progress output to stderr.
func DownloadTo(url, dest, label string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	var totalBytes int64
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		totalBytes, _ = strconv.ParseInt(cl, 10, 64)
	}

	buf := make([]byte, 256*1024)
	var downloaded int64
	lastPct := -1

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("writing file: %w", writeErr)
			}
			downloaded += int64(n)

			if totalBytes > 0 {
				pct := int(downloaded * 100 / totalBytes)
				if pct != lastPct && (pct%2 == 0 || pct == 100) {
					fmt.Fprintf(os.Stderr, "\r%s: %3d%% (%s / %s)", label, pct, formatBytes(downloaded), formatBytes(totalBytes))
					lastPct = pct
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("reading response: %w", readErr)
		}
	}
	fmt.Fprintln(os.Stderr)
	return nil
}

func formatBytes(n int64) string {
	v := float64(n)
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if v < 1024 || unit == "GB" {
			if unit == "B" {
				return fmt.Sprintf("%d%s", int(v), unit)
			}
			return fmt.Sprintf("%.1f%s", v, unit)
		}
		v /= 1024
	}
	return fmt.Sprintf("%dB", int(v))
}
