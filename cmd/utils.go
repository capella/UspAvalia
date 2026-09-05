package cmd

import (
	"net/http"
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
	doc, _, err := httpGetWithCharsetAndStatus(url, timeout)
	return doc, err
}

// httpGetWithCharsetAndStatus performs an HTTP GET request and returns a goquery Document
// with proper charset handling and HTTP status code. Every call is counted in
// the process-wide scrape stats.
func httpGetWithCharsetAndStatus(
	url string,
	timeout time.Duration,
) (doc *goquery.Document, statusCode int, err error) {
	defer func() { stats.recordHTTPResult(statusCode, err) }()

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

	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return doc, resp.StatusCode, nil
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

	if len(units) == 0 {
		// The page loaded but nothing matched: Jupiter changed its HTML.
		stats.parseErrors.Add(1)
	}
	stats.unitsFound.Store(int64(len(units)))

	return units, nil
}
