package main

import (
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly"
	"golang.org/x/net/context"

	//"github.com/gocolly/colly"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// WriteCounter counts the number of bytes written to it. By implementing the Write method,
// it is of the io.Writer interface and we can pass this into io.TeeReader()
// Every write to this writer, will print the progress of the file write.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress prints the progress of a file write
func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 50))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: download url directory")
		os.Exit(1)
	}
	fmt.Println("Download Started")

	url := os.Args[1]
	dir := os.Args[2]
	host := getHostName(url)

	// Create folder if it not exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
	}

	c := colly.NewCollector()

	var links []string

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		//fmt.Println("image link: ", link)

		if strings.Contains(link, "/photo/") {
			links = append(links, link)
		}

		//err := DownloadFile(url, dir)
		//if err != nil {
		//	panic(err)
		//}

	})

	//// Find and visit all links
	//c.OnHTML("img[src]", func(e *colly.HTMLElement) {
	//	url := e.Attr("src")
	//	fmt.Println("image src: ", url)
	//
	//	links = append(links, url)
	//
	//	//err := DownloadFile(url, dir)
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//
	//})

	c.Visit(url)

	for _, link := range links {

		//\/sfwalbum.com/photo/62081820

		fmt.Println(host + link)

		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		var img string
		var example string
		if err := chromedp.Run(ctx,
			chromedp.Navigate(host+link),
			//chromedp.WaitVisible(`body > footer`),
			chromedp.Click("#downloadPhoto", chromedp.NodeVisible),
			//chromedp.OuterHTML("img", &img),
			chromedp.Value("html", &example),
		); err != nil {
			panic(err)
		}

		fmt.Println(img)

		fmt.Println(example)

		//c.OnHTML("div[class]", func(e *colly.HTMLElement) {
		//	//src := e.Attr("data-magnify-src")
		//	src := e
		//
		//	fmt.Println("image src: ", src)
		//
		//	//err := DownloadFile(url, dir)
		//	//if err != nil {
		//	//	panic(err)
		//	//}
		//
		//})

		//c.Visit(host + link)

		break
	}

	//fmt.Printf("links: %v", links)

	fmt.Println("Grabbing completed!")
}

// DownloadFile will download a url and store it in local filepath.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
// We pass an io.TeeReader into Copy() to report progress on the download.
func DownloadFile(url string, dir string) error {
	fileName := getFileName(url)

	// Create the file with .tmp extension, so that we won't overwrite a
	// file until it's downloaded fully
	out, err := os.Create(dir + "/" + fileName + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our bytes counter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Println()

	// Rename the tmp file back to the original file
	err = os.Rename(dir+"/"+fileName+".tmp", dir+"/"+fileName)
	if err != nil {
		return err
	}

	return nil
}

// getFileName
func getFileName(fullUrlFile string) string {
	fileUrl, err := url.Parse(fullUrlFile)
	if err != nil {
		panic(err)
	}

	path := fileUrl.Path
	segments := strings.Split(path, "/")

	return segments[len(segments)-1]
}

// getHostName
func getHostName(fullUrlFile string) string {
	fileUrl, err := url.Parse(fullUrlFile)
	if err != nil {
		panic(err)
	}

	return fileUrl.Scheme + "://" + fileUrl.Host
}
