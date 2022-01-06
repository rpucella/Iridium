package main

import (
	"fmt"
	"os"
	"path"
)

func initialize(dir string) {
	fmt.Printf("Creating %s\n", dir)
	err := os.Mkdir(dir, 0755)
	if err != nil {
		stop(fmt.Sprint(err))
	}

	fmt.Printf("Creating %s/%s\n", dir, SRC_HTML)
	fileHTML, err := os.Create(path.Join(dir, SRC_HTML))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer fileHTML.Close()
	fmt.Fprintln(fileHTML, gameHTML)
	
	fmt.Printf("Creating %s/%s\n", dir, SRC_JSON)
	fileJSON, err := os.Create(path.Join(dir, SRC_JSON))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer fileJSON.Close()
	fmt.Fprintln(fileJSON, gameJSON)
	
	fmt.Printf("Creating %s/%s\n", dir, SRC_PASSAGES)
	err = os.Mkdir(path.Join(dir, SRC_PASSAGES), 0755)
	if err != nil {
		stop(fmt.Sprint(err))
	}
	
	fmt.Printf("Creating %s/%s\n", path.Join(dir, SRC_PASSAGES), "start.txt")
	filePassage, err := os.Create(path.Join(dir, SRC_PASSAGES, "start.txt"))
	if err != nil {
		stop(fmt.Sprint(err))
	}
	defer filePassage.Close()
	fmt.Fprintln(filePassage, gamePassage)
	
	fmt.Printf("Creating %s/%s\n", dir, SRC_ASSETS)
	err = os.Mkdir(path.Join(dir, SRC_ASSETS), 0755)
	if err != nil {
		stop(fmt.Sprint(err))
	}
}

const gameHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="icon" href="data:;base64,iVBORw0KGgo=">

    <title>Iridium Game</title>

    <link href="https://fonts.googleapis.com/css?family=Roboto:400,900" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css?family=Roboto+Slab:400,700" rel="stylesheet"> 
    <link href="https://fonts.googleapis.com/css?family=Libre+Baskerville" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css?family=Libre+Baskerville:700" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css?family=Droid+Serif:400,700" rel="stylesheet">

    <style>
      p.io-log {
          font-style: italic;
          color: tomato;
      }
      
      p.date {
          font-weight: bold;
          color: tomato;
      }

      p {
          text-align: justify;
      }

      .io-active-choice { 
          color: blue;
          text-decoration: none;
          cursor: pointer;
      }

      .io-selected-choice {
          font-style: italic;
      }

      .io-active-choice:hover { 
          text-decoration: underline;
      }

      h1.io-splash {
          font-size: 24px;
          font-family: "Roboto Slab", sans-serif;
          display: flex;
          justify-content: center;
          text-transform: uppercase;
          font-weight: bold;
      }
      
      h2.io-splash {
          font-size: 24px;
          font-family: "Roboto Slab", sans-serif;
          display: flex;
          justify-content: center;
      }
      
      h3.io-splash {
          font-size: 18px;
          font-family: "Roboto Slab", sans-serif;
          display: flex;
          justify-content: center;
          padding-bottom: 20px;
      }

      .io-title {
          font-size: 24px;
          font-family: "Roboto Slab", sans-serif;
      }

      body { 
          font-size: 20px;
          line-height: 1.3;
          font-family: "Georgia", "Libre Baskerville",  sans-serif;
          margin: 50px;
      }

      div.notes {
          padding: 20px;
          background: wheat;
      }
    </style>
  </head>
  
  <body>

    <div style="max-width: 900px; margin-left: auto; margin-right: auto;">
      <div id="play"></div>
    </div>
    
  </body>
  
</html>
`

const gameJSON = `{
    "title": "Title",
    "subtitle": "Subtitle",
    "author": "Author",
    "init": "start",
    "config": {
        "clear": true,
        "debug": true
    }
}
`

const gamePassage = `
The game start here.

{option next-screen}
  Go to next screen
{/option}
`


