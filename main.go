package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// sync.Once type guarantees that the function we pass as an argument
	// will only be executed once, regardless of how many goroutines are calling ServeHTTP
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, nil)
}

func main() {
	r := newRoom()
	// root
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	// we are running the room in a separate Go routine
	// so that the chatting operations occur in the background,
	// allowing our main thread to run the web server
	go r.run()
	// start the web server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
