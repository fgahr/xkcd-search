// Package for interactions with the xkcd website
package xkcd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// The XKCD base URL.
const baseUrl = "https://xkcd.com"

// The URL suffix for comic strip info data (JSON).
const infoSuffix = "info.0.json"

// Struct containing info about an XKCD comic strip.
type ComicInfo struct {
	Num        int    // The strip number
	Day        string // Publication day
	Month      string // Publication month
	Year       string // Publication year
	News       string // News(?)
	Link       string // Link(?)
	SafeTitle  string `json:"safe_title"`
	Img        string // Image URL
	Alt        string // Mouseover text
	Title      string // Title
	Transcript string // The comic transcript
}

// The permanent link to the comic.
func (c ComicInfo) Url() string {
	return slashConnect(baseUrl, fmt.Sprint(c.Num))
}

// All of the comic's text fields.
func (c ComicInfo) allTextFields() []string {
	return []string{c.Title, c.Alt, c.Transcript}
}

// Whether the given term can is contained in any of the fields.
// Matching is case-insensitive.
func keyInFields(key string, fields ...string) bool {
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), strings.ToLower(key)) {
			return true
		}
	}
	return false
}

// Whether any of the given keys is present in any of the given fields.
func anyKeyInFields(keys, fields []string) bool {
	for _, key := range keys {
		if key == "" {
			continue
		}
		key = strings.ToLower(key)
		if keyInFields(key, fields...) {
			return true
		}
	}
	return false
}

// Whether all of the given keys are present in any of the given fields.
func allKeysInFields(keys, fields []string) bool {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if !keyInFields(key, fields...) {
			return false
		}
	}
	return true
}

// Wether a comic strip info contains any of the given keywords in a
// relevant text field.
func ContainsAnyKeyword(c ComicInfo, keys ...string) bool {
	return anyKeyInFields(keys, c.allTextFields())
}

// Whether a comic strip info contains all of the given keywords in a
// relevant text field.
func ContainsAllKeywords(c ComicInfo, keys ...string) bool {
	return allKeysInFields(keys, c.allTextFields())
}

// Whether the comic title contains any of the given keywords.
func TitleContainsAnyKeyword(c ComicInfo, keys ...string) bool {
	return anyKeyInFields(keys, []string{c.Title})
}

// Whether the comic title contains all of the given keywords.
func TitleContainsAllKeywords(c ComicInfo, keys ...string) bool {
	return allKeysInFields(keys, []string{c.Title})
}

// Obtain info for the given comic strip ID from the XKCD server.
// Non-positive IDs will result in the latest comic being fetched.
func FetchSingleComic(stripNumber int) (ComicInfo, error) {
	var infoUrl string
	if stripNumber < 1 {
		infoUrl = slashConnect(baseUrl, infoSuffix)
	} else {
		infoUrl = slashConnect(baseUrl, fmt.Sprint(stripNumber), infoSuffix)
	}
	resp, err := http.Get(infoUrl)
	if err != nil {
		panic("Failed to GET from URL: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ComicInfo{},
			fmt.Errorf("Unexpected response fetching comic number %d: %d",
				stripNumber, resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	var stripInfo ComicInfo
	err = decoder.Decode(&stripInfo)
	return stripInfo, err
}

// Fetch info of all comics from first to last, both included.
// Returns an empty slice if first > last.
func FetchComicRange(first int, last int) ([]ComicInfo, error) {
	if first > last {
		return nil, nil
	}
	comicChan := make(chan ComicInfo, 100)
	errorChan := make(chan error, 100)
	schedChan := make(chan struct{}, 100)
	defer close(comicChan)
	defer close(errorChan)
	defer close(schedChan)
	for k := 0; k < 100; k++ {
		schedChan <- struct{}{}
	}
	for i := first; i <= last; i++ {
		if i == 404 {
			continue
		}
		go func(idx int) {
			<-schedChan
			info, err := FetchSingleComic(idx)
			if err != nil {
				errorChan <- err
			} else {
				comicChan <- info
			}
			schedChan <- struct{}{}
		}(i)
	}

	var comics []ComicInfo
	var errors []error
	for i := first; i <= last; i++ {
		if i == 404 {
			continue
		}
		select {
		case comic := <-comicChan:
			comics = append(comics, comic)
		case err := <-errorChan:
			errors = append(errors, err)
		}
	}
	if errors != nil {
		// Returning the obtained comics could result in an incomplete data set
		// being stored locally which may never get repaired.
		return nil, errors[0]
	}

	sort.Slice(comics, func(i, j int) bool { return comics[i].Num < comics[j].Num })
	return comics, nil
}

// Connect the given components with slashes (to form a URL string).
func slashConnect(components ...string) string {
	return strings.Join(components, "/")
}
