// Controls the functionality of xkcd-search
package main

import (
	"fmt"
	"github.com/freag/xkcd-search/storage"
	"github.com/freag/xkcd-search/xkcd"
	"log"
	"os"
	"path/filepath"
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
	// NOTE: The latest comic is fetched asynchronously. This is not a very
	// useful feature as it increases complexity for what is likely no
	// measurable performance gain. This was done mainly to satisfy my own
	// curiosity. Please forgive me.
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

// Enum type describing which fields to consider for matching.
type matchWhere int

const (
	allFields matchWhere = iota
	title
	altText
)

// Enum type describing how keywords are matched to the selected fields.
type matchHow int

const (
	matchAll matchHow = iota
	matchAny
)

// Struct containing all essential configuration options for this program.
type config struct {
	where       matchWhere
	how         matchHow
	fetchRemote bool
}

// Determine the predicate to select comics from the program's configuration.
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

// Print usage information and exit the program.
// Exits with return value 0, indicating success.
func usageAndExit() {
	execName, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Usage:  %s [options] keywords...\n", filepath.Base(execName))
	fmt.Print(`
Available options:
-h,--help   Print this message
   --all       Search for comics containing all of the keywords (default)
   --any       Search for comics containing any of the keywords

   --local     Only search to local database, don't connect to the server
   --title     Only search for matches in a comic's title
   --alt-text  Only search for matches in a comic's alt-text
`)
	os.Exit(0)
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
		case "-h":
			usageAndExit()
		case "--help":
			usageAndExit()
		default:
			// Found a search term, not a switch
			continue
		}
	}
	return conf
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

	conf := getConfig(args)
	comics := getComics(conf.fetchRemote)
	predicate := conf.getPredicate()

	for _, comic := range comics {
		printIf(comic, predicate, args...)
	}
}
