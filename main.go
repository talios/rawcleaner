package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
	"path/filepath"
)

var deleteFiles bool
var verboseMode bool
var veryVerboseMode bool
var basePath string
var savedSize int64 = 0

func init() {
	defaultPath := fmt.Sprintf("/Users/%s/Pictures", os.Getenv("USER"))

	flag.BoolVar(&deleteFiles, "delete", false, "actually delete the side car files")
	flag.BoolVar(&verboseMode, "v", false, "run in verbose mode")
	flag.BoolVar(&veryVerboseMode, "vv", false, "run in very verbose mode")
	flag.StringVar(&basePath, "path", defaultPath, "base path to check")
}

func main() {
	flag.Parse()

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	if deleteFiles {
		log.Println("WARN: raw-cleaner will delete files")
	}

	fsys := os.DirFS(basePath)

	log.Println("Looking for FUJI raw files in " + basePath)

	var count int
	extensions := []string{"JPG", "jpg"}

	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(strings.ToLower(filepath.Ext(p)), ".raf") {
			if veryVerboseMode {
				log.Printf("Found %s/%s/%s\n", basePath, p, d.Name())
			}
			for _, ext := range extensions {
				findSideCar(basePath, p, ext)
			}

			count++
		}
		return nil
	})

	s.Stop()

	log.Printf("Saved %s bytes.\n", humanize.Bytes(uint64(savedSize)))
	log.Printf("Found %d files.\n", count)

}

func findSideCar(path string, p string, requestedExt string) string {
	sideCarFile := p[0:len(p)-3] + requestedExt
	sideCarFilePath := path + "/" + sideCarFile
	if file, err := os.Stat(sideCarFilePath); err == nil {
		if verboseMode || veryVerboseMode {
			log.Printf("Found duplicate %s %s\n", requestedExt, sideCarFilePath)
		}
		savedSize += file.Size()
		if deleteFiles {
			log.Printf("Removing duplicate %s %s\n", requestedExt, sideCarFilePath)
			if err := os.Remove(sideCarFilePath); err != nil {
				log.Fatal(err)
			}
		}
	}
	return sideCarFilePath
}
