package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tchajed/goose/machine/filesys"
	"github.com/tchajed/mailboat/server"
)

var recover = flag.Bool("recover", false, "recover existing mailboxes rather than initialize")

func main() {
	flag.Parse()
	initialize := !*recover
	if initialize {
		fmt.Println("initializing mailboxes")
		os.RemoveAll("/tmp/mailboat")
		os.MkdirAll("/tmp/mailboat", 0744)
	}
	filesys.Fs = filesys.NewDirFs("/tmp/mailboat/")
	fmt.Println("listening on localhost:2110 (POP3) and localhost:2525 (SMTP)")
	server.Start(initialize)
}
