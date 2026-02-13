package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// SelfUpdate downloads the latest release and replaces the current binary.
func SelfUpdate(currentVersion string) error {
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update a development build")
	}

	fmt.Println("Checking for updates...")

	release, err := fetchLatestRelease(10 * time.Second)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	latest := release.TagName
	if !isNewer(latest, currentVersion) {
		fmt.Printf("Already up to date (%s).\n", currentVersion)
		return nil
	}

	archiveExt := "tar.gz"
	if runtime.GOOS == "windows" {
		archiveExt = "zip"
	}

	version := strings.TrimPrefix(latest, "v")
	assetName := fmt.Sprintf("switcher_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, archiveExt)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (%s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	fmt.Printf("Downloading %s...\n", latest)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	archiveData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading download: %w", err)
	}

	var binary []byte
	if archiveExt == "zip" {
		binary, err = extractFromZip(archiveData, "switcher")
	} else {
		binary, err = extractFromTarGz(archiveData, "switcher")
	}
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	if err := replaceBinary(binary); err != nil {
		return err
	}

	WriteCache(latest)

	fmt.Printf("Successfully updated to %s (was %s).\n", latest, currentVersion)
	return nil
}

func extractFromTarGz(data []byte, binaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Match the binary by base name (archive may have directory prefixes)
		name := hdr.Name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		if name == binaryName && hdr.Typeflag == tar.TypeReg {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("%q not found in archive", binaryName)
}

func extractFromZip(data []byte, binaryName string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	// On Windows the binary has .exe extension
	targets := []string{binaryName, binaryName + ".exe"}

	for _, f := range r.File {
		name := f.Name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		for _, target := range targets {
			if name == target {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
	}
	return nil, fmt.Errorf("%q not found in archive", binaryName)
}
