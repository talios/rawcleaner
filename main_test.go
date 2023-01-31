package main

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsRawFile(t *testing.T) {

	log.WithFields(log.Fields{"file": "test.raf"}).Info("Testing Raw Files")
	assert.Equal(t, true, isRawFile("test.raf"), ".raf files are raw")
	assert.Equal(t, true, isRawFile("test.RAF"), ".raf files are raw")
	assert.Equal(t, false, isRawFile("test.RAF.comask"), ".raf files are raw")
	assert.Equal(t, true, isRawFile("test.dng"), ".dng files are raw")
	assert.Equal(t, false, isRawFile("test.JPG"), ".jpg files are NOT raw")
}

func TestIsSideCarFile(t *testing.T) {
	assert.Equal(t, true, isSideCarFile("Puscifer - 1.raf", "Puscifer - 1.JPG"))
	assert.Equal(t, false, isSideCarFile("Puscifer - 1.raf", "Puscifer - 112 copy copy.jpg"))
	assert.Equal(t, true, isSideCarFile("2017.11 Glen Matlock - 001.RAF", "2017.11 Glen Matlock - 001.JPG"))
	assert.Equal(t, true, isSideCarFile("/tmp/2017.11 Glen Matlock - 001.RAF", "/tmp/2017.11 Glen Matlock - 001.JPG"))
}

func TestIsDuplicateRawFile(t *testing.T) {
	assert.Equal(t, true, isDuplicateRawFile("Puscifer - 1.raf", "Puscifer - 1-1.raf"))
	assert.Equal(t, true, isDuplicateRawFile("Puscifer - 1.raf", "Puscifer - 1-2.raf"))
	assert.Equal(t, true, isDuplicateRawFile("/tmp/Puscifer - 1.raf", "/tmp/Puscifer - 1-2.raf"))
	assert.Equal(t, false, isDuplicateRawFile("Puscifer - 1.raf", "Puscifer -11.raf"))
}

func TestIsHidden(t *testing.T) {
	isHidden := strings.HasPrefix("2019.01 - Ivo Pyper - DSC26693.RAF.comask", ".")
	assert.Equal(t, false, isHidden, "should not be hidden")

	isHidden = strings.HasPrefix(".DSC26693.RAF.comask", ".")
	assert.Equal(t, true, isHidden, "should not be hidden")

	isHidden = strings.HasPrefix("/Volumes/Raw Media/Photography/2018 General.cocatalog/Adjustments/LAM/2019/01/20/1/Video Test and Walk About - DSC24939.RAF.comask", ".")
	assert.Equal(t, false, isHidden, "should not be hidden")
}
