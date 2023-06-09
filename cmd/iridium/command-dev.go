package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"bufio"
	"os"
	"os/exec"
	"time"
	"strings"
	"io"
	"io/ioutil"
	"encoding/json"
)

func devCommand(srcdir string) {
	log.Printf("Starting server at port 8080\n")

	assetsFileServer := http.FileServer(http.Dir(path.Join(srcdir, "/assets")))
	http.HandleFunc("/", rootHandler(srcdir))
	http.HandleFunc("/passage/", passageHandler(srcdir))
	http.HandleFunc("/raw/", rawHandler(srcdir))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsFileServer))

	go startBrowser(1)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// TODO: Abstract this away somehow.
func startBrowser(secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
	cmd := exec.Command("/Applications/Firefox.app/Contents/MacOS/firefox", "-private-window", "localhost:8080")
	_ = cmd.Run()
}

func passageHandler(srcdir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		passageName := strings.TrimPrefix(r.URL.Path, "/passage/")
		log.Println("Processing", passageName)
		passage, err := readPassage(path.Join(srcdir, SRC_PASSAGES), passageName)
		// need to check if the passage exists - if so, return 404!
		if err != nil {
			http.Error(w, "500 internal error.", http.StatusInternalServerError)
			return
		}
		rdr := strings.NewReader(passage)
		p := NewParser(rdr)
		psg, err := p.Parse()
		j, err := json.Marshal(*psg)
		if err != nil {
			http.Error(w, "500 internal error.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(j))
	}
}

func rawHandler(srcdir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		passageName := strings.TrimPrefix(r.URL.Path, "/raw/")
		if r.Method == "GET" { 
			log.Println("Getting", passageName)
			passage, err := readPassage(path.Join(srcdir, SRC_PASSAGES), passageName)
			// need to check if the passage exists - if so, return 404!
			if err != nil {
				http.Error(w, "500 internal error.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, passage)
		} else if r.Method == "PUT" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "500 internal error.", http.StatusInternalServerError)
				return
			}
			text := string(body)
			log.Println("Writing", passageName)
			err = writePassage(path.Join(srcdir, SRC_PASSAGES), passageName, text)
			if err != nil {
				http.Error(w, "500 internal error.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, "ok")
		} else {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
		}
	}
}

func notesHandler(srcdir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" { 
			log.Println("Getting notes")
			content, err := ioutil.ReadFile(path.Join(srcdir, SRC_NOTES))
			notes := ""
			if err == nil {
				// Silently swallow errors and return "" for notes instead.
				notes = string(content)
			}
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, notes)
		} else if r.Method == "PUT" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "500 internal error.", http.StatusInternalServerError)
				return
			}
			text := string(body)
			log.Println("Writing notes")
			err = ioutil.WriteFile(path.Join(srcdir, SRC_NOTES), []byte(text), 0644)
			if err != nil {
				http.Error(w, "500 internal error.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, "ok")
		} else {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
		}
	}
}

func rootHandler(srcdir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/notes" {
			notesHandler(srcdir)(w, r)
			return
		}
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
		file, err := os.Open(path.Join(srcdir, SRC_HTML))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			index := strings.Index(line, "</body>")
			if index >= 0 { 
				fmt.Fprintln(w, line[:index])
				fmt.Fprintln(w, "<script>")
				dumpCoreJS(w)
				fmt.Fprintln(w, "</script>")
				fmt.Fprintln(w, "<script>")
				dumpGameJS(w, path.Join(srcdir, SRC_JSON))
				fmt.Fprintln(w, "</script>")
				fmt.Fprintln(w, "<script>")
				dumpDevContent(w, path.Join(srcdir, SRC_PASSAGES))
				fmt.Fprintln(w, "</script>")
				fmt.Fprintln(w, "<script> document.querySelector('head > title').innerText = game.title; engine.run(game, content); </script>")
				fmt.Fprintln(w, line[index:])
			} else { 
				fmt.Fprintln(w, line)
			}
		}
	}
}

func dumpDevContent(w io.Writer, psgdir string) {
	fmt.Fprintln(w, `

function content() { 
  const cntnt = {}; 
  cntnt[game.init] = function(state) { 
     processPassage(game.init, false);
  };
  return cntnt;
}

const buttonStyle = 'margin-left: 16px; padding: calc(.5em - 1px) 1em; background-color: #00947e; color: #fff; border-radius: 2px; border-width: 1px; border-color: transparent; font-size: .8rem; cursor: pointer;';

const devMessageStyle = 'color: #00947e;'

function processPassage(psg, clear) { 
   if (clear) { 
     io.newp();
   }
   io.html('<div style="display: flex; flex-direction: row; align-items: center; margin-bottom: 16px;"><span style="' + devMessageStyle + '"><b>Passage: ' + psg + '</b></span> <button style="' + buttonStyle + '" onclick="edit(\'' + psg + '\', true)">Edit</button> <button style="' + buttonStyle + '" onclick="editNotes(\'' + psg + '\')">Notes</button></div>');
   fetch(encodeURI('/passage/' + psg))
     .then(response => { 
        if (response.status === 200) { 
          response.json()
            .then(json => processJSON(json, psg));
        } else { 
          io.html('<span style="color: red;"><b>No such passage</b></span>');
        }
    })
}

// put image name here when rendering so that if we hit edit we can access it
let imageName = null;

function joinText(items) { 
  return items.map(itemText).join(' ')
}

function itemText(item) { 
  ///console.log(item)
  switch(item.Kind) { 
    case 0: // WORD
      return item.Word
    case 1: // QUOTE
      return '<q>' + joinText(item.Content) + '</q>'
    default:
      return '??'
  }
}

function processJSON(json, psg) {
   imageName = null;
   for (let b of json.Blocks) { 
     switch(b.Kind) { 
       case 0:   // TEXT
         io.p(joinText(b.Content));
         break;
       case 1:   // IMAGE
         io.img(b.Image, b.Style);
         if (!imageName) { 
           imageName = b.Image;
         }
         break;
     }
   }
   let c = io.choices();
   for (let opt of json.Options) { 
     c = c.option(joinText(opt.Content), function() { processPassage(opt.Target, true) });
   }
   c.show();
}

function edit(psg, exists) { 
   io.newp();
   io.html('<div style="display: flex; flex-direction: row; align-items: center; margin-bottom: 16px;"><span style="' + devMessageStyle + '"><b>Passage: ' + psg + '</b></span> <button style="' + buttonStyle + '" onclick="save(\'' + psg + '\')">Save</button> <button style="' + buttonStyle + '" onclick="processPassage(\'' + psg + '\', true)">Cancel</button></div>');

   if (exists) { 
     fetch(encodeURI('/raw/' + psg))
       .then(response => { 
          if (response.status === 200) { 
            response.text()
              .then(text => createTextArea(text));
          } else { 
            createTextArea('');
          }
      })
   } else { 
      createTextArea('');
   }
}

function editNotes(psg) { 
   io.newp();
   io.html('<div style="display: flex; flex-direction: row; align-items: center; margin-bottom: 16px;"><span style="' + devMessageStyle + '"><b>NOTES</b></span> <button style="' + buttonStyle + '" onclick="saveNotes(\'' + psg + '\')">Save</button> <button style="' + buttonStyle + '" onclick="processPassage(\'' + psg + '\', true)">Cancel</button></div>');

   fetch(encodeURI('/notes'))
     .then(response => { 
        if (response.status === 200) { 
          response.text()
            .then(text => createTextArea(text, true));
        } else { 
          createTextArea('', true);
        }
      })
}

function saveNotes(psg) { 
   const text = document.querySelector('textarea').value;
   fetch(encodeURI('/notes'), {
       method: 'PUT',
       headers: { 
          'Content-Type': 'text/plain'
       },
       body: text
   }).then(response => processPassage(psg, true))
}

function createTextArea(init, noImage) {
   if (imageName && !noImage) { 
      io.html('<div style="display: flex; flex-direction: row; align-items: flex-start; width: 100%;"><textarea style="flex: 1 0; resize: vertical; width: 60%; height: 80vh; font-size: 70%;">' + init + '</textarea> <img src="' + imageName + '" style="width: 30%; margin-left: 16px;"></div>');
   } else { 
      io.html('<textarea style="resize: vertical; width: 100%; height: 80vh; font-size: 70%;">' + init + '</textarea>');
   }
}

function save(psg) { 
   const text = document.querySelector('textarea').value;
   fetch(encodeURI('/raw/' + psg), {
       method: 'PUT',
       headers: { 
          'Content-Type': 'text/plain'
       },
       body: text
   }).then(response => processPassage(psg, true))
}
`)
}
