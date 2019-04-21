package main

import (
	"log"
	"sync"
	"testing"

	"github.com/tchajed/goose/machine/filesys"

	"github.com/tchajed/mailboat"
)

func TestBenchLoop(t *testing.T) {
	filesys.Fs = filesys.NewMemFs()
	filesys.Fs.Mkdir(mailboat.SpoolDir)
	for uid := uint64(0); uid < mailboat.NumUsers; uid++ {
		filesys.Fs.Mkdir(mailboat.GetUserDir(uid))
	}

	var wg sync.WaitGroup
	niter := 100
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			err := do_bench_loop(g, "Hello world.", niter, 1, 1)
			if err != nil {
				log.Fatal(err)
			}
		}(g)
	}
	wg.Wait()
}
