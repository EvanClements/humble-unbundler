package main

import (
	"log"
	"net/http"
	"os"

	// Import Strings module as 's'
	s "strings"

	// importing Colly for web scraping
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	//"github.com/velebak/colly-sqlite3-storage/colly/sqlite3"

	// Importing Grab (v3) for direct downllading
	// "github.com/cavaliergopher/grab/v3"

	// Importing FastJSON for parsing JSON responses
	"github.com/valyala/fastjson"
	// Importing SQlite module for data storagr
	//"zombiezen.com/go/sqlite"
	//"zombiezen.com/go/sqlite/sqlitex"
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
	// Init thebparser as p for reusability
	var p fastjson.Parser

	// Read config.json
	dat, err := os.ReadFile("config.json")
	if err != nil {
		ErrorLogger.Fatalf("cannot read file: %s", err)
	}

	// Parse file contents and grab cookie (if present)
	val, e := fastjson.ParseBytes(dat)
	if e != nil {
		ErrorLogger.Fatalf("cannot parse json: %s", e)
	}

	// Grab 'cookie', if no cookie, th
	sessCookie := val.Get("cookie").String()
	if sessCookie == "" {
		ErrorLogger.Fatalf("no cookie present in config.json")
	}

	// TODO: add SQLite3 init code that would do the following:
	// - [ ] Check for existing 'humble.db' file
	// - [ ] If not existing, then create it
	// - [ ] If existing, then  check for necessary Tables
	// - [ ] If not present, then create them
	// - [ ] If present, then complete init

	// URL to pass to Colly for initial scraping
	var uri string = "https://www.humblebundle.com/home/library"

	// Init Colly
	c := colly.NewCollector()

	// Clone the original collector for use with the api calls
	apiCall := c.Clone()

	// Create a request queue with 2 consumer threads
	q, _ := queue.New(4, &queue.InMemoryQueueStorage{MaxSize: 10000})

	c.OnRequest(func(r *colly.Request) {
		InfoLogger.Println("Visiting: ", r.URL)

		// Create a new cookie
		cookie := &http.Cookie{
			Name:  "_simpleauth_sess",
			Value: s.Trim(sessCookie, "\""),
			// need to find out how to make this not hard coded
		}
		// Add the cookie to the request
		r.Headers.Add("Cookie", cookie.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		ErrorLogger.Println("Something went wrong: ", err)
	})

	c.OnHTML("script#user-home-json-data", func(e *colly.HTMLElement) {
		v, err := p.Parse(e.Text)
		if err != nil {
			ErrorLogger.Fatalf("cannot parse json: %s", err)
		}

		// There is a lot of info here: do we want more than just 'gamekeys'?
		// Also, should we parse the info into a DB or JSON file?

		InfoLogger.Println("User JSON: ", v.Get())

		gk := v.GetArray("gamekeys")

		for i := 0; i <= len(gk)-1; i++ {
			var key string = s.Trim(gk[i].String(), "\"")
			var endpoint string = s.Join([]string{s.Join([]string{"https://www.humblebundle.com/api/v1/order", key}, "/"), "all_tpkds=true"}, "?")

			q.AddURL(endpoint)
		}

		// Next steps:
		// Create a 'User' struct (what will it contain?)
		// Thought: type User struct {
		//	SteamID string,
		//	OriginID string,
		//	UplayID string,
		//	GogID string,
		//	Gamekeys array,
		//}
		// Populate the User table/JSON file
	})

	apiCall.OnResponse(func(r *colly.Response) {
		v, err := fastjson.ParseBytes(r.Body)
		if err != nil {
			ErrorLogger.Fatalf("cannot parse json: %s", err)
		}

		// We have both direct downloads and keys encompassed in these vars
		ddl := v.Get().Get("subproducts")
		tpkd := v.Get().Get("tpkd_dict").Get("all_tpks")

		// TODO: hand pick necessary fields for bundle, item, link, and key
		// Bundle and Item can be created befor the key/link logic
		// then the key/link logic would encompass the for loop that would
		// iterate over each link or key and add to the appropriate table

		if len(ddl.GetArray()) != 0 {
			// this is where the SQLite code goes for links

			InfoLogger.Println("dl: ", ddl)
		} else {
			if len(tpkd.GetArray()) != 0 {
				// this is where the SQLite code goes for keys
				InfoLogger.Println("keys: ", tpkd)
			} else {
				InfoLogger.Println("No keys found: output: ", v.Get())
			}
		}

		// create a struct or map to to pass to an SQL query
	})

	// Go get the website
	c.Visit(uri)

	// Go through queue
	q.Run(apiCall)
}
