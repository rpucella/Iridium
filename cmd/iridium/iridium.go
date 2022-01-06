package main

import (
	"fmt"
	"os"
)

const SRC_JSON = "game.json"
const SRC_HTML = "game.html"
const SRC_NOTES = "notes.txt"
const SRC_PASSAGES = "passages"
const SRC_ASSETS = "assets"

const GAME_DIST = "dist"
const GAME_HTML = "game.html"
const GAME_ASSETS = "assets"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("USAGE: iridium <command> [<args>]")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println(" init <folder>")
		fmt.Println(" run [<folder>]")
		fmt.Println(" build [<folder>]")
		fmt.Println(" dev [<folder>]")
		return
	}

	switch(args[0]) {
	case "init":
		if len(args) != 2 {
			stop("USAGE: iridium init <folder>")
		}
		initialize(args[1])
		
	case "build":
		if len(args) > 2 {
			stop("USAGE: iridium build [<folder>]")
		}
		if len(args) == 1 {
			build(".")
		} else {
			build(args[1])
		}

	case "dev":
		if len(args) > 2 {
			stop("USAGE: iridium dev [<folder>]")
		}
		if len(args) == 1 {
			devCommand(".")
		} else {
			devCommand(args[1])
		}

	case "run":
		if len(args) > 2 {
			stop("USAGE: iridium run [<folder>]")
		}
		if len(args) == 1 {
			run(".")
		} else {
			run(args[1])
		}
	default:
		stop(fmt.Sprintf("Unknown command: %s", args[0]))
	}
}

// for most errors, don't try to recover, just stop

func stop(message string) {
	fmt.Println(message)
	os.Exit(1)
}

