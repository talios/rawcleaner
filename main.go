package main

import (
	"fmt"
	"io/fs"

	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
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

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if !strings.HasSuffix(*basePath, "/") {
		*basePath = *basePath + "/"
	}

	pathLogger := log.WithFields(log.Fields{"path": *basePath})

	if *deleteFiles {
		pathLogger.Warn("raw-cleaner will delete files")
	}

	fsys := os.DirFS(*basePath)

	pathLogger.Info("Looking for raw files")

	allFound := []string{}

	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if isRawFile(p) {
			if *veryVerboseMode {
				pathLogger.WithFields(log.Fields{"file": p}).Info("Found raw file")
			}
			found := findSideCarFiles(*basePath, p)
			allFound = append(allFound, found...)
		}
		return nil
	}); err != nil {
		pathLogger.Fatalf("Walkdir returned error %v", err)
	}

	pathLogger.Infof("Found %d duplicate files.", len(allFound))

	if !*runInline {
		for _, found := range allFound {
			removeSideCar(found)
		}
	}

	if len(allFound) > 0 {
		foundLogger := pathLogger.WithFields(log.Fields{"bytes": humanize.Bytes(uint64(savedSize))})
		if *deleteFiles {
			foundLogger.Info("Saved bytes")
		} else {
			foundLogger.Warn("Run with -delete to save %s bytes.")
		}
	}

}

func isRawFile(filename string) bool {
	match, _ := regexp.MatchString("\\.(raf|dmg)$", strings.ToLower(filename))
	return match
}

func findSideCarFiles(path string, filename string) []string {
	found := []string{}

	globPattern := fmt.Sprintf("%s/%s*", path, strings.TrimRight(filename, filepath.Ext(filename)))

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		fmt.Println(err)
	}
	for _, sideCarFilePath := range matches {
		if strings.ToLower(filepath.Ext(sideCarFilePath)) == ".jpg" {
			isHidden := strings.HasPrefix(filepath.Base(sideCarFilePath), ".")
			if !isHidden || *includeHidden {
				found = append(found, sideCarFilePath)
				if *runInline {
					removeSideCar(sideCarFilePath)
				}

			} else {
				log.Warnf("Skipping hidden file %s", sideCarFilePath)
			}
		}
	}
	return found
}

func removeSideCar(sideCarFilePath string) {
	dupeLogger := log.WithFields(log.Fields{"file": sideCarFilePath})
	if file, err := os.Stat(sideCarFilePath); err == nil {
		savedSize += file.Size()
		if *deleteFiles {
			dupeLogger.Warn("Removing duplicate file")
			if err := os.Remove(sideCarFilePath); err != nil {
				log.Fatal(err)
			}
		} else {
			if *verboseMode || *veryVerboseMode {
				dupeLogger.Warn("Found duplicate file")
			}
		}
	}
}
