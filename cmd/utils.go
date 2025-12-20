package cmd

import (
	"compress/gzip"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

// UnitInfo represents a teaching unit from Jupiter Web
type UnitInfo struct {
	Code string
	Name string
}

// httpGetWithCharset performs an HTTP GET request and returns a goquery Document
// with proper charset handling for USP Jupiter Web (iso-8859-1)
func httpGetWithCharset(url string, timeout time.Duration) (*goquery.Document, error) {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := charset.NewReader(resp.Body, "text/html; charset=iso-8859-1")
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// httpGetWithCharsetAndStatus performs an HTTP GET request and returns a goquery Document
// with proper charset handling and HTTP status code
func httpGetWithCharsetAndStatus(
	url string,
	timeout time.Duration,
) (*goquery.Document, int, error) {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	reader, err := charset.NewReader(resp.Body, "text/html; charset=iso-8859-1")
	if err != nil {
		return nil, resp.StatusCode, err
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return doc, resp.StatusCode, nil
}

// createGzipFile creates a gzipped version of the given file
func createGzipFile(originalPath string, data []byte) error {
	gzipPath := originalPath + ".gz"

	file, err := os.Create(gzipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	defer writer.Close()

	_, err = writer.Write(data)
	return err
}

// getTeachingUnits fetches all teaching units from Jupiter Web
func getTeachingUnits() ([]UnitInfo, error) {
	doc, err := httpGetWithCharset(
		"https://uspdigital.usp.br/jupiterweb/jupColegiadoLista?tipo=T",
		120*time.Second,
	)
	if err != nil {
		return nil, err
	}

	var units []UnitInfo
	doc.Find("a[href*='jupColegiadoMenu']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		name := s.Text()

		re := regexp.MustCompile(`codcg=(\d+)`)
		matches := re.FindStringSubmatch(href)
		if len(matches) > 1 {
			units = append(units, UnitInfo{
				Code: matches[1],
				Name: name,
			})
		}
	})

	return units, nil
}
