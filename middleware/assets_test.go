package middleware

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

type (
	testFilesystemStatError struct {
		fstest.MapFS
		stat func(name string) error
	}

	testFileStatError struct {
		fs.File
		err error
	}
)

func (fsys testFilesystemStatError) Open(name string) (fs.File, error) {
	f, err := fsys.MapFS.Open(name)
	if err != nil {
		return nil, err
	}
	if err = fsys.stat(name); err != nil {
		return testFileStatError{File: f, err: err}, nil
	}
	return f, nil
}

func (f testFileStatError) Stat() (fs.FileInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.File.Stat()
}

func TestAssetsHandlerPathEscapeError(t *testing.T) {
	handler := NewAssetsHandler(AssetsConfig{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	req.URL.Path = "/%.html"

	handler.ServeHTTP(rec, req)
	assert.Equal(t, 400, rec.Code)
}

func TestAssetsHandlerStatError(t *testing.T) {
	filesystem := testFilesystemStatError{
		MapFS: fstest.MapFS{
			"index.html": {
				Data: []byte("index.html"),
			},
			"dir": {
				Mode: fs.ModeDir,
			},
			"dir/index.html": {
				Data: []byte("dir/index.html"),
			},
		},
		stat: func(name string) error {
			if name == "index.html" || name == "dir/index.html" {
				return errors.New("stat-error")
			}
			return nil
		},
	}

	handler := NewAssetsHandler(AssetsConfig{
		Filesystem: http.FS(filesystem),
	})
	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/dir/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 404, rec.Code)
}

func TestAssetsHandler(t *testing.T) {
	type testReq struct {
		uri        string
		expectCode int
		expectBody string
	}

	const (
		bodyIndex    = "<html><body>Hello</body></html>"
		bodyNewIndex = "<html><body>Hello New Index</body></html>"
		bodyExists   = "<html><body>Hello Exists</body></html>"
		bodyDirIndex = "<html><body>Hello Dir Index</body></html>"
	)

	tests := []struct {
		name   string
		config AssetsConfig
		reqs   []testReq
	}{{
		name:   "config:empty",
		config: AssetsConfig{},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 404,
		}, {
			uri:        "/index.html",
			expectCode: 404,
		}, {
			uri:        "/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:normal",
		config: AssetsConfig{
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
				"dir": {
					Mode: fs.ModeDir,
				},
				"dir/index.html": {
					Data: []byte(bodyDirIndex),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/not-exists.html",
			expectCode: 404,
		}, {
			uri:        "/dir/",
			expectCode: 200,
			expectBody: bodyDirIndex,
		}, {
			uri:        "/dir/index.html",
			expectCode: 200,
			expectBody: bodyDirIndex,
		}, {
			uri:        "/dir/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:prefix",
		config: AssetsConfig{
			Prefix: "/assets",
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 404,
		}, {
			uri:        "/assets/",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/index.html",
			expectCode: 404,
		}, {
			uri:        "/assets/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/exists.html",
			expectCode: 404,
		}, {
			uri:        "/assets/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/assets/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:index:notexists",
		config: AssetsConfig{
			Index: "/new-index.html",
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 404,
		}, {
			uri:        "/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/new-index.html",
			expectCode: 404,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:index:exists",
		config: AssetsConfig{
			Index: "/new-index.html",
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"new-index.html": {
					Data: []byte(bodyNewIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 200,
			expectBody: bodyNewIndex,
		}, {
			uri:        "/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/new-index.html",
			expectCode: 200,
			expectBody: bodyNewIndex,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:spa",
		config: AssetsConfig{
			SPA: true,
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/not-exists.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}},
	}, {
		name: "config:spa:notexists",
		config: AssetsConfig{
			SPA: true,
			Filesystem: http.FS(fstest.MapFS{
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 404,
		}, {
			uri:        "/index.html",
			expectCode: 404,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
		}, {
			uri:        "/not-exists.html",
			expectCode: 404,
		}},
	}, {
		name: "config:spa:prefix",
		config: AssetsConfig{
			Prefix: "/assets",
			SPA:    true,
			Filesystem: http.FS(fstest.MapFS{
				"index.html": {
					Data: []byte(bodyIndex),
				},
				"exists.html": {
					Data: []byte(bodyExists),
				},
			}),
		},
		reqs: []testReq{{
			uri:        "/",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/assets/",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri: "/index.html",

			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/assets/index.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/exists.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}, {
			uri:        "/assets/exists.html",
			expectCode: 200,
			expectBody: bodyExists,
		}, {
			uri:        "/assets/not-exists.html",
			expectCode: 200,
			expectBody: bodyIndex,
		}},
	}}

	for _, test := range tests {
		handler := NewAssetsHandler(test.config)
		for i := range test.reqs {
			t.Run(fmt.Sprintf("%s:%s", test.name, test.reqs[i].uri), func(t *testing.T) {
				ast := assert.New(t)
				req := httptest.NewRequest(http.MethodGet, test.reqs[i].uri, nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				ast.Equal(test.reqs[i].expectCode, rec.Code)
				if test.reqs[i].expectBody != "" {
					ast.Equal(test.reqs[i].expectBody, rec.Body.String())
				}
			})
		}
	}
}
