// Controls the functionality of xkcd-search
package main

import (
	"fmt"
	"github.com/freag/xkcd-search/storage"
	"github.com/freag/xkcd-search/xkcd"
	"log"
	"os"
)

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

// Print info about the given comic if it matches the predicate with respect
// to the search terms.
func printIf(comic xkcd.ComicInfo, pred func(xkcd.ComicInfo, ...string) bool, searchTerms ...string) {
	if pred(comic, searchTerms...) {
		fmt.Printf("%s: %s\n", comic.SafeTitle, comic.Url())
	}
}

// Search all XKCD comics for the given search terms.
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("No arguments given.")
	}

	predicate := xkcd.MatchesAllKeywords
	searchTerms := args
	fetchRemote := true
	for i, arg := range args {
		switch arg {
		case "--any":
			predicate = xkcd.MatchesAnyKeyword
			args[i] = ""
		case "--all":
			predicate = xkcd.MatchesAllKeywords
			args[i] = ""
		case "--local":
			fetchRemote = false
			args[i] = ""
		}
	}

	if len(searchTerms) == 0 {
		log.Fatal("No search terms given.")
	}

	comics := getComics(fetchRemote)

	for _, comic := range comics {
		printIf(comic, predicate, searchTerms...)
	}
}
