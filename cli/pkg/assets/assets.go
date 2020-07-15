// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// raw-assets/Dockerfile
// raw-assets/baseimages-base.Dockerfile
// raw-assets/baseimages-packages.Dockerfile
package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _dockerfile = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x90\x4f\x8f\xda\x30\x14\xc4\xef\xfe\x14\x4f\x6a\x8e\xe0\xd0\x2b\x12\x87\x50\x02\x44\x15\x09\x0a\xa5\x7f\x4e\xc8\x24\x8f\xc4\x52\x62\xbb\xf6\x83\x96\xb5\xfc\xdd\x57\x0b\x1b\x2d\xd2\xee\xde\xc6\xbf\x19\x6b\xc6\x4e\xca\x15\xcc\x93\x5d\x7a\xc8\x36\xc9\x2a\x65\xcb\xb2\xd8\x40\xf4\x00\xd8\x4b\x60\x9d\xec\x0e\xab\xed\x9e\xa5\xf9\xcf\x41\xcf\xa2\x01\xde\xe8\x22\x9d\x67\x49\x7e\x58\x96\x45\xfe\x23\xcd\x17\x33\xa5\x95\x54\x84\x56\x54\x24\x2f\xc8\xbc\xb7\x42\x35\x08\x91\x1c\x41\x24\x95\x23\xd1\x75\x30\x9d\x01\xcf\xee\x3a\x04\x56\xee\x73\xf0\x7e\x30\x43\xf0\x1e\x55\x1d\x02\xf3\x5e\x9e\x80\xaf\x85\xdb\x5e\xa9\xd5\xaa\xc4\xbf\x67\x69\xb1\x47\x45\x2e\x04\xf6\xad\xd8\xfe\x01\xef\xf9\x47\x26\xc4\xd4\x9b\xd8\x3e\x30\x4e\xff\xe9\x56\x64\xa4\x81\x61\xc6\xd8\x7e\x12\x1c\x16\x7c\x81\x65\xf6\x7b\x93\x4e\x81\xb0\x37\xda\x0a\x7b\x1d\xc1\x59\x91\xec\x80\x5a\xe9\x40\x3a\xd0\x0a\xcc\xd5\x48\xd0\x16\xfe\x21\x9c\xa4\xaa\x41\xc0\x11\x89\xd0\xbe\x5d\x02\x67\xf4\xfb\xf6\x96\xc8\xb8\x69\x1c\x3b\xd2\x56\x34\xc8\x1b\xad\x9b\x0e\x85\x91\x8e\x57\xba\x8f\x8f\x27\x69\x5d\x3b\xae\xf1\x12\x5b\x34\x9d\xac\x04\xe1\xd8\xdc\x1e\xfb\x00\x26\x7c\xc2\xbf\x72\x12\x96\x37\x4f\xec\xfe\x27\x1c\xe2\x4a\xd7\xc8\x7e\x15\xe5\xf7\x45\x56\xbe\x9e\x9e\x03\x00\x00\xff\xff\xcb\x09\xc0\x19\xee\x01\x00\x00")

func dockerfileBytes() ([]byte, error) {
	return bindataRead(
		_dockerfile,
		"Dockerfile",
	)
}

func dockerfile() (*asset, error) {
	bytes, err := dockerfileBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "Dockerfile", size: 494, mode: os.FileMode(420), modTime: time.Unix(1594845012, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _baseimagesBaseDockerfile = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x53\x4d\x6b\xdb\x40\x10\xbd\xeb\x57\x2c\x26\x98\xa6\xb0\x5a\xda\xd2\x8b\x21\x87\x84\x38\x34\x87\x3a\xc1\xa4\x39\x15\xca\xae\x34\x96\xa7\x99\xfd\xf0\xee\xc8\x89\xed\xfa\xbf\x17\xc9\xae\x2b\x4b\xd1\x6d\xdf\x9b\xf7\x66\xe6\x31\xba\x9b\x3f\x7c\x17\xbb\x5d\x7e\xa3\x13\xdc\x5b\x5d\xc1\x7e\x9f\x65\xd3\xd9\xb3\xb8\x9d\xde\xdc\x5f\xcf\x7e\xdd\xcd\x1f\x66\x4f\xd3\xd9\xed\x95\xf3\x0e\x1d\x43\xd4\x05\xe3\x1a\xb2\x6c\xfe\x63\x26\x74\x60\x59\x01\x8b\x3a\x94\x9a\x41\xc8\x95\x18\x8f\x4f\x20\xba\xc4\x9a\x48\xc8\xd5\x46\x48\xe9\xbc\x3c\x02\x32\x42\xe1\xad\x05\x57\x26\xf1\x33\x13\xc7\xcf\xea\x17\xe8\x3c\x4d\x8d\x54\x4a\x48\x09\x1c\xa3\xa6\x0e\x43\x68\x52\x22\x59\xc2\xba\x03\x6e\x09\xcd\xa7\xaa\x07\x12\x1a\xb3\xfd\x3c\x04\x23\xe8\x92\xd0\xc1\x90\x49\x2b\x42\x86\x2f\x3d\xe2\xb5\xd9\xe6\xff\xb3\xa8\xe3\xd9\x3c\xb4\xb6\xe7\x2e\xae\xa8\x63\x82\xf4\x75\xe8\x7f\x64\x5e\xfb\xd4\xdb\x56\xd6\x8c\xd4\xcd\x83\x5f\x86\xf2\xc5\x02\x87\x20\x6d\xad\xee\xa1\x61\xc3\x4b\xef\xa4\x0f\xe0\x52\xea\xce\x5a\xe1\xd9\x22\x5a\x16\x10\x19\x17\x58\x68\x86\x6e\xf3\x92\xe3\x5b\xe7\x39\x1e\x8b\x68\x85\x8c\x0b\xa1\xd6\x3a\x2a\x42\xa3\x74\x60\x45\x98\x38\xa9\x8f\x87\x53\x68\x53\x59\x32\x87\x34\x51\x2a\x6c\xc0\xad\xf3\x58\x3b\xf1\x47\x18\x9d\x96\xed\x41\x3d\x5e\x3f\x7d\xbb\x1a\xa9\xe8\x3d\xab\xbc\xad\x50\x69\x89\x36\x4d\xce\x20\x83\x6e\x72\xd1\x94\x8e\x5a\xdb\x66\xe2\x82\xbc\x83\x93\x77\x85\xbc\xac\x4d\x5e\x78\xab\xac\xb7\x5e\x92\x36\x87\x7e\xa7\xfb\xa2\x66\x1b\xce\x1b\xe9\xe8\xe2\x43\xcb\x89\xa6\xc5\xe5\x48\x05\xaa\x2b\x74\xe9\x5d\x41\xb3\x66\x37\xc3\x46\xd6\xab\xd8\xed\xf2\xc7\x36\xdb\x67\x88\x09\xbd\xdb\xef\xdb\x21\x0f\xb5\x15\x79\xa3\x49\xfc\xeb\xd8\x93\x4a\x19\x22\xba\xf7\x2c\x2e\x0f\x01\x06\x0c\xa7\x5f\xe6\x77\x1d\x36\x0c\x31\xfb\x1b\x00\x00\xff\xff\xaf\x95\x41\x01\x9d\x03\x00\x00")

func baseimagesBaseDockerfileBytes() ([]byte, error) {
	return bindataRead(
		_baseimagesBaseDockerfile,
		"baseimages-base.Dockerfile",
	)
}

func baseimagesBaseDockerfile() (*asset, error) {
	bytes, err := baseimagesBaseDockerfileBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "baseimages-base.Dockerfile", size: 925, mode: os.FileMode(420), modTime: time.Unix(1594831248, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _baseimagesPackagesDockerfile = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x0b\xf2\xf7\x55\xa8\xae\xd6\x73\x4a\x2c\x4e\xf5\xcc\x4d\x4c\x4f\xad\xad\xe5\xe2\x0a\x0a\xf5\x53\x28\xc8\x2c\x50\xc8\xcc\x2b\x2e\x49\xcc\xc9\x01\xc9\x07\x54\x96\x64\xe4\xe7\x05\x24\x26\x67\x27\xa6\xa7\x16\xd7\xd6\x72\x01\x02\x00\x00\xff\xff\xb5\xd4\xe0\x91\x39\x00\x00\x00")

func baseimagesPackagesDockerfileBytes() ([]byte, error) {
	return bindataRead(
		_baseimagesPackagesDockerfile,
		"baseimages-packages.Dockerfile",
	)
}

func baseimagesPackagesDockerfile() (*asset, error) {
	bytes, err := baseimagesPackagesDockerfileBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "baseimages-packages.Dockerfile", size: 57, mode: os.FileMode(420), modTime: time.Unix(1594831264, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"Dockerfile":                     dockerfile,
	"baseimages-base.Dockerfile":     baseimagesBaseDockerfile,
	"baseimages-packages.Dockerfile": baseimagesPackagesDockerfile,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"Dockerfile":                     &bintree{dockerfile, map[string]*bintree{}},
	"baseimages-base.Dockerfile":     &bintree{baseimagesBaseDockerfile, map[string]*bintree{}},
	"baseimages-packages.Dockerfile": &bintree{baseimagesPackagesDockerfile, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
