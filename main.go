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
var savedSize int64

func init() {
	defaultPath := fmt.Sprintf("/Users/%s/Pictures", os.Getenv("USER"))

	flag.BoolVar(&deleteFiles, "delete", false, "actually delete the side car files")
	flag.BoolVar(&verboseMode, "v", false, "run in verbose mode")
	flag.BoolVar(&veryVerboseMode, "vv", false, "run in very verbose mode")
	flag.StringVar(&basePath, "path", defaultPath, "base path to check")
}

func main() {
	flag.Parse()

	s := spinner.New(spinner.CharSets[9], 1024*time.Millisecond)
	s.Start()

	if deleteFiles {
		log.Println("WARN: raw-cleaner will delete files")
	}

	fsys := os.DirFS(basePath)

	log.Println("Looking for Fuji raw files in " + basePath)

	var count int

	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if strings.ToLower(filepath.Ext(p)) == ".raf" {
			if veryVerboseMode {
				log.Printf("Found %s/%s/%s\n", basePath, p, d.Name())
			}
			findSideCarFiles(s, basePath, p)
			count++
		}
		return nil
	}); err != nil {
		log.Printf("Walkdir returned error %v", err)
	}

	s.Stop()

	log.Printf("Saved %s bytes.\n", humanize.Bytes(uint64(savedSize)))
	log.Printf("Found %d files.\n", count)

}

func findSideCarFiles(spinner *spinner.Spinner, path string, filename string) []string {
	found := []string{}

	globPattern := fmt.Sprintf("%s/%s*", path, strings.TrimRight(filename, filepath.Ext(filename)))

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		fmt.Println(err)
	}
	for _, sideCarFilePath := range matches {
		if strings.ToLower(filepath.Ext(sideCarFilePath)) == ".jpg" {
			found = append(found, sideCarFilePath)
			spinner.Suffix = fmt.Sprintf("  : Found %d duplicates", len(found))
			if verboseMode || veryVerboseMode {
				log.Printf("Found duplicate %s\n", sideCarFilePath)
			}
			removeSideCar(sideCarFilePath)
		}
	}
	return found
}

func removeSideCar(sideCarFilePath string) {
	if file, err := os.Stat(sideCarFilePath); err == nil {
		savedSize += file.Size()
		if deleteFiles {
			log.Printf("Removing duplicate %s\n", sideCarFilePath)
			if err := os.Remove(sideCarFilePath); err != nil {
				log.Fatal(err)
			}
		}
	}
}
