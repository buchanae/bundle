package main

import (
  "log"
  "net/http"
)

func main() {
  log.SetFlags(0)
  fs := http.FileSystem(Bundle)
  http.Handle("/", http.FileServer(fs))
  log.Println("Listening on http://localhost:8000")
  log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
