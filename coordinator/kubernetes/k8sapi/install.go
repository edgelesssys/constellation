package k8sapi

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/spf13/afero"
	"golang.org/x/text/transform"
)

// osInstaller installs binary components of supported kubernetes versions.
type osInstaller struct {
	fs      *afero.Afero
	hClient httpClient
}

// newOSInstaller creates a new osInstaller.
func newOSInstaller() *osInstaller {
	return &osInstaller{
		fs:      &afero.Afero{Fs: afero.NewOsFs()},
		hClient: &http.Client{},
	}
}

// Install downloads a resource from a URL, applies any given text transformations and extracts the resulting file if required.
// The resulting file(s) are copied to all destinations.
func (i *osInstaller) Install(
	ctx context.Context, sourceURL string, destinations []string, perm fs.FileMode,
	extract bool, transforms ...transform.Transformer,
) error {
	tempPath, err := i.downloadToTempDir(ctx, sourceURL, transforms...)
	if err != nil {
		return err
	}
	defer func() {
		_ = i.fs.Remove(tempPath)
	}()
	for _, destination := range destinations {
		var err error
		if extract {
			err = i.extractArchive(tempPath, destination, perm)
		} else {
			err = i.copy(tempPath, destination, perm)
		}
		if err != nil {
			return fmt.Errorf("installing from %q: copying to destination %q: %w", sourceURL, destination, err)
		}
	}
	return nil
}

// extractArchive extracts tar gz archives to a prefixed destination.
func (i *osInstaller) extractArchive(archivePath, prefix string, perm fs.FileMode) error {
	archiveFile, err := i.fs.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive file: %w", err)
	}
	defer archiveFile.Close()
	gzReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("reading archive file as gzip: %w", err)
	}
	defer gzReader.Close()
	if err := i.fs.MkdirAll(prefix, fs.ModePerm); err != nil {
		return fmt.Errorf("creating prefix folder: %w", err)
	}
	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parsing tar header: %w", err)
		}
		if err := verifyTarPath(header.Name); err != nil {
			return fmt.Errorf("verifying tar path %q: %w", header.Name, err)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if len(header.Name) == 0 {
				return errors.New("cannot create dir for empty path")
			}
			if err := i.fs.Mkdir(path.Join(prefix, header.Name), fs.FileMode(header.Mode)&perm); err != nil && !errors.Is(err, os.ErrExist) {
				return fmt.Errorf("creating folder %s: %w", path.Join(prefix, header.Name), err)
			}
		case tar.TypeReg:
			if len(header.Name) == 0 {
				return errors.New("cannot create file for empty path")
			}
			out, err := i.fs.OpenFile(path.Join(prefix, header.Name), os.O_WRONLY|os.O_CREATE, fs.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("creating file %s for writing: %w", path.Join(prefix, header.Name), err)
			}
			defer out.Close()
			if _, err := io.Copy(out, tarReader); err != nil {
				return fmt.Errorf("writing extracted file contents: %w", err)
			}
		case tar.TypeSymlink:
			if err := verifyTarPath(header.Linkname); err != nil {
				return fmt.Errorf("invalid tar path %q: %w", header.Linkname, err)
			}
			if len(header.Name) == 0 {
				return errors.New("cannot symlink file for empty oldname")
			}
			if len(header.Linkname) == 0 {
				return errors.New("cannot symlink file for empty newname")
			}
			if symlinker, ok := i.fs.Fs.(afero.Symlinker); ok {
				if err := symlinker.SymlinkIfPossible(path.Join(prefix, header.Name), path.Join(prefix, header.Linkname)); err != nil {
					return fmt.Errorf("creating symlink: %w", err)
				}
			} else {
				return errors.New("fs does not support symlinks")
			}
		default:
			return fmt.Errorf("unsupported tar record: %v", header.Typeflag)
		}
	}
}

// downloadToTempDir downloads a file to a temporary location, applying transform on-the-fly.
func (i *osInstaller) downloadToTempDir(ctx context.Context, url string, transforms ...transform.Transformer) (string, error) {
	out, err := afero.TempFile(i.fs, "", "")
	if err != nil {
		return "", fmt.Errorf("creating destination temp file: %w", err)
	}
	defer out.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("request to download %q: %w", url, err)
	}
	resp, err := i.hClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to download %q: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to download %q failed with status code: %v", url, resp.Status)
	}
	defer resp.Body.Close()

	transformReader := transform.NewReader(resp.Body, transform.Chain(transforms...))

	if _, err = io.Copy(out, transformReader); err != nil {
		return "", fmt.Errorf("downloading %q: %w", url, err)
	}
	return out.Name(), nil
}

// copy copies a file from oldname to newname.
func (i *osInstaller) copy(oldname, newname string, perm fs.FileMode) (err error) {
	old, openOldErr := i.fs.OpenFile(oldname, os.O_RDONLY, fs.ModePerm)
	if openOldErr != nil {
		return fmt.Errorf("copying %q to %q: cannot open source file for reading: %w", oldname, newname, openOldErr)
	}
	defer func() { _ = old.Close() }()
	// create destination path if not exists
	if err := i.fs.MkdirAll(path.Dir(newname), fs.ModePerm); err != nil {
		return fmt.Errorf("copying %q to %q: unable to create destination folder: %w", oldname, newname, err)
	}
	new, openNewErr := i.fs.OpenFile(newname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, perm)
	if openNewErr != nil {
		return fmt.Errorf("copying %q to %q: cannot open destination file for writing: %w", oldname, newname, openNewErr)
	}
	defer func() {
		_ = new.Close()
		if err != nil {
			_ = i.fs.Remove(newname)
		}
	}()
	if _, err := io.Copy(new, old); err != nil {
		return fmt.Errorf("copying %q to %q: copying file contents: %w", oldname, newname, err)
	}

	return nil
}

// verifyTarPath checks if a tar path is valid (must not contain ".." as path element).
func verifyTarPath(pat string) error {
	n := len(pat)
	r := 0
	for r < n {
		switch {
		case os.IsPathSeparator(pat[r]):
			// empty path element
			r++
		case pat[r] == '.' && (r+1 == n || os.IsPathSeparator(pat[r+1])):
			// . element
			r++
		case pat[r] == '.' && pat[r+1] == '.' && (r+2 == n || os.IsPathSeparator(pat[r+2])):
			// .. element
			return errors.New("path contains \"..\"")
		default:
			// skip to next path element
			for r < n && !os.IsPathSeparator(pat[r]) {
				r++
			}
		}
	}
	return nil
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
