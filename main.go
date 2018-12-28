package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
  "os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	log.SetFlags(0)

	var dev, keepModTime bool
	var pkg, prefix, tplPath string
	var err error
  varName := "Bundle"

	flag.StringVar(&pkg, "pkg", pkg, "Package name (required).")
  flag.StringVar(&varName, "var", varName, "Variable name where bundle is stored.")
  flag.BoolVar(&keepModTime, "modtime", keepModTime, "Keep the modtime intact.")
	flag.BoolVar(&dev, "dev", dev, "Dev mode.")
	flag.StringVar(&tplPath, "tpl", tplPath, "Path to template file.")
	flag.StringVar(&prefix, "prefix", prefix, "Strip the given prefix from all paths.")
	flag.Parse()
	args := flag.Args()

	if pkg == "" {
		log.Fatal("-pkg is required")
	}

	tpl := defaultTpl
	// Load template if passed via CLI flag.
	if tplPath != "" {
		tpl, err = template.ParseFiles(tplPath)
		if err != nil {
			log.Fatalf("parsing -tpl: %v\n", err)
		}
	}

	var files []file

	// Load all the files given as CLI args.
	for _, arg := range args {
		log.Println(arg)

    s, err := os.Stat(arg)
    if err != nil {
      log.Fatal(err)
    }

		var bytes []byte
		var abs string

    // In dev mode, the file isn't loaded as bytes,
    // it's referenced via an absolute file path,
    // which can be loaded at runtime.
		if dev {
			abs, err = filepath.Abs(arg)
		} else if !s.IsDir() {
			bytes, err = ioutil.ReadFile(arg)
		}
    if err != nil {
      log.Fatal(err)
    }

		arg = strings.TrimPrefix(arg, prefix)
    key := filepath.ToSlash(arg)

    // Arbitrary modtime
    modtime := int64(15459433680)
    if keepModTime {
      modtime = s.ModTime().Unix()
    }

		files = append(files, file{
      Key: key,
			Name:  filepath.Base(s.Name()),
			Abs:   abs,
      Mode: s.Mode(),
      ModTime: modtime,
      Size: s.Size(),
			Bytes: bytes,
		})
	}

	output := &bytes.Buffer{}
	err = tpl.Execute(output, map[string]interface{}{
		"Pkg":   pkg,
    "VarName": varName,
		"Dev":   dev,
		"Files": files,
	})
	if err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		fmt.Print(output.String())
		log.Fatal(err)
	}

	fmt.Print(string(formatted))
}

type file struct {
  Key string
	Name  string
	Abs   string
  Mode os.FileMode
  ModTime int64
  Size int64
	Bytes []byte
}

var defaultTpl = template.Must(template.New("default").Parse(`
package {{ .Pkg }}

// This file is generated. DO NOT EDIT.

import "os"
import "bytes"
import "errors"
import "time"
import "io"
import "path"
import "strings"

type bundleFileInfo struct {
  name string
  abs string
  size int64
  mode os.FileMode
  modTime time.Time
  bytes []byte
}
func (bf *bundleFileInfo) Name() string {
  return bf.name
}
func (bf *bundleFileInfo) Size() int64 {
  return bf.size
}
func (bf *bundleFileInfo) Mode() os.FileMode {
  return bf.mode
}
func (bf *bundleFileInfo) ModTime() time.Time {
  return bf.modTime
}
func (bf *bundleFileInfo) IsDir() bool {
  return bf.mode.IsDir()
}
func (bf *bundleFileInfo) Sys() interface{} {
  return nil
}

type bundleFile struct {
    *bytes.Reader
    f *bundleFileInfo
    // If this is a directory, keep a list of files.
    list []*bundleFileInfo
    // track position across multiple calls to Readdir.
    pos int
}

func (b *bundleFile) Readdir(n int) ([]os.FileInfo, error) {
  if !b.f.IsDir() {
    return nil, errors.New("can't call Readdir on a regular file.")
  }
  var ret []os.FileInfo
  j := 0
  for i := b.pos; i < len(b.list) && (n <= 0 || j < n); i++ {
    ret = append(ret, b.list[i])
    j++
  }
  if n > 0 && b.pos == len(b.list) - 1 {
    return ret, io.EOF
  }
  return ret, nil
}

func (b *bundleFile) Stat() (os.FileInfo, error) {
  return b.f, nil
}

func (b *bundleFile) Close() error {
  return nil
}

type bundleFS map[string]*bundleFileInfo

func (b bundleFS) Open(name string) (http.File, error) {
  f, ok := b[name]
  if !ok {
    // TODO look into https://golang.org/src/net/http/fs.go#L42
    return nil, os.ErrNotExist
  }
  return os.Open(f)
}

func (b bundleFS) Open(name string) (*bundleFile, error) {
  name = strings.TrimPrefix(path.Clean(name), "/")

  // Read root of bundle filesystem.
  if name == "" {
    var list []*bundleFileInfo
    for _, v := range b {
      list = append(list, v)
    }
    return &bundleFile{
      Reader: bytes.NewReader(nil),
      list: list,
      f: &bundleFileInfo{
        name: "/",
        size: 0,
        mode: os.ModeDir,
        modTime: time.Now(),
      },
    }, nil
  }

  f, ok := b[name]
  if !ok {
    return nil, os.ErrNotExist
  }

  if f.IsDir() {
    pattern := name + "/*"
    var list []*bundleFileInfo
    for k, f := range b {
      ok, _ := path.Match(pattern, k)
      if ok {
        list = append(list, f)
      }
    }
    return &bundleFile{
      Reader: bytes.NewReader(nil),
      list: list,
      f: f,
    }, nil
  }

  return &bundleFile{
    Reader: bytes.NewReader(f.bytes),
    f: f,
  }, nil
}

var {{ .VarName }} = bundleFS{
  {{ range .Files -}}
  {{ .Key | printf "%q" }}: {
    name:  {{ .Name | printf "%q" }},
    abs: {{ .Abs | printf "%q" }},
    size: {{ .Size }},
    mode: {{ .Mode | printf "%#v" }},
    modTime: time.Unix({{ .ModTime | printf "%d" }}, 0),
    bytes: {{ .Bytes | printf "%#v" }},
  },
  {{ end }}
}
`))
