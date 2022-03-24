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
	"golang.org/x/exp/slices"
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
		ForceQuote:    true,
	})

	if !strings.HasSuffix(*basePath, "/") {
		*basePath = *basePath + "/"
	}

	fsys := os.DirFS(*basePath)

	pathLogger := log.WithFields(log.Fields{"path": *basePath})

	if *deleteFiles {
		pathLogger.Warn("raw-cleaner will delete files")
	}

	pathLogger.Info("looking for raw files")

	allFound := []string{}
	allPaths := []string{}

	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {

		currPath := filepath.Dir(p)
		if !slices.Contains(allPaths, currPath) {
			pathLogger.WithFields(log.Fields{"subdir": currPath}).Info("checking subdir")
			allPaths = append(allPaths, currPath)
		}

		if isRawFile(p) {
			rawFile := fmt.Sprintf("%s/%s", *basePath, p)
			rawLogger := log.WithFields(log.Fields{"rawfile": rawFile})

			if *veryVerboseMode {
				rawLogger.Info("found raw file")
			}
			found := findSideCarFiles(rawLogger, *basePath, p)
			allFound = append(allFound, found...)
		}
		return nil
	}); err != nil {
		pathLogger.Fatalf("Walkdir returned error %v", err)
	}

	pathLogger.Infof("Found %d duplicate files.", len(allFound))

	if !*runInline {
		for _, found := range allFound {
			removeSideCar(pathLogger, found)
		}
	}

	if len(allFound) > 0 {
		foundLogger := pathLogger.WithFields(log.Fields{"bytes": humanize.Bytes(uint64(savedSize))})
		if *deleteFiles {
			foundLogger.Info("Saved bytes")
		} else {
			foundLogger.Warn("run with -delete to save %s bytes.")
		}
	}

}

func isRawFile(filename string) bool {
	match, _ := regexp.MatchString("\\.(raf|dmg)$", strings.ToLower(filename))
	return match
}

func findSideCarFiles(logger *log.Entry, path string, filename string) []string {
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
					removeSideCar(logger, sideCarFilePath)
				}

			} else {
				logger.Warnf("skipping hidden file %s", sideCarFilePath)
			}
		}
	}
	return found
}

func removeSideCar(logger *log.Entry, sideCarFilePath string) {
	dupeLogger := logger.WithFields(log.Fields{"sidecarfile": sideCarFilePath})
	if file, err := os.Stat(sideCarFilePath); err == nil {
		savedSize += file.Size()
		if *deleteFiles {
			if err := os.Remove(sideCarFilePath); err != nil {
				dupeLogger.Fatal(err)
			}
			dupeLogger.WithFields(log.Fields{"savedbytes": humanize.Bytes(uint64(savedSize))}).Warn("removed duplicate file")
		} else {
			if *verboseMode || *veryVerboseMode {
				dupeLogger.Warn("found duplicate file")
			}
		}
	}

}
