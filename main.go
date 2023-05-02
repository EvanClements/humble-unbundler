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

// Table 'Bundle'
// 'id','gamekey','name','name_pretty','created'

// Table 'Item'
// 'id','gamekey','name','name_pretty','platform','last_updated'

// Table 'Links'
// 'id','item_id','name','filesize','md5','sha1','url', 'downloaded'(timestamp),

// Table 'Keys'
// 'id','item_id','name','key_type','is_redeemed','key','steam_app_id'

// Create bundle first, then iterate over each item in bundle.
// While iterating, create link or key

type Link struct {
	item_id    int
	name       string
	filesize   int
	md5        string
	sha1       string
	url        string
	torrent    string
	platform   string
	downloaded string
}

type Key struct {
	item_id      int
	name         string
	key_type     string
	is_redeemed  bool
	key          string
	steam_app_id int
}

type Item struct {
	gamekey      string
	name         string
	name_pretty  string
	platform     string
	last_updated string
	is_link      bool
	link_id      int
	is_key       bool
	key_id       int
}

type Bundle struct {
	gamekey     string
	name        string
	name_pretty string
	created     string
}

type user struct {
	steam_id  string
	gog_id    string
	origin_id string
	cookie    string
}

var currentUser user

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
	// Init th bparsers as p and pb for reusability
	var p fastjson.Parser

	// Read config.json
	dat, err := os.ReadFile("config.json")
	if err != nil {
		ErrorLogger.Fatalf("cannot read file: %s", err)
	}

	// Parse file contents and grab cookie (if present)
	val, e := p.ParseBytes(dat)
	if e != nil {
		ErrorLogger.Fatalf("cannot parse json: %s", e)
	}

	// Grab 'cookie', if no cookie, th
	currentUser.cookie = val.Get("cookie").String()
	if currentUser.cookie == "" {
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
			Value: s.Trim(currentUser.cookie, "\""),
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

		gk := v.GetArray("gamekeys")

		for i := 0; i <= len(gk)-1; i++ {
			var key string = s.Trim(gk[i].String(), "\"")
			var endpoint string = s.Join([]string{s.Join([]string{"https://www.humblebundle.com/api/v1/order", key}, "/"), "all_tpkds=true"}, "?")

			q.AddURL(endpoint)
		}

		// Grab the GOG.com relevant field from the user JSON
		gog := v.Get("userOptions").Get("gog_account_id").String()

		// if it isn't empty, add the result to the User struct
		if gog != "" {
			currentUser.gog_id = gog
		}

		// Grab the EA Origin relevant field from the user JSON
		origin := v.Get("userOptions").Get("origin_username").String()

		// if it isn't empty, add the result to the User struct
		if origin != "" {
			currentUser.origin_id = origin
		}

		// Print the User Struct in the log.txt file tor testing
		InfoLogger.Println("User Struct: ", currentUser)

		// Populate the User table/JSON file
	})

	apiCall.OnResponse(func(r *colly.Response) {
		v, err := p.ParseBytes(r.Body)
		if err != nil {
			ErrorLogger.Fatalf("cannot parse json: %s", err)
		}

		// We have both direct downloads and keys encompassed in these vars
		ddl := v.Get().Get("subproducts")
		tpkd := v.Get().Get("tpkd_dict").Get("all_tpks")

		// Bundles have Items,
		// Items have Links and/or Keys

		// b = bundle{
		// gamekey:     v.GetString("gamekey"),
		// name:        v.GetString("machine_name"),
		// name_pretty: v.GetString(""),
		// created:     v.GetString(""),
		// }

		//currentBundle.gamekey = gk[i]
		// InfoLogger.Println("Response: ", v.Get())

		// TODO: hand pick necessary fields for bundle, item, link, and key
		// Bundle and Item can be created befor the key/link logic
		// then the key/link logic would encompass the for loop that would
		// iterate over each link or key and add to the appropriate table

		// Bundle
		b := Bundle{
			gamekey:     v.Get("gamekey").String(),
			name:        v.Get("product").Get("machine_name").String(),
			name_pretty: v.Get("product").Get("human_name").String(),
			created:     v.Get("created").String(),
		}

		InfoLogger.Println("Bundle: ", b)
		// this will be where the SQL code will go

		// Create FOR block in each IF statement to create items
		// When each item is created, create keys or links

		if len(ddl.GetArray()) != 0 {
			// this is where the SQLite code goes for links
			// links := []Link{}
			for i := 0; i <= len(ddl.GetArray()); i++ {
				// curr := ddl.GetArray()[i]
				// links = append(links, Link{
				// name: curr.GetArray("downloads")[1].GetString("machine_name").String(), // Same as name in item
				// gamekey: v.Get("gamekey").String(),
				// filesize int,
				// md5 string,
				// sha1 string,
				// url string,
				// torrent string,
				// platform string,
				// filetype string,
				// })

			}
			// InfoLogger.Println("dl: ", links)

		} else if len(tpkd.GetArray()) != 0 {
			// this is where the SQLite code goes for keys
			// InfoLogger.Println("keys: ", tpkd)

		} else {
			InfoLogger.Println("No keys or direct download links found - output: ", v.Get())
		}

		InfoLogger.Println("Items: ", items)

		// create a struct or map to to pass to an SQL query
	})

	// Go get the website
	c.Visit(uri)

	// Go through queue
	q.Run(apiCall)
}
