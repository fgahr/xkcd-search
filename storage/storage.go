// Package for managing persistent local metadata of XKCD comics
package storage

import (
	"encoding/json"
	"github.com/fgahr/xkcd-search/xkcd"
	"os"
	"path/filepath"
	"sort"
)

// Get the database file.
// File mode is Read/Write, Append.
func getDbFile() (*os.File, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbDir := filepath.Join(homedir, ".cache", "xkcd-search")
	_, err = os.Stat(dbDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dbDir, 0700)
		if err != nil {
			return nil, err
		}
	}
	dbFileName := filepath.Join(dbDir, "store.db")
	return os.OpenFile(dbFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
}

// Load all comic information from storage, also returns the index of the
// latest one found.
func LoadAll() ([]xkcd.ComicInfo, int, error) {
	dbFile, err := getDbFile()
	if err != nil {
		return nil, 0, err
	}
	defer dbFile.Close()

	decoder := json.NewDecoder(dbFile)
	var stored []xkcd.ComicInfo
	highestId := 0
	for decoder.More() {
		var next xkcd.ComicInfo
		err = decoder.Decode(&next)
		if err != nil {
			return nil, 0, err
		}
		if next.Num > highestId {
			highestId = next.Num
		}
		stored = append(stored, next)
	}
	sort.Slice(stored, func(i, j int) bool { return stored[i].Num < stored[j].Num })

	return stored, highestId, nil
}

// Save the given comic infos to the database file.
func Store(infos []xkcd.ComicInfo) error {
	dbFile, err := getDbFile()
	if err != nil {
		return err
	}
	defer dbFile.Close()

	encoder := json.NewEncoder(dbFile)
	for _, comic := range infos {
		err = encoder.Encode(comic)
		if err != nil {
			return err
		}
	}
	return nil
}
