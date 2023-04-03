package main

import (
	//"fmt"
	"log"
	"net/http"
	"os"
	//"encoding/json"

	// Import Strings module as 's'
	s "strings"

	// importing Colly
	"github.com/gocolly/colly/v2"

	// Importing Grab (v3)
	// "github.com/cavaliergopher/grab/v3"

	// Importing FastJSON for parsing JSON responses
	"github.com/valyala/fastjson"
	// Importing SQlite module
	//"zombiezen.com/go/sqlite"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Status
	// InfoLogger.Println("Starting the process")

	// Init thebparser as p for reusability
	var p fastjson.Parser

	// scraping logic...
	var uri string = "https://www.humblebundle.com/home/library"
	//InfoLogger.Println("Visiting ", uri)

	// Init Colly
	c := colly.NewCollector()
	//InfoLogger.Println(",Colly output: ", c)

	// Clone the original collector for use with the api calls
	// apiCall := c.Clone()

	c.OnRequest(func(r *colly.Request) {
		InfoLogger.Println("Visiting: ", r.URL)

		// Create a new cookie
		cookie := &http.Cookie{
			Name:  "_simpleauth_sess",
			Value: "eyJyZXBsX3ZhbHVlIjoiODI3OGFhZGJiNjVhNzIwMmFjZGU4NWUwNWExNTAxYmUiLCJ1c2VyX2lkIjo1ODIyNzAyOTUwNTQ3NDU2LCJpZCI6ImxxWDdZa3U5UkEiLCJhdXRoX3RpbWUiOjE2Nzk0OTk2Mzl9|1679499869|ea53c37488c9c499bed5bb523aac7eb52c282d37",
			// need to find out how to make this not hard coded
		}
		// Add the cookie to the request
		r.Headers.Add("Cookie", cookie.String())
	})

	// Clone the original collector for use w
	//apiCall := c.Clone()

	c.OnError(func(_ *colly.Response, err error) {
		ErrorLogger.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		InfoLogger.Println("Page visited: ", r.Request.URL)
	})

	c.OnHTML("script#user-home-json-data", func(e *colly.HTMLElement) {
		v, err := p.Parse(e.Text)
		if err != nil {
			ErrorLogger.Fatalf("cannot parse json: %s", err)
		}

		gk := v.GetArray("gamekeys")
		// For testing
		for i := 0; i <= len(gk)-1; i++ {
			//InfoLogger.Println("FastJSON Output: ", gk[i])
			var key string = s.Trim(gk[i].String(), "\"")
			var endpoint string = s.Join([]string{"https://www.humblebundle.com/api/v1/order", key}, "/")
			InfoLogger.Println("URL Requested: ", endpoint)
			// endpoint := s.Join(endpointPre, gk[i])
			//apiCall.OnResponse(func(q *colly.Response) {

			//rs, err := p.Parse(q.Body)
			//if err != nil || len(rs.Data) == 0 {
			//	stop = true
			//	return
			//}

			//InfoLogger.Println("OnResponse: ", q)
			//})

			//apiCall.OnHTML("html", func(z *colly.HTMLElement) {
			//	InfoLogger.Println("OnHTML: ", z)
			//})

			//apiCall.OnScraped(func(w *colly.Response) {
			//	InfoLogger.Println("OnScraped: ", w)
			//})

			//apiCall.Visit(endpoint)
		}
		// Do I process this here? or in c.OnScraped?
		// Next steps:
		// for loop > populate sql table with game keys

		// populate the user table
	})

	c.OnScraped(func(r *colly.Response) {
		InfoLogger.Println(r.Request.URL, " scraped!")

	})

	// Go get the website
	c.Visit(uri)

}
