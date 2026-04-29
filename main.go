package main

import (
	"flag"
	"fmt"
	"os"

	"eagleusb/eagleusb/imdb"
	"eagleusb/eagleusb/lastfm"
	"eagleusb/eagleusb/util"
)

const usageText = `Usage: stats <command> [flags]

Commands:
  lastfm  Fetch Last.fm top albums and update README
  imdb    Fetch latest IMDB ratings and generate poster collage
  all     Run both lastfm and imdb and update README
  help    Show this help message

Flags:
  -verbose
        verbose output
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usageText)
		os.Exit(1)
	}

	verbose := false

	switch os.Args[1] {
	case "lastfm":
		fs := flag.NewFlagSet("lastfm", flag.ExitOnError)
		fs.BoolVar(&verbose, "verbose", false, "verbose output")
		fs.Parse(os.Args[2:])
		util.InitLogger(verbose)
		if err := lastfm.Run(verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "imdb":
		fs := flag.NewFlagSet("imdb", flag.ExitOnError)
		fs.BoolVar(&verbose, "verbose", false, "verbose output")
		fs.Parse(os.Args[2:])
		util.InitLogger(verbose)
		if err := imdb.Run(verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "all":
		fs := flag.NewFlagSet("all", flag.ExitOnError)
		fs.BoolVar(&verbose, "verbose", false, "verbose output")
		fs.Parse(os.Args[2:])
		util.InitLogger(verbose)
		if err := lastfm.Run(verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := imdb.Run(verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Print(usageText)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		fmt.Print(usageText)
		os.Exit(1)
	}
}
