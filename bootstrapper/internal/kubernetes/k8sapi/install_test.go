package k8sapi

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
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/icholy/replace"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/transform"
	"google.golang.org/grpc/test/bufconn"
)

func TestInstall(t *testing.T) {
	testCases := map[string]struct {
		server      httpBufconnServer
		destination string
		extract     bool
		transforms  []transform.Transformer
		readonly    bool
		wantErr     bool
		wantFiles   map[string][]byte
	}{
		"download works": {
			server:      newHTTPBufconnServerWithBody([]byte("file-contents")),
			destination: "/destination",
			wantFiles:   map[string][]byte{"/destination": []byte("file-contents")},
		},
		"download with extract works": {
			server:      newHTTPBufconnServerWithBody(createTarGz([]byte("file-contents"), "/destination")),
			destination: "/prefix",
			extract:     true,
			wantFiles:   map[string][]byte{"/prefix/destination": []byte("file-contents")},
		},
		"download with transform works": {
			server:      newHTTPBufconnServerWithBody([]byte("/usr/bin/kubelet")),
			destination: "/destination",
			transforms: []transform.Transformer{
				replace.String("/usr/bin", "/run/state/bin"),
			},
			wantFiles: map[string][]byte{"/destination": []byte("/run/state/bin/kubelet")},
		},
		"download fails": {
			server:      newHTTPBufconnServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }),
			destination: "/destination",
			wantErr:     true,
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

			inst := osInstaller{
				fs:      &afero.Afero{Fs: afero.NewMemMapFs()},
				hClient: &hClient,
			}
			err := inst.Install(context.Background(), "http://server/path", []string{tc.destination}, fs.ModePerm, tc.extract, tc.transforms...)
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

			inst := osInstaller{
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

func TestDownloadToTempDir(t *testing.T) {
	testCases := map[string]struct {
		server     httpBufconnServer
		transforms []transform.Transformer
		readonly   bool
		wantErr    bool
		wantFile   []byte
	}{
		"download works": {
			server:   newHTTPBufconnServerWithBody([]byte("file-contents")),
			wantFile: []byte("file-contents"),
		},
		"download with transform works": {
			server: newHTTPBufconnServerWithBody([]byte("/usr/bin/kubelet")),
			transforms: []transform.Transformer{
				replace.String("/usr/bin", "/run/state/bin"),
			},
			wantFile: []byte("/run/state/bin/kubelet"),
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
			inst := osInstaller{
				fs:      &afero.Afero{Fs: afs},
				hClient: &hClient,
			}
			path, err := inst.downloadToTempDir(context.Background(), "http://server/path", tc.transforms...)
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

			inst := osInstaller{fs: &afero.Afero{Fs: afs}}
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
		result = multierror.Append(result, err)
	}
	if err := w.gzWriter.Close(); err != nil {
		result = multierror.Append(result, err)
	}
	return result
}

func stringPtr(s string) *string {
	return &s
}
