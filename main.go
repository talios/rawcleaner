package main

import (
	"fmt"
	"io/fs"

	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-colorable"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"golang.org/x/exp/slices"
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

	initLogger()

	if !strings.HasSuffix(*basePath, "/") {
		*basePath = *basePath + "/"
	}

	fsys := os.DirFS(*basePath)

	guid := xid.New()

	pathLogger := log.WithFields(log.Fields{"path": *basePath, "xid": guid.String()})

	if *deleteFiles {
		pathLogger.Warn("raw-cleaner will delete files")
	}

	pathLogger.Info("looking for raw files")

	allFound := []string{}
	allPaths := []string{}

	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {

		currPath := filepath.Dir(p)
		if !slices.Contains(allPaths, currPath) {
			if *verboseMode {
				pathLogger.WithFields(log.Fields{
					"subdir":     currPath,
					"savedbytes": humanize.Bytes(uint64(savedSize)),
				}).Info("checking subdir")
			}
			allPaths = append(allPaths, currPath)
		}

		if isRawFile(p) {
			rawFile := fmt.Sprintf("%s/%s", *basePath, p)
			rawLogger := pathLogger.WithFields(log.Fields{"rawfile": rawFile})

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
		foundLogger := pathLogger.WithFields(log.Fields{"savedbytes": humanize.Bytes(uint64(savedSize))})
		if *deleteFiles {
			foundLogger.Info("Saved bytes")
		} else {
			foundLogger.Warn("run with -delete to remove files.")
		}
	}

}

func initLogger() {
	var logLevel = log.InfoLevel
	var logfile string

	currentUser, err := user.Current()

	if err == nil {
		logfile = currentUser.HomeDir + "/rawcleaner.log"
	} else {
		logfile = "rawcleaner.log"
	}

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   logfile,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter:  &log.JSONFormatter{},
	})

	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceQuote:    true,
	})
	log.AddHook(rotateFileHook)
}

func isRawFile(filename string) bool {
	match, _ := regexp.MatchString("\\.(raf|dng)$", strings.ToLower(filename))
	return match
}

func isSideCarFile(filename string, sideCarFilePath string) bool {
	return isSideCarFileForExt(filename, sideCarFilePath, ".jpg") || isSideCarFileForExt(filename, sideCarFilePath, ".jpeg")
}

func isSideCarFileForExt(filename string, sideCarFilePath string, ext string) bool {
	computedSideCar := strings.ToLower(strings.TrimRight(filepath.Base(filename), filepath.Ext(filename)) + ext)
	return strings.ToLower(filepath.Base(sideCarFilePath)) == computedSideCar
}

func isDuplicateRawFile(filename string, sideCarFilePath string) bool {
	rex := strings.ToLower(strings.TrimRight(filepath.Base(filename), filepath.Ext(filename)) + "-\\d+\\.raf")
	match, _ := regexp.MatchString(rex, strings.ToLower(sideCarFilePath))
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
		if isDuplicateRawFile(filename, sideCarFilePath) {
			logger.WithFields(log.Fields{"duplicaterawfile": sideCarFilePath}).Warn("duplicate raw file identified")
		}

		if isSideCarFile(filename, sideCarFilePath) {
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
