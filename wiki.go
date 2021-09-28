package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"errors"
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: body}, nil
}

// reg-ex to validate url path
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request)(string, error){
	m := validPath.FindStringSubmatch(r.URL.Path)

	// if it doesn't match, return 404
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}

	// the title is stored at m[2]
	return m[2], nil
}

func viewHandler(w http.ResponseWriter, r *http.Request){
	// check if page exists and fetch
	title, err := getTitle(w, r)

	if err != nil {
		return 
	}

	p, err := loadPage(title)

	// redirect if page doesn't exist
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request){
	// check if page exists and fetch it if so
	title, err := getTitle(w, r)

	if err != nil {
		return
	}


	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request){
	// check if page exists and fetch it if so
	title, err := getTitle(w, r)

	if err != nil {
		return
	}

	body := r.FormValue("body")

	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

// variable to hold/cache all pages rather than rendering them multiple times
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// fn to render to a given html template
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
	t, err := templates.ExecuteTemplate(w, tmpl + ".html", p)

	// handle err if not nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main(){
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
