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
		size:    3489,
		modtime: 1456116541,
		compressed: `
H4sIAAAJbogA/7xXW3PaPBO+z69Q/WUmMP1sQ2iaNC8wkwTaDpM0tFDS9k5YMpZjy0aSA5Ty39+VbQ7m
kCY3r9uZ6LDSrp599pGov2nd3/R/dtvIU2HQPKrrPyjAfNQwKDf0AMWkeYTgq4dUYeR4WEiqGsb3/kfz
wsinFFMBbc7nVl83Fots1M6Gs450BIsVksJpGLbtRIRa/jihYmY5UWhnTbNqVU+tihUybvnSaNbtbFW+
xRvTRLdYUakQrIlZQAnCnCAwZy6Dzk2vh0wztw4Yf0SCBg1DqllApUepMpAnqNswPKVieWnbIZ46hFvD
KFJSCRzrjg5nNWDXrJr13nakXI+l4cGIkfrJPsYVHQmmZuDNw7WLd2Z1fBH2O/dXvemFX71K3uKzh9aA
d9lp8PjRnUzaV/jCa7WI/yuIb+lo6vmDu3bVHfkP3U/h4295biBHRFJGgo0YbxiYR3wWRgmgcrRG4z5W
LOI4QMqjIf0Pzm6mjl6GgHv7cPqlUg3uxj5+vH6c1gL77kMbe8kk7rn0y9Pgfa1zRn/zWvLrN477leS8
/VP+uPO/Dt5W2vxMvAiBZ/nQwU+4l/FuBc0mD18Lhb/NAr8AwT4QKmFv2Gm1PzMcuGFyff21+/7q3VcR
i/HZ/cB9qJ13v32r+Wft2/FUyupsML5XnMb886D7AXfOp73i/ocAWRfK0d5qg/MQrODfEMhgcapsqLQK
/NdHystQG/QzgwP195pNN3bbgWwz3L/S9VlHmpsHPWmCbgHSXKF5XCKRk4SUq7IlQONmJTfhjq4mVCqj
eQH149LJ/wSNI8kUYE9l6u2kbLWWnksF8+U3N1iLyTjAs1vKR8ozLlG1UlkUbMv/rLqLvL0BTt3O5Lc+
jMhM971q8xNT6NtGMCiFBSyr2oCwJ+QEWMqG4URcYcapMN0gYWQp1fEag75HkZMIASAgfT6h0ARLNKKw
CMqKIAADNB2OCZKOJkx5WmeQGwVBNGF8BHR3IxHiFLXLPPh4mdSNUEQ0MdZui0EG5lSa1VOkWyFZXSkr
4/R4S/O0s2WRWWUI7aRgLuAuo8i6ibjLRu2A6ozLxWLPDmJ3eTZB4GJD1hcMArtYwKVGDhvWcc5aveK7
CGCB0Vy36zZu7t8ARvf4n89pIOlro61DCUR81OQRAuAl8Cnrv9IzJ/sc23uQhkGdl40M25DinG8bzS3q
DSnQKKMvYhJJLyPVxMMKYUFTrnUof2RcIj8aSkSnTCptAqxDFDse0sWwqsyZhepDYTe3iZh7IMDD7SI2
CsTKYjG96ImKTb6q9SNoL/hg0OyKyKeOAii83UnNnf0zHThXcaaYDz235byI/4rhqSZccaJ33EgcBLs8
o0wch6aaWIxii7b5STbZmw8dZPCz1F8VzqGVB6v2mHFCp/9Hx5B9dNlA1tbZ1vbMzY0Xiz9/DnIX75dp
vZqOUy/WLZbqOmEB6cF9k0hUAeFbUoROlbnEMC9N9MzSKtpaS/ShhLGq6sJkmIDgGgfrTn9/hfdFVVzM
QZFtO3qzh+ov1ZednQuRFHSkoB+bKtGLEuGA+MOPhfRhB0qgRJLe0hKBSqw4t3zIjeCOSobp2w1z5cH7
iGAPc05tUIofuZgYzVeZr5BNJSXXs8J74rhkKHJJw1jNjLIVY32hlsoWJuRGZ7h0MsGCg26d6KdGCOqS
D+dkOikfre/8up3hAhe6/kX2bwAAAP//omeuUKENAAA=
`,
	},

	"/": {
		isDir: true,
		local: "/",
	},
}
