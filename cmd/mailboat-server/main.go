package main

import (
	"fmt"
	"os"

	"github.com/tchajed/goose/machine/filesys"
	"github.com/tchajed/mailboat/server"
)

func main() {
	os.RemoveAll("/tmp/mailboat")
	os.MkdirAll("/tmp/mailboat", 0744)
	filesys.Fs = filesys.NewDirFs("/tmp/mailboat/")
	fmt.Println("listening on localhost:2110 (POP3) and localhost:2525 (SMTP)")
	server.Start()
}
