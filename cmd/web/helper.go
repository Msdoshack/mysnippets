package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"github.com/robfig/cron/v3"
)

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	
	// Extract the value for a given key from the query string. If no key exists this will return the empty string "". 
   s := qs.Get(key)
	// If no key exists (or the value is empty) then return the default value.
	if s == "" {
	return defaultValue
	}
	// Otherwise return the string.
	return s
}

 // The readInt() helper reads a string value from the query string and converts it to an 
// integer before returning. 
func (app *application) readInt(qs url.Values, key string, defaultValue int,) int {
    // Extract the value from the query string.
    s := qs.Get(key)
    // If no key exists (or the value is empty) then return the default value.
    if s == "" {
        return defaultValue
    }
    // Try to convert the value to an int. If this fails, add an error message to the 
    // validator instance and return the default value.
    i, err := strconv.Atoi(s)

    if err != nil {
        return defaultValue
    }
    // Otherwise, return the converted integer value.
    return  i
}

func (app *application) serverError(w http.ResponseWriter,  err error){
	trace := fmt.Sprintf("%s\n%s",err.Error(),debug.Stack())
	app.errorLog.Output(2,trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)

}

func (app *application) sendMailError(w http.ResponseWriter,  err error){
	trace := fmt.Sprintf("%s\n%s",err.Error(),debug.Stack())
	app.errorLog.Output(2,trace)

	var serverError Error

	serverError.Message = "server error"
	serverError.Details = "server can't reset your password at the moment, try again later"
	serverError.IsServerError = true

	data := &templateData{
		Error: serverError,
	}

	app.render(w, http.StatusInternalServerError, "error.html", data)

}

func (app *application) clientError (w http.ResponseWriter,status int){
	// http.Error(w, http.StatusText(status),status)

	var serverError Error

	serverError.Message = "Client Error"
	serverError.Details = "Something went wrong processing your request"
	

	data := &templateData{
		Error: serverError,
	}

	app.render(w, status, "error.html", data)
}


func (app *application) notFound(w http.ResponseWriter){
	
	// app.clientError(w,http.StatusNotFound)

	var serverError Error

	serverError.Message = "Client Error"
	serverError.Details = "The page you requested for doesnot exist"

	data := &templateData{
		Error: serverError,
	}

	app.render(w, http.StatusInternalServerError, "error.html", data)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData){
// Retrieve the appropriate template set from the cache based on the page
// name (like 'home.tmpl'). If no entry exists in the cache with the provided name,
	ts,ok := app.templateCache[page]

	if !ok {
		err := fmt.Errorf("the template %s does not exist" , page)
		app.serverError(w,err)
		return
	}

// initailize new buffer
	buf := new(bytes.Buffer)


// Write the template to the buffer, instead of straight to the
// http.ResponseWriter. If there's an error, call our serverError()helper
	err:= ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w,err)
		return
	}

	w.WriteHeader(status)

// // Execute the template set and write the response body. Again, if there
// // is any error we call the the serverError() helper.
// 	err := ts.ExecuteTemplate(w,"base",data)

// 	if err != nil {
// 		app.serverError(w, err)
// 	}
 buf.WriteTo(w)
}


func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
		Flash: app.sessionManager.PopString(r.Context(),"flash",),
		IsAuthenticated: app.isAuthenticated(r),
		AuthenticatedUserId: app.sessionManager.GetInt(r.Context(),"authenticatedUserID"),
		CSRFToken: nosurf.Token(r),
		
	}
}


// is the target destination that we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request,dst any)error{
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method will return an error with the type *form.InvalidDecoderError.We use errors.As() to check for this and raise a panic rather than returning the error
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError){
			panic(err)
		}
// for all other errors, we return them as normal.
		return err
	}

	return nil
}


func (app *application) isAuthenticated (r *http.Request) bool {
	isAuthenticated ,ok := r.Context().Value(isAuthenticatedContextKey).(bool)

	if !ok {
		return false
	}

	return isAuthenticated
}

func(app * application) hitEndPoint () {
	c := cron.New()

	// Add a job that runs every minute
	_, err := c.AddFunc("*/10 * * * *", func() {
		err := fetchData()
		if err != nil {
			app.errorLog.Println("Failed to fetch data:", err)
		} else {
			app.infoLog.Println("Successfully hit endpoint")
		}
	})

	if err != nil {
		app.errorLog.Println("Failed to add cron job:", err)
	}


	go func() {
		// Start the cron scheduler in a separate goroutine
		app.infoLog.Println("Starting scheduler")
		c.Start()
	}()
}

func fetchData() error {
	resp, err := http.Get("https://mysnippets.onrender.com")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response status is not OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch data, status code: %d", resp.StatusCode)
	}

	return nil
}