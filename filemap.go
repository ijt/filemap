// Package filemap defines a type of map that is stored on the file system.
package filemap

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
)

// NotFound is a sentinel error signifying that an entry is not contained in the map.
type NotFound struct {
}

func (nf NotFound) Error() string {
	return "filemap.NotFound"
}

// Map associates keys with values, storing them in a directory on the filesystem.
type Map struct {
	mu  sync.Mutex
	dir string
}

// New creates a filemap.Map using the given directory dir for storage.
func New(dir string) *Map {
	return &Map{dir: dir}
}

// Set sets an entry in the map for a given key k and value v.
func (m *Map) Set(k string, v []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.path(k)
	if err := ioutil.WriteFile(p, v, 0644); err != nil {
		return errors.Wrapf(err, "writing %s as entry for %s in map", p, k)
	}
	return nil
}

// Get retrieves an entry in the map for key k.
func (m *Map) Get(k string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.path(k)
	if !fileExists(p) {
		return nil, NotFound{}
	}
	bs, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file %s for entry %s", p, k)
	}
	return bs, nil
}

// Has returns a boolean telling whether the map contains a given key k.
func (m *Map) Has(k string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.path(k)
	return fileExists(p)
}

// NumEntries returns the number of entries in the directory.
func (m *Map) NumEntries() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, err := os.Open(m.dir)
	if err != nil {
		return 0, errors.Wrap(err, "opening directory with map entries")
	}
	names, err := d.Readdirnames(0)
	if err != nil {
		return 0, errors.Wrap(err, "getting names of files in map directory")
	}
	return len(names), nil
}

// Del deletes the file corresponding to key k.
func (m *Map) Del(k string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.path(k)
	if !fileExists(p) {
		return NotFound{}
	}
	if err := os.Remove(p); err != nil {
		return errors.Wrapf(err, "removing file %s for key %s", p, k)
	}
	return nil
}

// Range calls the given function f on all entries (k, v) in the map.
func (m *Map) Range(f func(string, []byte) error) error {
	d, err := os.Open(m.dir)
	if err != nil {
		return errors.Wrap(err, "opening directory with map entries")
	}
	names, err := d.Readdirnames(0)
	if err != nil {
		return errors.Wrap(err, "getting names of files in map directory")
	}
	for _, name := range names {
		k, err := decodeBase64(name)
		if err != nil {
			return errors.Wrap(err, "decoding base-64 encoded key")
		}
		p := filepath.Join(m.dir, name)
		v, err := ioutil.ReadFile(p)
		if err != nil {
			return errors.Wrapf(err, "reading contents of file %s for key %s", p, k)
		}
		if err := f(k, v); err != nil {
			return errors.Wrap(err, "calling callback")
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (m *Map) path(k string) string {
	return filepath.Join(m.dir, encodeBase64(k))
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func decodeBase64(s string) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
