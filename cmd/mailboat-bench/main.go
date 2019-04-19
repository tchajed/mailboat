package main

import (
	"github.com/tchajed/goose/machine/filesys"
	"github.com/tchajed/mailboat"

	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
)

// Pick up is different from gomail, which only retrieves msgids
func do_bench_loop(tid int, msg string, niter int, nsmtpiter int, npopiter int) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for l := 0; l < niter; l++ {
		for i := 0; i < nsmtpiter; i++ {
			u := uint64(r.Int()) % mailboat.NumUsers
			mailboat.Deliver(u, []byte(msg))
		}
		for i := 0; i < npopiter; i++ {
			u := uint64(r.Int()) % mailboat.NumUsers
			msgs := mailboat.Pickup(u)
			for _, m := range msgs {
				mailboat.Delete(u, m.Id)
			}
			mailboat.Unlock(u)
		}
	}
	return nil
}

func main() {
	os.RemoveAll("/tmp/mailboat")
	os.MkdirAll("/tmp/mailboat", 0744)
	filesys.Fs = filesys.NewDirFs("/tmp/mailboat/")
	filesys.Fs.Mkdir(mailboat.SpoolDir)
	for uid := uint64(0); uid < mailboat.NumUsers; uid++ {
		filesys.Fs.Mkdir(mailboat.GetUserDir(uid))
	}

	mailboat.Open()

	nprocEnv := os.Getenv("GOMAIL_NPROC")
	if nprocEnv == "" {
		nprocEnv = "1"
	}
	nproc64, err := strconv.ParseInt(nprocEnv, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	niterEnv := os.Getenv("GOMAIL_NITER")
	if niterEnv == "" {
		niterEnv = "1000"
	}
	niter64, err := strconv.ParseInt(niterEnv, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

        f, err := os.Create("prof")
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()

	nproc := int(nproc64)
	niter := int(niter64)

	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(nproc)
	for g := 0; g < nproc; g++ {
		go func(g int) {
			defer wg.Done()
			err := do_bench_loop(g, "Hello world.", niter, 1, 1)
			if err != nil {
				log.Fatal(err)
			}
		}(g)
	}
	wg.Wait()

	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("%d threads, %d iter, %v elapsed\n", nproc, niter, elapsed)
}
