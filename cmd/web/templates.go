package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"time"

	"github.com/msdoshack/mycodedairy/internal/models"
)

type Error struct{
	Message string
	Details string
	IsServerError bool
}

type templateData struct {
	CurrentYear 		int
	Snippet 			*models.Snippet
	Snippets 			[]*models.Snippet
	Languages			[]string
	SnippetCount  		int
	SnippetsByLang     	map[string][]*models.Snippet
	User 				*models.User
	Form 				any
	Error               Error
	Flash 				string
	IsAuthenticated 	bool
	AuthenticatedUserId int
	CSRFToken 			string
}


// CUSTOM TEMPLATE FUNCTION
func humanDate(t time.Time) string {
	return t.Format("02 jan 2006 at 15:04")
}

func max(currentPage, defaultP string) string {
    // Convert the currentPage string to an integer
    cPage, err := strconv.Atoi(currentPage)
    if err != nil {
        return "1" // Return a default value as a string
    }

    // Convert the defaultP string to an integer
    dP, err := strconv.Atoi(defaultP)
    if err != nil {
        return "1" // Return a default value as a string
    }

    // Determine the maximum value and convert it back to string
    if cPage > dP {
        return strconv.Itoa(cPage) // Convert int to string
    }

    return strconv.Itoa(dP) // Convert int to string
}

func prev(currentPage string) string {
	cp, err := strconv.Atoi(currentPage)

	if err != nil {
		return "1"
	}

	cp = cp - 1
	return strconv.Itoa(cp)
}

func next (currentPage string) string {
	cp,err := strconv.Atoi(currentPage)

	if err != nil {
		fmt.Println(err)
		return "1"
	}
	cp = cp + 1
	
	return strconv.Itoa(cp)
}

var functions = template.FuncMap{
	"humanDate":humanDate,
	"prev":prev,
	"next":next,
	"max":max,
}


func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

// Use the filepath.Glob() function to get a slice of all filepaths that
// match the pattern "./ui/html/pages/*.html". This will essentiallygives
// us a slice of all the filepaths for our application 'page' templates
// like: [ui/html/pages/home.tmpl ui/html/pages/view.html]
	pages, err := filepath.Glob("./ui/html/pages/*.html")

	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one.
	for _, page := range pages{
// Extract the file name (like 'home.html') from the full filepath
// and assign it to the name variable.

		name := filepath.Base(page)

// Parse the base template file into a template set.
// Registering new template functions before parsing
		ts,err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")

		if err != nil {
			return nil, err
		}

// Call ParseGlob() *on this template set* to add any partials.
		ts,err = ts.ParseGlob("./ui/html/partials/*.html")

		if err != nil {
			return nil,err
		}

// Call ParseFiles() *on this template set* to add the page template.

		ts,err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache,nil
}