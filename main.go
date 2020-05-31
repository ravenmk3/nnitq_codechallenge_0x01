package main

import (
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type Context struct {
	Errors    int
	Files     int
	EmptyDirs int
}

func main() {
	start := time.Now()
	log.SetOutput(os.Stdout)

	workers := 4
	rootDir := "/"

	var wg sync.WaitGroup
	dirCh := make(chan string, 19999999)
	wg.Add(1)
	dirCh <- rootDir
	ctxs := []*Context{}

	for i := 0; i < workers; i++ {
		ctx := new(Context)
		ctxs = append(ctxs, ctx)
		go func(c *Context) {
			for {
				dir, ok := <-dirCh
				if ! ok {
					return
				}
				infos, err := readDir(dir)
				if err != nil {
					c.Errors++
				} else if len(infos) < 1 {
					c.EmptyDirs++
				} else {
					for _, info := range infos {
						if info.IsDir() {
							wg.Add(1)
							dirCh <- path.Join(dir, info.Name())
						} else {
							c.Files++
						}
					}
				}
				wg.Done()
			}
		}(ctx)
	}

	wg.Wait()
	close(dirCh)

	errors := 0
	files := 0
	emptyDirs := 0
	for _, ctx := range ctxs {
		errors += ctx.Errors
		files += ctx.Files
		emptyDirs += ctx.EmptyDirs
	}

	elapsed := time.Since(start)

	log.Printf("Elapsed: %s\n", elapsed)
	log.Printf("Errors: %d\n", errors)
	log.Printf("Files: %d\n", files)
	log.Printf("Empty Dirs: %d\n", emptyDirs)
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
