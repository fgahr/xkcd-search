# xkcd-search
A simple search tool for the popular XKCD webcomic. This is a solution to exercise 4.12 of the book *The Go Programming Language* by Donovan & Kernighan. As such it is not meant to be used by, well, you. But feel free to try anyway.

This program uses the `info.0.json` files for each comic, downloading them as required. They don't always closely resemble the graphical information of each comic, sometimes deviating considerably. The search is limited to strings (no regular expressions) and by default searches for occurrences of all arguments in any of the title, transcript, or alt (mouseover) fields. Search scope can be constricted with options, see below.

The program creates a local cache of comic information data under `~/.cache/xkcd-search/store.db`. When run normally, it fetches the latest comic, checks which are missing locally, and downloads them. For pure offline usage, use the `--local` switch (see below).

# Installation

If you have the go tool installed, simply run
```
go get github.com/fgahr/xkcd-search
```
If you don't, you're out of luck for now. Sorry :(

# Usage

Some switches are accepted, all other arguments are assumed to be search terms. Switches can appear in any position. Search is case-insensitive.

```
 xkcd-search --help
Usage: xkcd-search [options] keywords...
Available options:
-h,--help      Print this message
   --all       Search for comics containing all of the keywords (default)
   --any       Search for comics containing any of the keywords

   --local     Only search the local database, don't connect to the server
   --title     Only search for matches in a comic's title
   --alt-text  Only search for matches in a comic's alt-text


# EXAMPLES:

# Search for all of the given terms
 xkcd-search the game

# Made explicit
 xkcd-search --all you just won the game you\'re free

# One match is enough
 xkcd-search --any math physics chemistry

# Only query the local cache
 xkcd-search --any foo bar --local

# Only search in comic titles
 xkcd-search --title xkcd

# Only search in alt text (mouseover)
 xkcd-search --alt-text build environment grinning holding spatula
```
