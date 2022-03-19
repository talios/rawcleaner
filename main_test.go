package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRawFile(t *testing.T) {

	assert.Equal(t, true, isRawFile("test.raf"), ".raf files are raw")
	assert.Equal(t, true, isRawFile("test.RAF"), ".raf files are raw")
	assert.Equal(t, true, isRawFile("test.dmg"), ".dmg files are raw")
	assert.Equal(t, false, isRawFile("test.JPG"), ".jpg files are NOT raw")

}
