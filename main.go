package main

import (
	"log"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

type Context struct {
	Errors    int32
	Files     int32
	EmptyDirs int32
}

func readDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func main() {
	start := time.Now()
	log.SetOutput(os.Stdout)

	workers := 4
	rootDir := "/"

	ctx := new(Context)
	var wg sync.WaitGroup
	dirCh := make(chan string, 99999999)
	wg.Add(1)
	dirCh <- rootDir

	for i := 0; i < workers; i++ {
		go func() {
			for {
				dir, ok := <-dirCh
				if ! ok {
					return
				}
				infos, err := readDir(dir)
				if err != nil {
					atomic.AddInt32(&ctx.Errors, 1)
				} else if len(infos) < 1 {
					atomic.AddInt32(&ctx.EmptyDirs, 1)
				} else {
					for _, info := range infos {
						if info.IsDir() {
							wg.Add(1)
							dirCh <- path.Join(dir, info.Name())
						} else {
							atomic.AddInt32(&ctx.Files, 1)
						}
					}
				}
				wg.Done()
			}
		}()
	}

	wg.Wait()
	close(dirCh)
	elapsed := time.Since(start)

	log.Printf("Elapsed: %s\n", elapsed)
	log.Printf("Errors: %d\n", ctx.Errors)
	log.Printf("Files: %d\n", ctx.Files)
	log.Printf("Empty Dirs: %d\n", ctx.EmptyDirs)
}
