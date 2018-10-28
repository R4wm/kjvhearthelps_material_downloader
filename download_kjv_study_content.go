package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var wg sync.WaitGroup

func main() {
	urlContent := "https://www.kjvhearthelps.com/lessons/"
	downloadDir := filepath.Join(os.Getenv("HOME"), "Downloads", "kjvhearthelps")
	// url := urlContent + "/I%20II%20Peter%20011610.pdf"

	// Make the directory to download files into.
	if err := os.MkdirAll(downloadDir, 0700); err != nil {
		log.Fatalf("Could not create dir %s\n", downloadDir)
	}

	resp, err := http.Get(urlContent)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Get a string value of the page
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	// Tokenize the page
	doc, err := html.Parse(strings.NewReader(string(bodyBytes)))

	// Parse through the content
	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {

			for _, a := range n.Attr {

				if a.Key == "href" {

					if strings.HasSuffix(a.Val, "pdf") {

						pathToDownload := filepath.Join(downloadDir, a.Val)

						wg.Add(1)
						go DownloadFile(pathToDownload, urlContent+a.Val)
						break
					} else {
						log.Printf("Skipping %s\n", a.Val)
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	wg.Wait()
}

// DownloadFile downloads the content
// Most of this functino is ripped from https://golangcode.com/download-a-file-with-progress/
func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the content
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	fmt.Printf("Done downloading %s to %s\n", url, filepath)

	wg.Done()
	return nil
}
