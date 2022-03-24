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
	assert.Equal(t, true, isRawFile("test.dmg"), ".dmg files are raw")
	assert.Equal(t, false, isRawFile("test.JPG"), ".jpg files are NOT raw")

}

func TestIsHidden(t *testing.T) {

	isHidden := strings.HasPrefix("2019.01 - Ivo Pyper - DSC26693.RAF.comask", ".")
	assert.Equal(t, false, isHidden, "should not be hidden")

	isHidden = strings.HasPrefix("/Volumes/Raw Media/Photography/2018 General.cocatalog/Adjustments/LAM/2019/01/20/1/Video Test and Walk About - DSC24939.RAF.comask", ".")
	assert.Equal(t, false, isHidden, "should not be hidden")
}
