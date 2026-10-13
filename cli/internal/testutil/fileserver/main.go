// fileserver serves a directory over HTTP for tests (CI's install.sh check): python3 -m http.server
// was flaky on the macOS runner.
//
//	go run ./internal/testutil/fileserver DIR ADDR
package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: fileserver DIR ADDR")
	}
	log.Fatal(http.ListenAndServe(os.Args[2], http.FileServer(http.Dir(os.Args[1]))))
}
