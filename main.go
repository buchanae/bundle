package main

import (
  "bytes"
  "flag"
  "fmt"
  "go/format"
  "path/filepath"
  "io/ioutil"
  "log"
  "strings"
  "text/template"
)

//go-bindata -debug -o shaders/shaders.go -prefix shaders -pkg shaders shaders/*.vert shaders/*.frag shaders/*.glsl

func main() {
  log.SetFlags(0)

  var dev bool
  var pkg, prefix string

  flag.StringVar(&pkg, "pkg", pkg, "Package name (required).")
  flag.BoolVar(&dev, "dev", dev, "Dev mode.")
  flag.StringVar(&prefix, "prefix", prefix, "Prefix to strip.")
  flag.Parse()
  args := flag.Args()

  if pkg == "" {
    log.Fatal("-pkg is required")
  }

  var files []file

  for _, arg := range args {
    log.Println(arg)

    var bytes []byte
    var abs string
    var err error

    if dev {
      abs, err = filepath.Abs(arg)
      if err != nil {
        log.Fatal(err)
      }
    } else {
      bytes, err = ioutil.ReadFile(arg)
      if err != nil {
        log.Fatal(err)
      }
    }

    arg = strings.TrimPrefix(arg, prefix)

    files = append(files, file{
      Name: arg,
      Abs: abs,
      Bytes: bytes,
    })
  }

  output := &bytes.Buffer{}
  err := tpl.Execute(output, map[string]interface{}{
    "Pkg": pkg,
    "Dev": dev,
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
  Name string
  Abs string
  Bytes []byte
}

var tpl = template.Must(template.New("output").Parse(`
{{ $dev := .Dev }}

package {{ .Pkg }}

{{- if $dev }}
import "io/ioutil"
{{ end -}}

var Bundle = map[string]struct {
  Name string
  Abs string
  Bytes func() ([]byte, error)
}{
  {{ range .Files -}}
  {{ .Name | printf "%q" }}: {
    Name:  {{ .Name | printf "%q" }},
    Abs:   {{ .Abs | printf "%q" }},
    Bytes: func() ([]byte, error) {
      {{ if $dev }}
      return ioutil.ReadFile({{ .Abs | printf "%q" }})
      {{ else }}
      return {{ .Bytes | printf "%#v" }}, nil
      {{ end }}
    },
  },
  {{ end }}
}
`))
