package models

import "errors"

var (
	
	ErrorNoRecord= errors.New("models: no matching record found")

	ErrInvalidCredentials = errors.New("models: invalid credentials")
	
	ErrDuplicateEmail = errors.New("models: duplicate email")
)


// <!-- <textarea type="text" name="content">{{.Form.Content}}</textarea> -->
// <!-- <pre><code>{{.Snippet.Content}}</code></pre> -->