package main

import (
	"flag"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
)

var visitedURLs = make(map[string]interface{})
var mx *sync.RWMutex = &sync.RWMutex{}
var basePath = "html"

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
	inURL := flag.String("url", "", "base url to start crawling")
	if inURL == nil {
		fmt.Println("bad url")
	}
	initialURL, err := url2.Parse(*inURL)
	if err != nil {
		fmt.Println("bad url")
		return
	}
	crawl(initialURL, initialURL.Host)
}

func crawl(url *url2.URL, host string) {
	if !validURL(url, host) {
		return
	}

	err := markAsVisited(url)
	if err != nil {
		return
	}

	urls := getPageURLs(url.String())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		downloadURL(url)
		wg.Done()
	}()

	for _, childURL := range urls {
		wg.Add(1)
		go func(childURL string) {
			parsedURL, err := url2.Parse(childURL)
			if err != nil {
				return
			}
			crawl(parsedURL, host)
			wg.Done()
		}(childURL)
	}
	wg.Wait()
}

func validURL(url *url2.URL, host string) bool {
	return url.Hostname() == host &&
		url.String() != ""
}

func getPageURLs(url string) []string {
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("%s bad url child\n", url)
		return nil
	}
	list := htmlquery.Find(doc, "//a/@href")
	urls := make([]string, len(list))
	for i, l := range list {
		urls[i] = htmlquery.SelectAttr(l, "href")
	}
	return urls
}

func downloadURL(url *url2.URL) {
	resp, err := http.Get(url.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fileName := path.Base(url.EscapedPath())
	fileDir := path.Dir(url.EscapedPath())
	filePath := fmt.Sprintf("%s/%s%s", basePath, url.Hostname(), fileDir)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(filePath, 0700) // Create your file
	}

	fileCompletePath := fmt.Sprintf("%s/%s", filePath, fileName)
	// handles cases when file already exists
	if _, err := os.Stat(fileCompletePath); err != nil {
		return
	}

	f, err := os.Create(fileCompletePath)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return
}

func markAsVisited(url *url2.URL) (err error) {
	mx.Lock()
	_, ok := visitedURLs[url.String()]
	if ok {
		err = fmt.Errorf("%s exists in map", url.String())
	}
	mx.Unlock()
	return err
}

func cleanup() {
	fmt.Println("stopping and cleaning")
}
