package main

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Main(t *testing.T) {
	basePath = "test/html"
	// we should mock a server and a response, so we can check if content is the same too
	flag.Set("url", "https://acomer.pe")
	main()
	assert.NotEmpty(t, visitedURLs)
	assert.DirExists(t, basePath)
	assert.FileExists(t, basePath)
}
