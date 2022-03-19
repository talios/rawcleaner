package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	deleteFiles     = kingpin.Flag("delete", "Actually delete the side car files.").Bool()
	verboseMode     = kingpin.Flag("verbose", "Run in verbose mode.").Short('v').Bool()
	veryVerboseMode = kingpin.Flag("very-verbose", "Run in VERY verbose mode.").Bool()
	runInline       = kingpin.Flag("inline", "Run deletion process as we scan.").Bool()
	includeHidden   = kingpin.Flag("hidden", "Include hidden files in scan.").Bool()
	basePath        = kingpin.Arg("path", "Base path to scan.").Required().String()

	savedSize int64
)

func main() {
	kingpin.Parse()

	if !strings.HasSuffix(*basePath, "/") {
		*basePath = *basePath + "/"
	}

	s := spinner.New(spinner.CharSets[9], 1024*time.Millisecond)
	s.Start()

	if *deleteFiles {
		log.Println("WARN: raw-cleaner will delete files")
	}

	fsys := os.DirFS(*basePath)

	log.Println("Looking for raw files in " + *basePath)

	allFound := []string{}

	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if isRawFile(p) {
			if *veryVerboseMode {
				log.Printf("Found %s%s\n", *basePath, p)
			}
			found := findSideCarFiles(s, *basePath, p)
			allFound = append(allFound, found...)
			s.Suffix = fmt.Sprintf("  : Found %d duplicates totalling %s", len(allFound), humanize.Bytes(uint64(savedSize)))
		}
		return nil
	}); err != nil {
		log.Printf("Walkdir returned error %v", err)
	}

	log.Printf("Found %d duplicate files.\n", len(allFound))

	if !*runInline {
		for _, found := range allFound {
			removeSideCar(found)
		}
	}

	if len(allFound) > 0 {
		if *deleteFiles {
			log.Printf("Saved %s bytes.\n", humanize.Bytes(uint64(savedSize)))
		} else {
			log.Printf("Run with -delete to save %s bytes.\n", humanize.Bytes(uint64(savedSize)))
		}
	}

	s.Stop()

}

func isRawFile(filename string) bool {
	match, _ := regexp.MatchString("\\.(raf|dmg)", strings.ToLower(filename))
	return match
}

func findSideCarFiles(spinner *spinner.Spinner, path string, filename string) []string {
	found := []string{}

	globPattern := fmt.Sprintf("%s/%s*", path, strings.TrimRight(filename, filepath.Ext(filename)))

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		fmt.Println(err)
	}
	for _, sideCarFilePath := range matches {
		isHidden := strings.HasPrefix(filepath.Base(sideCarFilePath), ".")
		if isHidden && *includeHidden {
			if strings.ToLower(filepath.Ext(sideCarFilePath)) == ".jpg" {
				found = append(found, sideCarFilePath)
				if *runInline {
					removeSideCar(sideCarFilePath)
				}
			}
		} else {
			log.Printf("Skipping hidden file %s\n", sideCarFilePath)
		}
	}
	return found
}

func removeSideCar(sideCarFilePath string) {
	if file, err := os.Stat(sideCarFilePath); err == nil {
		savedSize += file.Size()
		if *deleteFiles {
			log.Printf("Removing duplicate %s\n", sideCarFilePath)
			if err := os.Remove(sideCarFilePath); err != nil {
				log.Fatal(err)
			}
		} else {
			if *verboseMode || *veryVerboseMode {
				log.Printf("Found duplicate %s\n", sideCarFilePath)
			}
		}
	}
}
