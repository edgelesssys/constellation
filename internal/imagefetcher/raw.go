/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package imagefetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

// Downloader downloads raw images.
type Downloader struct {
	httpc httpc
	fs    *afero.Afero
}

// NewDownloader creates a new Downloader.
func NewDownloader() *Downloader {
	return &Downloader{
		httpc: http.DefaultClient,
		fs:    &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// Download downloads the raw image from source.
func (d *Downloader) Download(ctx context.Context, errWriter io.Writer, showBar bool, source, imageName string) (string, error) {
	url, err := url.Parse(source)
	if err != nil {
		return "", fmt.Errorf("parsing image source URL: %w", err)
	}
	imageName = filepath.Base(imageName)
	var partfile, destination string
	switch url.Scheme {
	case "http", "https":
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting current working directory: %w", err)
		}
		partfile = filepath.Join(cwd, imageName+".raw.part")
		destination = filepath.Join(cwd, imageName+".raw")
	case "file":
		return url.Path, nil
	default:
		return "", fmt.Errorf("unsupported image source URL scheme: %s", url.Scheme)
	}
	if !d.shouldDownload(destination) {
		return destination, nil
	}
	if err := d.downloadWithProgress(ctx, errWriter, showBar, source, partfile); err != nil {
		return "", err
	}
	return destination, d.fs.Rename(partfile, destination)
}

// shouldDownload checks if the image should be downloaded.
func (d *Downloader) shouldDownload(destination string) bool {
	_, err := d.fs.Stat(destination)
	return errors.Is(err, fs.ErrNotExist)
}

// downloadWithProgress downloads the raw image from source to the destination.
func (d *Downloader) downloadWithProgress(ctx context.Context, errWriter io.Writer, showBar bool, source, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := d.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading from %q: %s", source, resp.Status)
	}

	f, err := d.fs.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var bar io.WriteCloser
	if showBar {
		bar = prepareBar(errWriter, resp.ContentLength)
	} else {
		bar = &nopWriteCloser{}
	}
	defer bar.Close()

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func prepareBar(writer io.Writer, total int64) io.WriteCloser {
	return progressbar.NewOptions64(
		total,
		progressbar.OptionSetWriter(writer),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() { fmt.Fprintf(writer, "Done.\n\n") }),
	)
}

type nopWriteCloser struct{}

func (*nopWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (*nopWriteCloser) Close() error {
	return nil
}

type httpc interface {
	Do(req *http.Request) (*http.Response, error)
}
