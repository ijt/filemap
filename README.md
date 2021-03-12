# filemap
** Maps on the filesystem **

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/ijt/filemap/CI?style=flat-square)](https://github.com/ijt/filemap/actions?query=workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/ijt/filemap?style=flat-square)](https://goreportcard.com/report/github.com/ijt/filemap)

If you have a big map that's forcing you to run your Go program on an expensive instance with lots of RAM, filemap may be for you.

## Example:

```go
package main

import (
	"log"

	"github.com/ijt/filemap"
)

func main() {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	m, err := filemap.MakeMap(d)
	if err != nil {
		log.Fatal(err)
	}
	k := "hello"
	v := []byte("world")
	if err := m.Set(k, v); err != nil {
		log.Fatal(err)
	}
	if m.Has(k) {
		log.Fatal("map does not contain key after it was set")
	}
	v2, err := m.Get(k)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(v2, v) {
		log.Fatalf("Get returned %s, want %s", v2, v)
	}
}
```

## Testing
```sh
make test
```

## Benchmarking
```sh
make bench
```

