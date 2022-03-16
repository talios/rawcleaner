package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRawFile(t *testing.T) {

	assert.Equal(t, true, IsRawFile("test.raf"), ".raf files are raw")
	assert.Equal(t, true, IsRawFile("test.RAF"), ".raf files are raw")
	assert.Equal(t, true, IsRawFile("test.dmg"), ".dmg files are raw")
	assert.Equal(t, false, IsRawFile("test.JPG"), ".jpg files are NOT raw")

}
