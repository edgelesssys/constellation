/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package installer

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"google.golang.org/grpc/test/bufconn"
	testclock "k8s.io/utils/clock/testing"
)

func TestInstall(t *testing.T) {
	serverURL := "http://server/path"
	testCases := map[string]struct {
		server      httpBufconnServer
		component   versions.ComponentVersion
		hash        string
		destination string
		extract     bool
		wantErr     bool
		wantFiles   map[string][]byte
	}{
		"download works": {
			server: newHTTPBufconnServerWithBody([]byte("file-contents")),
			component: versions.ComponentVersion{
				URL:         serverURL,
				Hash:        "sha256:f03779b36bece74893fd6533a67549675e21573eb0e288d87158738f9c24594e",
				InstallPath: "/destination",
			},
			wantFiles: map[string][]byte{"/destination": []byte("file-contents")},
		},
		"download with extract works": {
			server: newHTTPBufconnServerWithBody(createTarGz([]byte("file-contents"), "/destination")),
			component: versions.ComponentVersion{
				URL:         serverURL,
				Hash:        "sha256:a52a1664ca0a6ec9790384e3d058852ab8b3a8f389a9113d150fdc6ab308d949",
				InstallPath: "/prefix",
				Extract:     true,
			},
			wantFiles: map[string][]byte{"/prefix/destination": []byte("file-contents")},
		},
		"hash validation fails": {
			server: newHTTPBufconnServerWithBody([]byte("file-contents")),
			component: versions.ComponentVersion{
				URL:         serverURL,
				Hash:        "sha256:abc",
				InstallPath: "/destination",
			},
			wantErr: true,
		},
		"download fails": {
			server: newHTTPBufconnServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }),
			component: versions.ComponentVersion{
				URL:         serverURL,
				Hash:        "sha256:abc",
				InstallPath: "/destination",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			defer tc.server.Close()

			hClient := http.Client{
				Transport: &http.Transport{
					DialContext:    tc.server.DialContext,
					Dial:           tc.server.Dial,
					DialTLSContext: tc.server.DialContext,
					DialTLS:        tc.server.Dial,
				},
			}

			// This test was written before retriability was added to Install. It makes sense to test Install as if it wouldn't retry requests.
			inst := OsInstaller{
				fs:        &afero.Afero{Fs: afero.NewMemMapFs()},
				hClient:   &hClient,
				clock:     testclock.NewFakeClock(time.Time{}),
				retriable: func(err error) bool { return false },
			}

			err := inst.Install(context.Background(), tc.component)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			for path, wantContents := range tc.wantFiles {
				contents, err := inst.fs.ReadFile(path)
				assert.NoError(err)
				assert.Equal(wantContents, contents)
			}
		})
	}
}

func TestExtractArchive(t *testing.T) {
	tarGzTestFile := createTarGz([]byte("file-contents"), "/destination")
	tarGzTestWithFolder := createTarGzWithFolder([]byte("file-contents"), "/folder/destination", nil)

	testCases := map[string]struct {
		source      string
		destination string
		contents    []byte
		readonly    bool
		wantErr     bool
		wantFiles   map[string][]byte
	}{
		"extract works": {
			source:      "in.tar.gz",
			destination: "/prefix",
			contents:    tarGzTestFile,
			wantFiles: map[string][]byte{
				"/prefix/destination": []byte("file-contents"),
			},
		},
		"extract with folder works": {
			source:      "in.tar.gz",
			destination: "/prefix",
			contents:    tarGzTestWithFolder,
			wantFiles: map[string][]byte{
				"/prefix/folder/destination": []byte("file-contents"),
			},
		},
		"source missing": {
			source:      "in.tar.gz",
			destination: "/prefix",
			wantErr:     true,
		},
		"non-gzip file contents": {
			source:      "in.tar.gz",
			contents:    []byte("invalid bytes"),
			destination: "/prefix",
			wantErr:     true,
		},
		"non-tar file contents": {
			source:      "in.tar.gz",
			contents:    createGz([]byte("file-contents")),
			destination: "/prefix",
			wantErr:     true,
		},
		"mkdir prefix dir fails on RO fs": {
			source:      "in.tar.gz",
			contents:    tarGzTestFile,
			destination: "/prefix",
			readonly:    true,
			wantErr:     true,
		},
		"mkdir tar dir fails on RO fs": {
			source:      "in.tar.gz",
			contents:    tarGzTestWithFolder,
			destination: "/",
			readonly:    true,
			wantErr:     true,
		},
		"writing tar file fails on RO fs": {
			source:      "in.tar.gz",
			contents:    tarGzTestFile,
			destination: "/",
			readonly:    true,
			wantErr:     true,
		},
		"symlink can be detected (but is unsupported on memmapfs)": {
			source:      "in.tar.gz",
			contents:    createTarGzWithSymlink("source", "dest"),
			destination: "/prefix",
			wantErr:     true,
		},
		"unsupported tar header type is detected": {
			source:      "in.tar.gz",
			contents:    createTarGzWithFifo("/destination"),
			destination: "/prefix",
			wantErr:     true,
		},
		"path traversal is detected": {
			source:   "in.tar.gz",
			contents: createTarGz([]byte{}, "../destination"),
			wantErr:  true,
		},
		"path traversal in symlink is detected": {
			source:   "in.tar.gz",
			contents: createTarGzWithSymlink("/source", "../destination"),
			wantErr:  true,
		},
		"empty file name is detected": {
			source:   "in.tar.gz",
			contents: createTarGz([]byte{}, ""),
			wantErr:  true,
		},
		"empty folder name is detected": {
			source:   "in.tar.gz",
			contents: createTarGzWithFolder([]byte{}, "source", stringPtr("")),
			wantErr:  true,
		},
		"empty symlink source is detected": {
			source:   "in.tar.gz",
			contents: createTarGzWithSymlink("", "/target"),
			wantErr:  true,
		},
		"empty symlink target is detected": {
			source:   "in.tar.gz",
			contents: createTarGzWithSymlink("/source", ""),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			afs := afero.NewMemMapFs()
			if len(tc.source) > 0 && len(tc.contents) > 0 {
				require.NoError(afero.WriteFile(afs, tc.source, tc.contents, fs.ModePerm))
			}

			if tc.readonly {
				afs = afero.NewReadOnlyFs(afs)
			}

			inst := OsInstaller{
				fs: &afero.Afero{Fs: afs},
			}
			err := inst.extractArchive(tc.source, tc.destination, fs.ModePerm)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			for path, wantContents := range tc.wantFiles {
				contents, err := inst.fs.ReadFile(path)
				assert.NoError(err)
				assert.Equal(wantContents, contents)
			}
		})
	}
}

func TestRetryDownloadToTempDir(t *testing.T) {
	testCases := map[string]struct {
		responses []int
		cancelCtx bool
		wantErr   bool
		wantFile  []byte
	}{
		"Succeed on third try": {
			responses: []int{500, 500, 200},
			wantFile:  []byte("file-content"),
		},
		"Cancel after second try": {
			responses: []int{500, 500},
			cancelCtx: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// control the server's responses through stateCh
			stateCh := make(chan int)
			server := newHTTPBufconnServerWithState(stateCh, tc.wantFile)
			defer server.Close()

			hClient := http.Client{
				Transport: &http.Transport{
					DialContext:    server.DialContext,
					Dial:           server.Dial,
					DialTLSContext: server.DialContext,
					DialTLS:        server.Dial,
				},
			}

			afs := afero.NewMemMapFs()

			// control download retries through FakeClock clock
			clock := testclock.NewFakeClock(time.Now())
			inst := OsInstaller{
				fs:        &afero.Afero{Fs: afs},
				hClient:   &hClient,
				clock:     clock,
				retriable: func(error) bool { return true },
			}

			// abort retryDownloadToTempDir in some test cases by using the context
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			wg := sync.WaitGroup{}
			var downloadErr error
			var path string
			wg.Add(1)
			go func() {
				defer wg.Done()
				path, downloadErr = inst.retryDownloadToTempDir(ctx, "http://server/path")
			}()

			// control the server's responses through stateCh.
			for _, resp := range tc.responses {
				stateCh <- resp
				clock.Step(downloadInterval)
			}
			if tc.cancelCtx {
				cancel()
			}

			wg.Wait()

			if tc.wantErr {
				assert.Error(downloadErr)
				return
			}

			require.NoError(downloadErr)
			content, err := inst.fs.ReadFile(path)
			assert.NoError(err)
			assert.Equal(tc.wantFile, content)
		})
	}
}

func TestDownloadToTempDir(t *testing.T) {
	testCases := map[string]struct {
		server   httpBufconnServer
		readonly bool
		wantErr  bool
		wantFile []byte
	}{
		"download works": {
			server:   newHTTPBufconnServerWithBody([]byte("file-contents")),
			wantFile: []byte("file-contents"),
		},
		"download fails": {
			server:  newHTTPBufconnServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }),
			wantErr: true,
		},
		"creating temp file fails on RO fs": {
			server:   newHTTPBufconnServerWithBody([]byte("file-contents")),
			readonly: true,
			wantErr:  true,
		},
		"content length mismatch": {
			server: newHTTPBufconnServer(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Length", "1337")
				writer.WriteHeader(200)
			}),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			defer tc.server.Close()

			hClient := http.Client{
				Transport: &http.Transport{
					DialContext:    tc.server.DialContext,
					Dial:           tc.server.Dial,
					DialTLSContext: tc.server.DialContext,
					DialTLS:        tc.server.Dial,
				},
			}

			afs := afero.NewMemMapFs()
			if tc.readonly {
				afs = afero.NewReadOnlyFs(afs)
			}
			inst := OsInstaller{
				fs:      &afero.Afero{Fs: afs},
				hClient: &hClient,
			}
			path, err := inst.downloadToTempDir(context.Background(), "http://server/path")
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			contents, err := inst.fs.ReadFile(path)
			assert.NoError(err)
			assert.Equal(tc.wantFile, contents)
		})
	}
}

func TestCopy(t *testing.T) {
	contents := []byte("file-contents")
	existingFile := "/source"
	testCases := map[string]struct {
		oldname  string
		newname  string
		perm     fs.FileMode
		readonly bool
		wantErr  bool
	}{
		"copy works": {
			oldname: existingFile,
			newname: "/destination",
			perm:    fs.ModePerm,
		},
		"oldname does not exist": {
			oldname: "missing",
			newname: "/destination",
			wantErr: true,
		},
		"copy on readonly fs fails": {
			oldname: existingFile,
			newname: "/destination",
			perm:    fs.ModePerm,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			afs := afero.NewMemMapFs()
			require.NoError(afero.WriteFile(afs, existingFile, contents, fs.ModePerm))

			if tc.readonly {
				afs = afero.NewReadOnlyFs(afs)
			}

			inst := OsInstaller{fs: &afero.Afero{Fs: afs}}
			err := inst.copy(tc.oldname, tc.newname, tc.perm)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)

			oldfile, err := afs.Open(tc.oldname)
			assert.NoError(err)
			newfile, err := afs.Open(tc.newname)
			assert.NoError(err)

			oldContents, _ := io.ReadAll(oldfile)
			newContents, _ := io.ReadAll(newfile)
			assert.Equal(oldContents, newContents)

			newStat, _ := newfile.Stat()
			assert.Equal(tc.perm, newStat.Mode())
		})
	}
}

func TestVerifyTarPath(t *testing.T) {
	testCases := map[string]struct {
		path    string
		wantErr bool
	}{
		"valid relative path": {
			path: "a/b/c",
		},
		"valid absolute path": {
			path: "/a/b/c",
		},
		"valid path with dot": {
			path: "/a/b/.d",
		},
		"valid path with dots": {
			path: "/a/b/..d",
		},
		"single dot in path is allowed": {
			path: ".",
		},
		"simple path traversal": {
			path:    "..",
			wantErr: true,
		},
		"simple path traversal 2": {
			path:    "../",
			wantErr: true,
		},
		"simple path traversal 3": {
			path:    "/..",
			wantErr: true,
		},
		"simple path traversal 4": {
			path:    "/../",
			wantErr: true,
		},
		"complex relative path traversal": {
			path:    "a/b/c/../../../../c/d/e",
			wantErr: true,
		},
		"complex absolute path traversal": {
			path:    "/a/b/c/../../../../c/d/e",
			wantErr: true,
		},
		"path traversal at the end": {
			path:    "a/..",
			wantErr: true,
		},
		"path traversal at the end with trailing /": {
			path:    "a/../",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := verifyTarPath(tc.path)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.path, path.Clean(tc.path))
		})
	}
}

type httpBufconnServer struct {
	*httptest.Server
	*bufconn.Listener
}

func (s *httpBufconnServer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return s.Listener.DialContext(ctx)
}

func (s *httpBufconnServer) Dial(network, addr string) (net.Conn, error) {
	return s.Listener.Dial()
}

func (s *httpBufconnServer) Close() {
	s.Server.Close()
	s.Listener.Close()
}

func newHTTPBufconnServer(handlerFunc http.HandlerFunc) httpBufconnServer {
	server := httptest.NewUnstartedServer(handlerFunc)
	listener := bufconn.Listen(1024)
	server.Listener = listener
	server.Start()
	return httpBufconnServer{
		Server:   server,
		Listener: listener,
	}
}

func newHTTPBufconnServerWithBody(body []byte) httpBufconnServer {
	return newHTTPBufconnServer(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write(body); err != nil {
			panic(err)
		}
	})
}

func newHTTPBufconnServerWithState(state chan int, body []byte) httpBufconnServer {
	return newHTTPBufconnServer(func(w http.ResponseWriter, r *http.Request) {
		switch <-state {
		case 500:
			w.WriteHeader(500)
		case 200:
			if _, err := w.Write(body); err != nil {
				panic(err)
			}
		default:
			w.WriteHeader(402)
		}
	})
}

func createTarGz(contents []byte, path string) []byte {
	tgzWriter := newTarGzWriter()
	defer func() { _ = tgzWriter.Close() }()

	if err := tgzWriter.writeHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     path,
		Size:     int64(len(contents)),
		Mode:     int64(fs.ModePerm),
	}); err != nil {
		panic(err)
	}
	if _, err := tgzWriter.writeTar(contents); err != nil {
		panic(err)
	}

	return tgzWriter.Bytes()
}

func createTarGzWithFolder(contents []byte, pat string, dirnameOverride *string) []byte {
	tgzWriter := newTarGzWriter()
	defer func() { _ = tgzWriter.Close() }()

	dir := path.Dir(pat)
	if dirnameOverride != nil {
		dir = *dirnameOverride
	}

	if err := tgzWriter.writeHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     dir,
		Mode:     int64(fs.ModePerm),
	}); err != nil {
		panic(err)
	}
	if err := tgzWriter.writeHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     pat,
		Size:     int64(len(contents)),
		Mode:     int64(fs.ModePerm),
	}); err != nil {
		panic(err)
	}
	if _, err := tgzWriter.writeTar(contents); err != nil {
		panic(err)
	}

	return tgzWriter.Bytes()
}

func createTarGzWithSymlink(oldname, newname string) []byte {
	tgzWriter := newTarGzWriter()
	defer func() { _ = tgzWriter.Close() }()

	if err := tgzWriter.writeHeader(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     oldname,
		Linkname: newname,
		Mode:     int64(fs.ModePerm),
	}); err != nil {
		panic(err)
	}

	return tgzWriter.Bytes()
}

func createTarGzWithFifo(name string) []byte {
	tgzWriter := newTarGzWriter()
	defer func() { _ = tgzWriter.Close() }()

	if err := tgzWriter.writeHeader(&tar.Header{
		Typeflag: tar.TypeFifo,
		Name:     name,
		Mode:     int64(fs.ModePerm),
	}); err != nil {
		panic(err)
	}

	return tgzWriter.Bytes()
}

func createGz(contents []byte) []byte {
	tgzWriter := newTarGzWriter()
	defer func() { _ = tgzWriter.Close() }()

	if _, err := tgzWriter.writeGz(contents); err != nil {
		panic(err)
	}

	return tgzWriter.Bytes()
}

type tarGzWriter struct {
	buf       *bytes.Buffer
	bufWriter *bufio.Writer
	gzWriter  *gzip.Writer
	tarWriter *tar.Writer
}

func newTarGzWriter() *tarGzWriter {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	gzipWriter := gzip.NewWriter(bufWriter)
	tarWriter := tar.NewWriter(gzipWriter)

	return &tarGzWriter{
		buf:       &buf,
		bufWriter: bufWriter,
		gzWriter:  gzipWriter,
		tarWriter: tarWriter,
	}
}

func (w *tarGzWriter) writeHeader(hdr *tar.Header) error {
	return w.tarWriter.WriteHeader(hdr)
}

func (w *tarGzWriter) writeTar(b []byte) (int, error) {
	return w.tarWriter.Write(b)
}

func (w *tarGzWriter) writeGz(b []byte) (int, error) {
	return w.gzWriter.Write(b)
}

func (w *tarGzWriter) Bytes() []byte {
	_ = w.tarWriter.Flush()
	_ = w.gzWriter.Flush()
	_ = w.gzWriter.Close() // required to ensure clean EOF in gz reader
	_ = w.bufWriter.Flush()
	return w.buf.Bytes()
}

func (w *tarGzWriter) Close() (result error) {
	if err := w.tarWriter.Close(); err != nil {
		result = multierr.Append(result, err)
	}
	if err := w.gzWriter.Close(); err != nil {
		result = multierr.Append(result, err)
	}
	return result
}

func stringPtr(s string) *string {
	return &s
}
