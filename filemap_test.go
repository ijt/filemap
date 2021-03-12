package filemap_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/ijt/filemap"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func TestNumEntries(t *testing.T) {
	fm, err := makeMap()
	if err != nil {
		t.Fatal(err)
	}
	n, err := fm.NumEntries()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("got %d entries, want 0 initially", n)
	}
	for i := 1; i <= 5; i++ {
		k := makeKey()
		v := makeVal()
		fm.Set(k, v)
		n, err := fm.NumEntries()
		if err != nil {
			t.Fatal(err)
		}
		if n != i {
			t.Errorf("got %d entries, want %d", n, i)
		}
	}
}

func TestSetGet(t *testing.T) {
	fm, err := makeMap()
	if err != nil {
		t.Fatal(err)
	}
	k := makeKey()
	v := makeVal()
	if fm.Has(k) {
		t.Fatal("map contains key before it was set")
	}
	if err := fm.Set(k, v); err != nil {
		t.Fatal(err)
	}
	if !fm.Has(k) {
		t.Fatal("map does not contain key after it was set")
	}
	v2, err := fm.Get(k)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v2, v) {
		t.Errorf("Get returned %v, want %v", v2, v)
	}
}

func TestDel(t *testing.T) {
	fm, err := makeMap()
	if err != nil {
		t.Fatal(err)
	}
	k := makeKey()
	v := makeVal()
	if err := fm.Set(k, v); err != nil {
		t.Fatal(err)
	}
	if err := fm.Del(k); err != nil {
		t.Fatal(err)
	}
	if fm.Has(k) {
		t.Errorf("map has key after it was deleted")
	}
	_, err = fm.Get(k)
	if !errors.Is(err, filemap.NotFound{}) {
		t.Errorf("Get returned %v after Del, want filemap.NotFound", err)
	}
}

func TestRange(t *testing.T) {

	t.Run("empty case", func(t *testing.T) {
		fm, err := makeMap()
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		err = fm.Range(func(k string, v []byte) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("counted %d iterations, want 0", n)
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		fm, err := makeMap()
		if err != nil {
			t.Fatal(err)
		}
		m := make(map[string][]byte)
		for i := 0; i < 5; i++ {
			k := makeKey()
			v := makeVal()
			fm.Set(k, v)
			m[k] = v
		}
		fm.Range(func(k string, v []byte) error {
			if !bytes.Equal(v, m[k]) {
				t.Errorf("fm[%s] is %s, want %s", k, v, m[k])
			}
			delete(m, k)
			return nil
		})
		if len(m) != 0 {
			t.Errorf("len(m) is %d, want 0; this means filemap.Range missed some entries", len(m))
		}
	})

	t.Run("error return", func(t *testing.T) {
		fm, err := makeMap()
		if err != nil {
			t.Fatal(err)
		}
		if err := fm.Set("key", []byte("val")); err != nil {
			t.Fatal(err)
		}
		e := errors.New("whoops")
		err = fm.Range(func(k string, v []byte) error {
			return e
		})
		if !errors.Is(err, e) {
			t.Errorf("Range call returned %v, want %v", err, e)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	fm, err := makeMap()
	if err != nil {
		t.Fatal(err)
	}

	var eg errgroup.Group
	for i := 0; i < 10; i++ {
		eg.Go(func() error {
			for j := 0; j < 100; j++ {
				k := makeKey()
				v := makeVal()
				if err := fm.Set(k, v); err != nil {
					return err
				}
				v2, err := fm.Get(k)
				if err != nil {
					return err
				}
				if !bytes.Equal(v2, v) {
					return fmt.Errorf("got %s for %s, want %s", v2, k, v)
				}
				if err := fm.Del(k); err != nil {
					return err
				}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkFilemap(b *testing.B) {
	fm, err := makeMap()
	if err != nil {
		b.Fatal(err)
	}
	k := makeKey()
	v := makeVal()
	for n := 0; n < b.N; n++ {
		if err := fm.Set(k, v); err != nil {
			b.Fatal(err)
		}
		if _, err := fm.Get(k); err != nil {
			b.Fatal(err)
		}
		if err := fm.Del(k); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNativeMap(b *testing.B) {
	m := make(map[string][]byte)
	k := makeKey()
	v := makeVal()
	for n := 0; n < b.N; n++ {
		m[k] = v
		_ = m[k]
		delete(m, k)
	}
}

func BenchmarkNoop(b *testing.B) {
	for n := 0; n < b.N; n++ {
	}
}

func makeKey() string {
	return fmt.Sprintf("key-%d", rand.Int())
}

func makeVal() []byte {
	return []byte(fmt.Sprintf("val-%d", rand.Int()))
}

func makeMap() (*filemap.Map, error) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary directory")
	}
	return filemap.New(d), nil
}
