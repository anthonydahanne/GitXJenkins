package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/template.html": {
		local:   "template.html",
		size:    3023,
		modtime: 1454996477,
		compressed: `
H4sIAAAJbogA/7xWbXPaOBD+nl+h+jITmI5tCCVJc8BMEmg7TF5IoOTab8KWsYwsGUkOEMp/PwkbsHnJ
JV/O7Uyk1Wpfnn12Re1T8+Gm96vTAr4MSeOopv8AAumwbiBqaAGCbuMIqK8WIgmB40MukKwbP3vfzAsj
PZJYEtSYz62eXiwWidROxMlGOBxHEgju1A3bdpiLrGAcIz6zHBbaydIsW+VTq2SFmFqBMBo1O7mVmvhk
muAWSiQkUHciTJALIHWBUsceVpubbheYZqpNMB0BjkjdEHJGkPARkgbwOfLqhi9lJC5tO4RTx6XWgDEp
JIeR3uhw1gK7YlWsM9sRYiNbhqckxtJP8mEq0ZBjOVPefFi5+GKWxxdhr/1w1Z1eBOWr+DOsPjf7tINP
yeibN5m0ruCF32y6wW8S3aLh1A/6d62yNwyeO9/D0as4N4DDmRCM4yGmdQNSRmchixUqRxs0HiKJGYUE
SB+F6H/I3Vw6eh8C3u3z6X2pTO7GARxdj6YVYt99bUE/nkRdD92/9M8q7Sp6pZX49yuMeqX4vPVL/HMX
PPY/l1q0yt+FwJt8aMMX2E14t4Ymy8OPQhFssyDIQbAPhFLYHbSbrR8YEi+Mr68fO2dXXx55xMfVh773
XDnvPD1VgmrrdjwVojzrjx8kRRH90e98he3zaTdv/xAgm0Y52tttKh8XSvVvoMhgUSRt1Wkl9V+nlLah
VuglCgf67yNGM9Z2IMuG+590fdOR5uZBT5qgW4A01mgeF1zmxCGismhxNeNmBS+mju4mUCiCeQ7148LJ
XxxFTGCpsEdi6e2kaDVXngs59dU3N3ATi4jA2S2iQ+kbl+C0usipFv9ebxfpOoNNzU6mb23A3Jne++XG
dyzBUyYWsERFaZa1gotfgEOgEHXDYVRCTBE3PRJjdzWpow0EPR8BJ+ZcYQB0elyCCRRgiNQl1VUuUFio
ka6yVBMdTLD09ZgBHiOETTAdKrZ7jIdwCdplGny0qmkmFM4mxsZtPkhiToVZPgV6FbrrF2WtvExvpb7c
bGkkWglCOxWYc/WUIWDdMOrhYYsgXXCxWOyxwHevJweueteAdQ/VfF0s1JvmHlaswZS0+sZPTtQFo7FZ
12zY2G9ASff4n88REeij0dZUBzA6bFAGFPBC8SnZf9Azdfc5tvcgrYS6LpkK26rEKd8yyy3qDZCiUUJf
gAUQfkKqiQ8lgBwtudZGdISpAAEbCICmWEitolgHEHR8oJth3ZgzC9QG3G5sEzH14CoebvewkSNWEovp
sxfEs3yVm99Ae8FXCo0OZwFypILC3z3U3Nl/0lZ55U/y9dBnW87z+K8ZvpwJV9TVFjOFU8GuchSx46Dl
SMxHsUXbNJMse1PRQQa/Sf114xy6uU4hCX2vIZCxBLQp8OfPiqR5q3n8djpoT/He2zE7lnMdkuuMdUek
9M+9PscFQ7qXKIzkzChaEdTzt1C0oOve6EIVTiaQU0XzE/0whYqMqTit30nxaPNE1OzEqZr/+vf7vwEA
AP//jIj0uM8LAAA=
`,
	},

	"/": {
		isDir: true,
		local: "/",
	},
}
