// Controls the functionality of xkcd-search
package main

import (
	"fmt"
	"github.com/freag/xkcd-search/storage"
	"github.com/freag/xkcd-search/xkcd"
	"log"
	"os"
)

// Fetch the info for the latest comic, and put in on the resultChan.
// In case of an error, put it to the errChan.
func fetchLatest(resultChan chan<- xkcd.ComicInfo, errChan chan<- error) {
	latest, err := xkcd.FetchSingleComic(0)
	if err != nil {
		errChan <- err
	} else {
		resultChan <- latest
	}
}

// Get all comics, local or remote.
func getComics(remote bool) []xkcd.ComicInfo {
	// NOTE: The latest comic is fetched concurrently. This is not a very useful
	// feature as it increases complexity for what is likely no measurable
	// performance gain. This was done mainly to satisfy my own curiosity.
	// Please forgive me.
	latest := make(chan xkcd.ComicInfo)
	errors := make(chan error)
	defer close(latest)
	defer close(errors)
	if remote {
		go fetchLatest(latest, errors)
	}

	comics, lastStored, err := storage.LoadAll()
	if err != nil {
		log.Fatal(err)
	}

	if remote {
		var newComics []xkcd.ComicInfo
		select {
		case latestInfo := <-latest:
			newComics, err = xkcd.FetchComicRange(lastStored+1, latestInfo.Num)
			if err != nil {
				log.Print("Failed to fetch all new comics.",
					"Proceeding with local data only. Error was:", err)
			}
		case err = <-errors:
			log.Print("Failed to determine number of latest comic.",
				"Proceeding with local data only. Error was:", err)
		}
		storage.Store(newComics)
		comics = append(comics, newComics...)
	}

	return comics
}

type matchWhere int

const (
	allFields matchWhere = iota
	title
	altText
)

type matchHow int

const (
	matchAll matchHow = iota
	matchAny
)

type config struct {
	where       matchWhere
	how         matchHow
	fetchRemote bool
}

func (c config) getPredicate() func(xkcd.ComicInfo, ...string) bool {
	// TODO Improve error messages.
	switch c.where {
	case allFields:
		switch c.how {
		case matchAll:
			return xkcd.ContainsAllKeywords
		case matchAny:
			return xkcd.ContainsAnyKeyword
		default:
			panic("Invalid configuration")
		}
	case title:
		switch c.how {
		case matchAll:
			return xkcd.TitleContainsAllKeywords
		case matchAny:
			return xkcd.TitleContainsAnyKeyword
		default:
			panic("Invalid configuration")
		}
	case altText:
		switch c.how {
		case matchAll:
			return xkcd.AltTextContainsAllKeywords
		case matchAny:
			return xkcd.AltTextContainsAnyKeyword
		default:
			panic("Invalid configuration")
		}
	default:
		panic("Invalid configuration")
	}
}

// Print info about the given comic if it matches the predicate with respect
// to the search terms.
func printIf(comic xkcd.ComicInfo, pred func(xkcd.ComicInfo, ...string) bool, searchTerms ...string) {
	if pred(comic, searchTerms...) {
		fmt.Printf("%s: %s\n", comic.SafeTitle, comic.Url())
	}
}

// Determine the proper program configuration from the command line arguments.
func getConfig(args []string) config {
	// Default config
	conf := config{
		where:       allFields,
		how:         matchAll,
		fetchRemote: true,
	}

	for i, arg := range args {
		switch arg {
		case "--any":
			conf.how = matchAny;
			args[i] = ""
		case "--all":
			conf.how = matchAll
			args[i] = ""
		case "--local":
			conf.fetchRemote = false
			args[i] = ""
		case "--title":
			conf.where = title
			args[i] = ""
		case "--alt-text":
			conf.where = altText
			args[i] = ""
		default:
			// Found a search term, not a switch
			continue
		}
	}
	return conf
}

// Search all XKCD comics for the given search terms.
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("No arguments given.")
	}

	conf := getConfig(args)
	comics := getComics(conf.fetchRemote)
	predicate := conf.getPredicate()

	for _, comic := range comics {
		printIf(comic, predicate, args...)
	}
}
