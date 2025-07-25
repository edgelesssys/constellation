/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package installer provides functionality to install binary components of supported kubernetes versions.
package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/spf13/afero"
	"github.com/vincent-petithory/dataurl"
	"k8s.io/utils/clock"
)

const (
	// determines the period after which retryDownloadToTempDir will retry a download.
	downloadInterval = 10 * time.Millisecond
	executablePerm   = 0o544
)

// OsInstaller installs binary components of supported kubernetes versions.
type OsInstaller struct {
	fs      *afero.Afero
	hClient httpClient
	// clock is needed for testing purposes
	clock clock.WithTicker
	// retriable is the function used to check if an error is retriable. Needed for testing.
	retriable func(error) bool
}

// NewOSInstaller creates a new osInstaller.
func NewOSInstaller() *OsInstaller {
	return &OsInstaller{
		fs:        &afero.Afero{Fs: afero.NewOsFs()},
		hClient:   &http.Client{},
		clock:     clock.RealClock{},
		retriable: isRetriable,
	}
}

// Install downloads a resource from a URL, applies any given text transformations and extracts the resulting file if required.
// The resulting file(s) are copied to the destination. It also verifies the sha256 hash of the downloaded file.
func (i *OsInstaller) Install(ctx context.Context, kubernetesComponent *components.Component) error {
	tempPath, err := i.retryDownloadToTempDir(ctx, kubernetesComponent.Url)
	if err != nil {
		return err
	}

	file, err := i.fs.OpenFile(tempPath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("opening file %q: %w", tempPath, err)
	}
	sha := sha256.New()
	if _, err := io.Copy(sha, file); err != nil {
		return fmt.Errorf("reading file %q: %w", tempPath, err)
	}
	calculatedHash := fmt.Sprintf("sha256:%x", sha.Sum(nil))
	if len(kubernetesComponent.Hash) > 0 && calculatedHash != kubernetesComponent.Hash {
		return fmt.Errorf("hash of file %q %s does not match expected hash %s", tempPath, calculatedHash, kubernetesComponent.Hash)
	}

	defer func() {
		_ = i.fs.Remove(tempPath)
	}()
	if kubernetesComponent.Extract {
		err = i.extractArchive(tempPath, kubernetesComponent.InstallPath, executablePerm)
	} else {
		err = i.copy(tempPath, kubernetesComponent.InstallPath, executablePerm)
	}
	if err != nil {
		return fmt.Errorf("installing from %q: copying to destination %q: %w", kubernetesComponent.Url, kubernetesComponent.InstallPath, err)
	}

	return nil
}

// extractArchive extracts tar gz archives to a prefixed destination.
func (i *OsInstaller) extractArchive(archivePath, prefix string, perm fs.FileMode) error {
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
			prefixedPath := path.Join(prefix, header.Name)
			if err := i.fs.Mkdir(prefixedPath, fs.FileMode(header.Mode)&perm); err != nil && !errors.Is(err, os.ErrExist) {
				return fmt.Errorf("creating folder %q: %w", prefixedPath, err)
			}
		case tar.TypeReg:
			if len(header.Name) == 0 {
				return errors.New("cannot create file for empty path")
			}
			prefixedPath := path.Join(prefix, header.Name)
			out, err := i.fs.OpenFile(prefixedPath, os.O_WRONLY|os.O_CREATE, fs.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("creating file %q for writing: %w", prefixedPath, err)
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

func (i *OsInstaller) retryDownloadToTempDir(ctx context.Context, url string) (fileName string, someError error) {
	doer := downloadDoer{
		url:        url,
		downloader: i,
	}

	// Retries are canceled as soon as the context is canceled.
	// We need to call NewIntervalRetrier with a clock argument so that the tests can fake the clock by changing the osInstaller clock.
	retrier := retry.NewIntervalRetrier(&doer, downloadInterval, i.retriable, i.clock)
	if err := retrier.Do(ctx); err != nil {
		return "", fmt.Errorf("retrying downloadToTempDir: %w", err)
	}

	return doer.path, nil
}

// retriableHTTPStatusCodes are status codes that might flip to 200 if retried.
// This arguably depends on the web server implementation, but below list is
// a reasonable selection, cf. https://stackoverflow.com/a/74627395.
var retriableHTTPStatusCodes = []int{
	http.StatusRequestTimeout,
	http.StatusTooEarly,
	http.StatusTooManyRequests,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// downloadHTTP downloads the given URL with the embedded HTTP client and writes the content to out.
func (i *OsInstaller) downloadHTTP(ctx context.Context, url string, out io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("request to download %q: %w", url, err)
	}
	resp, err := i.hClient.Do(req)
	if err != nil {
		// A failure at this point might be transient, such as network connectivity.
		return fmt.Errorf("request to download %q: %w", url, &retriableError{err: err})
	}
	if resp.StatusCode != http.StatusOK {
		// The HTTP request went through, but the result is not what we
		// expected. Wrap the error return in case we think the request could
		// be retried.
		err = fmt.Errorf("request to download %q failed with status code: %v", url, resp.Status)
		if slices.Contains(retriableHTTPStatusCodes, resp.StatusCode) {
			err = &retriableError{err: err}
		}
		return err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("downloading %q: %w", url, &retriableError{err: err})
	}
	return nil
}

// unpackData parses the given data URL and writes the content to out.
func (i *OsInstaller) unpackData(url string, out io.Writer) error {
	dataURL, err := dataurl.DecodeString(url)
	if err != nil {
		return fmt.Errorf("parsing data URL: %w", err)
	}
	buf := bytes.NewBuffer(dataURL.Data)
	if _, err = io.Copy(out, buf); err != nil {
		return fmt.Errorf("writing content of data URL %q: %w", url, err)
	}
	return nil
}

// downloadToTempDir downloads a file from the given URL to a temporary location and returns the path to the downloaded file.
func (i *OsInstaller) downloadToTempDir(ctx context.Context, u string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("parsing component URL: %w", err)
	}

	out, err := afero.TempFile(i.fs, "", "")
	if err != nil {
		return "", fmt.Errorf("creating destination temp file: %w", err)
	}

	if url.Scheme == "data" {
		err = i.unpackData(u, out)
	} else {
		err = i.downloadHTTP(ctx, u, out)
	}
	out.Close()
	if err != nil {
		removeErr := i.fs.Remove(out.Name())
		return "", errors.Join(err, removeErr)
	}
	return out.Name(), nil
}

// copy copies a file from oldname to newname.
func (i *OsInstaller) copy(oldname, newname string, perm fs.FileMode) (err error) {
	old, openOldErr := i.fs.OpenFile(oldname, os.O_RDONLY, fs.ModePerm)
	if openOldErr != nil {
		return fmt.Errorf("copying %q to %q: cannot open source file for reading: %w", oldname, newname, openOldErr)
	}
	defer func() { _ = old.Close() }()
	// create destination path if not exists
	if err := i.fs.MkdirAll(path.Dir(newname), fs.ModePerm); err != nil {
		return fmt.Errorf("copying %q to %q: unable to create destination folder: %w", oldname, newname, err)
	}
	newFile, openNewErr := i.fs.OpenFile(newname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, perm)
	if openNewErr != nil {
		return fmt.Errorf("copying %q to %q: cannot open destination file for writing: %w", oldname, newname, openNewErr)
	}
	defer func() {
		_ = newFile.Close()
		if err != nil {
			_ = i.fs.Remove(newname)
		}
	}()
	if _, err := io.Copy(newFile, old); err != nil {
		return fmt.Errorf("copying %q to %q: copying file contents: %w", oldname, newname, err)
	}

	return nil
}

type downloadDoer struct {
	url        string
	downloader downloader
	path       string
}

type downloader interface {
	downloadToTempDir(ctx context.Context, url string) (string, error)
}

func (d *downloadDoer) Do(ctx context.Context) error {
	path, err := d.downloader.downloadToTempDir(ctx, d.url)
	d.path = path
	return err
}

// retriableError is an error that can be retried.
type retriableError struct{ err error }

func (e *retriableError) Error() string {
	return fmt.Sprintf("retriable error: %s", e.err.Error())
}

func (e *retriableError) Unwrap() error { return e.err }

// isRetriable returns true if the action resulting in this error can be retried.
func isRetriable(err error) bool {
	retriableError := &retriableError{}
	return errors.As(err, &retriableError)
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
