package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/msdoshack/mycodedairy/internal/models"
	"github.com/msdoshack/mycodedairy/internal/validator"
)

type addSnippetForm struct{
	Title 	string		`form:"title"`
	Content string		`form:"content"`
	Language string		`form:"language"`
	Visibility string	`form:"visibility"`
	UserId	   int		`form:"userId"`
	Expires int			`form:"expires"`

	validator.Validator `form:"-"`
}

type userSignupForm struct {
	FirstName 		string  	`form:"firstName"`
	LastName		string		`form:"lastName"`
	Email			string 		`form:"email"`
	Password		string  	`form:"password"`
	RetypePassword string		`form:"retypePassword"`
	validator.Validator			`form:"-"`
}

type userSignInForm struct {
	Email 		string 	`form:"email"`
	Password	string 	`form:"password"`
	validator.Validator	`form:"-"`	
}

type updateSnippetForm struct {
	Title string		`form:"title"`
	Content string		`form:"content"`
	Language string		`form:"language"`
	Visibility string	`form:"visibility"`
	UserId	int			`form:"userId"`
	Expires int			`form:"expires"`
	validator.Validator `form:"-"`
}

type updateProfileForm struct{
	FirstName	string		`form:"firstName"`
	LastName 	string		`form:"lastName"`
	Email		string		`form:"email"`
	validator.Validator		`form:"-"`
}

type changePasswordForm struct {
	Password string 	`form:"password"`
	validator.Validator
}

type resetPasswordForm struct  {
	UserID			int			`form:"userId"`
	Password       string		`form:"password"`
	RetypePassword string		`form:"reTypePassword"`
	validator.Validator
}

type forgetPasswordForm struct{
	UserID		int     `form:"userId"`
	Email 		string	`form:"email"`
	ResetCode 	string  `form:"resetCode"`
	validator.Validator
}

type queryForm struct {
	Title 		string	`json:"title"`
	Language	string	`json:"language"`
	Sort 		string	`json:"sort"`
	Page        string   `json:"page"`
}


func (app *application)home (w http.ResponseWriter, r *http.Request) {

	var filter struct {
        Title    string
      	Language  string
        Sort     string
		Limit	string
        Page     int
        PageSize int
    }

	qs := r.URL.Query()

	filter.Title = app.readString(qs,"title","")
	filter.Language = app.readString(qs,"language", "")
	filter.Sort = app.readString(qs,"sort", "")
	filter.Page = app.readInt(qs,"page", 1)

	page := strconv.Itoa(filter.Page)

	userId :=  app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	snippets,languages, err := app.snippetss.GetLatest(userId, filter.Page, filter.Title, filter.Language, filter.Sort)
	if err !=nil {
		app.serverError(w,err)
		return
	}


	data := app.newTemplateData(r)

	data.SnippetCount =len(snippets)
	
	data.Form = queryForm{
		Title: filter.Title,
		Language: filter.Language,
		Sort: filter.Sort,
		Page: page,
	}
	
	data.Snippets = snippets
	data.Languages = languages

	// app.render(w,http.StatusOK,"home.html", data)
	app.render(w,http.StatusOK,"home.html", data)
}


func (app *application)userSnippets (w http.ResponseWriter,r *http.Request){

	authenticatedUser := app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	userSnippets,err := app.snippetss.GetUserSnippets(authenticatedUser)
	
	if err != nil {
		app.serverError(w,err)
		return
	}
	data := app.newTemplateData(r)
	data.SnippetsByLang = userSnippets

	app.render(w,http.StatusOK,"usersnippet.html",data)
}

func (app *application)userSnippetsSearch (w http.ResponseWriter, r *http.Request) {

	var filter struct {
        Title    string
      	Language  string
        Sort     string
		Limit	string
        Page     int
        PageSize int
    }

	qs := r.URL.Query()

	filter.Title = app.readString(qs,"title","")
	filter.Language = app.readString(qs,"language", "")
	filter.Sort = app.readString(qs,"sort", "")
	filter.Page = app.readInt(qs,"page", 1)

	page := strconv.Itoa(filter.Page)

	userId :=  app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	snippets, languages, err := app.snippetss.GetLatest(userId, filter.Page, filter.Title, filter.Language, filter.Sort)
	if err !=nil {
		app.serverError(w,err)
		return
	}


	data := app.newTemplateData(r)

	data.SnippetCount =len(snippets)
	
	data.Form = queryForm{
		Title: filter.Title,
		Language: filter.Language,
		Sort: filter.Sort,
		Page: page,
	}
	
	data.Snippets = snippets
	data.Languages =languages

	// app.render(w,http.StatusOK,"home.html", data)
	app.render(w,http.StatusOK,"usersnippetssearch.html", data)
}


func (app *application) singleSnippets(w http.ResponseWriter, r *http.Request) {
	
	params := httprouter.ParamsFromContext(r.Context())
	
	id, err := strconv.Atoi(params.ByName("id"))

	if err != nil {
		app.notFound(w)
		return
	}

	snippet, err := app.snippetss.Get(id)

	if err != nil {
		if errors.Is(err,models.ErrorNoRecord){
			app.notFound(w)
		}else{
			app.serverError(w,err)
		}
	}

	// PopString() method to retrieve the value for the "flash"  key.
	// also deletes the key and value from the session data, so it acts like a one-time fetch. If there is no matching key in the session data this will return the empty string.

	// this logic was moved to the newTemplateData for automation
	// flash := app.sessionManager.PopString(r.Context(),"flash")

	data := app.newTemplateData(r)
	data.Snippet = snippet


	// flash message
	// this logic was moved to the newTemplateData for automation
	// data.Flash = flash


	
	app.render(w,http.StatusOK,"view.html",data)
}



func (app *application) addSnippetForm(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = addSnippetForm{
		Expires: 1,
	}

	app.render(w, http.StatusOK, "add.html", data)
}


func (app *application) addSnippet(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()

	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	// declared a new empty instance of the addSnippet struct
	var form addSnippetForm

	err= app.decodePostForm(r, &form)
	if err != nil{
		app.clientError(w,http.StatusBadRequest)
		return
	}

	// WITHOUT THE THIRD PARTY FORM VALIDATOR
	// expires, err := strconv.Atoi(r.PostForm.Get("expires")) 
	// if err != nil {
	// 	app.clientError(w, http.StatusBadRequest)
	// 	return
	// }
	// form := addSnippetForm {
	// 	Title : r.PostForm.Get("title"),
	// 	Content : r.PostForm.Get("content"),
	// 	Expires: expires,
	// }


	form.CheckField(validator.NotBlank(form.Title), "title", "this field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "Title cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "this field cannot be blank")
	form.CheckField(validator.NotBlank(form.Language), "language", "this field cannot be blank")
	form.CheckField(validator.NotBlank(form.Visibility), "visibility", "this field cannot be blank")
	form.CheckField(validator.PermittedVisibility(form.Visibility, "private" ,"public"), "visibility","this field must either be private or public")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w,http.StatusUnprocessableEntity,"add.html", data)
		return
	}


	userId :=  app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	id, err := app.snippetss.Insert(form.Title,form.Content,form.Language,form.Visibility, userId)

	if err != nil {
		app.serverError(w,err)
		return
	}

	// SETTING UP A FLASH MESSAGE IF CREATION OF NEW SNIPPET IS SUCCESSSFUL
	app.sessionManager.Put(r.Context(),"flash","Snippet successfully created")

	http.Redirect(w,r,fmt.Sprintf("/snippet/%d",id), http.StatusSeeOther)
}

func (app *application) updateSnippetForm(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	
	id, err := strconv.Atoi(params.ByName("id"))

	if err != nil {
		app.notFound(w)
		return
	}

	snippet, err := app.snippetss.Get(id)

	if err != nil {
		if errors.Is(err,models.ErrorNoRecord){
			app.notFound(w)
		}else{
			app.serverError(w,err)
		}
	}

	data := app.newTemplateData(r)

	data.Form = updateSnippetForm{
		Expires: 1,
	}


	data.Snippet = snippet

	app.render(w, http.StatusOK, "update.html", data)

}

func (app *application) updateSnippet(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	
	id, err := strconv.Atoi(params.ByName("id"))

	if err != nil {
		app.notFound(w)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}
	var form updateSnippetForm
	
	err = app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	


	userId :=  app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	if userId != form.UserId{
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title),"title","Title field is required")
	form.CheckField(validator.MaxChars(form.Title,100),"title","Title cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content","Content field is required")
	form.CheckField(validator.NotBlank(form.Visibility), "visibility", "this field cannot be blank")
	form.CheckField(validator.PermittedVisibility(form.Visibility, "private" ,"public"), "visibility","this field must either be private or public")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"add.html", data)
		return
	}



	err = app.snippetss.UpdateSnippet(form.Title, form.Content, form.Language, form.Visibility, id)
	
	if err != nil {
		app.serverError(w,err)
		return
	}

	// SETTING UP A FLASH MESSAGE IF CREATION OF NEW SNIPPET IS SUCCESSSFUL
	app.sessionManager.Put(r.Context(),"flash","Snippet successfully updated")

	http.Redirect(w,r,fmt.Sprintf("/snippet/%d",id), http.StatusSeeOther)
}

func(app *application) DeleteSnippet(w http.ResponseWriter,r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	
	id, err := strconv.Atoi(params.ByName("id"))

	if err != nil {
		app.notFound(w)
		return
	}
	
	err = app.snippetss.DelSnippet(id)
	if err != nil {
		app.serverError(w,err)
		return
	}

	http.Redirect(w,r,"/", http.StatusSeeOther)

}


func (app *application) signUpForm(w http.ResponseWriter, r *http.Request){
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w,http.StatusOK,"signup.html", data)
}

func (app *application) signUp(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.FirstName), "firstName", "Pls provide your first name")
	form.CheckField(validator.NotBlank(form.LastName), "lastName", "Pls provide your last name")

	form.CheckField(validator.NotBlank(form.Email),"email", "Pls provide your email address")

	form.CheckField(validator.Matches(form.Email,validator.EmailRX),"email","pls Provide a valid email address")

	form.CheckField(validator.NotBlank(form.Password), "password","please provide your password") 

	form.CheckField(validator.NotBlank(form.RetypePassword),"retypePassword","Pls enter your password again")

	form.CheckField(validator.IsPasswordMatch(form.Password, form.RetypePassword),"retypePassword","Password do no match")

	form.CheckField(validator.MinChars(form.Password, 8),"password", "password must be at least 8 characters long")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"signup.html",data)
		return
	}

	err = app.users.Insert(form.FirstName,form.LastName, form.Email, form.Password)
	if err != nil {
		if errors.Is(err,models.ErrDuplicateEmail){
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form= form
			app.render(w, http.StatusUnprocessableEntity,"signup.html", data)
			return
		} else {
			app.serverError(w,err)
			return
		}
		
	}

	app.sessionManager.Put(r.Context(),"flash","Your sign up was successful. please log in.")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) signInForm (w http.ResponseWriter, r *http.Request){
	
	data := app.newTemplateData(r)
	data.Form = userSignInForm{}

	app.render(w,http.StatusOK,"signin.html",data)

}

func (app *application) signIn (w http.ResponseWriter, r *http.Request){

	var form userSignInForm

	err := app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	// DO SOME VALIDATION

	form.CheckField(validator.NotBlank(form.Email),"email","Please enter your registered email address")

	form.CheckField(validator.Matches(form.Email,validator.EmailRX),"email","Please provide a valid email")

	
	form.CheckField(validator.NotBlank(form.Password), "password","Please enter your password")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"signin.html",data)
		return
	}


	// CHECK WITH DATABASE IF CREDENTIALS ARE VALID
	id , err := app.users.Authenticate(form.Email,form.Password)

	if err != nil {
		if errors.Is(err,models.ErrInvalidCredentials){
			form.AddNonFieldError("Incorrect email or password")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w,http.StatusUnprocessableEntity,"signin.html",data)
		}else {
			app.serverError(w,err)
		}
		return
	}

	// USE THE RENEWTOKEN() METHOD ON THE CURRENT SESSION TO CHANGE THE SESION
	// IT IS A GOOOD PRATICE TO GENERATE A NEW SESSION ID WHEN THE AUTHENTICATION STATE OR PRIVILEGE LEVELS CHANGES
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w,err)
		return
	}

	
	
	app.sessionManager.Put(r.Context(),"flash","Successfully signed in")

// ADD ID OF THE CURRENT USER TO THE SESSION SO THAT THEY ARE NOW LOGGED IN
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	
}

func (app *application) logout (w http.ResponseWriter, r *http.Request){

	// Use the RenewToken() method on the current session to change the session
	err := app.sessionManager.RenewToken(r.Context())


	
	if err != nil {
		fmt.Printf("check: %+v", err)
		app.serverError(w, err)
		return
	}

	// Remove the authenticatedUserID from the session data so that the user is logged out	
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(),"flash", "successfully logged out")
	

	http.Redirect(w,r,"/login", http.StatusSeeOther)
}

func (app *application) profileGet(w http.ResponseWriter, r *http.Request){

	params := httprouter.ParamsFromContext(r.Context())
	id,err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
	 app.serverError(w,err)
	 return
	}

	user, err := app.users.GetUser(id)

	if err != nil {
		if errors.Is(err, models.ErrorNoRecord){
			app.notFound(w)
		}else{
			app.serverError(w,err)
			return
		}
	}


	data := app.newTemplateData(r)
	data.User = user

	app.render(w,http.StatusOK,"profile.html",data)
}

func (app * application) UpdateProfileForm (w http.ResponseWriter, r*http.Request){
	params := httprouter.ParamsFromContext(r.Context())
	userId,err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		app.serverError(w,err)
		return
	}

	user,err := app.users.GetUser(userId)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			app.notFound(w)
			return
		}else{
			app.serverError(w,err)
			return
		}
	}
	

	data := app.newTemplateData(r)
	data.Form = updateProfileForm{}

	data.User = user

	app.render(w,http.StatusOK,"updateprofile.html",data)
}

func (app *application) UpdateProfile (w http.ResponseWriter, r *http.Request){
	param := httprouter.ParamsFromContext(r.Context())
	userId,err := strconv.Atoi(param.ByName("userId"))
	if err != nil {
		app.notFound(w)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	var form updateProfileForm

	err = app.decodePostForm(r,&form)
	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	user,err := app.users.GetUser(userId)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			app.notFound(w)
			return
		}else{
			app.serverError(w,err)
			return
		}
	}

	form.CheckField(validator.NotBlank(form.FirstName), "firstName","pls provide your first name")
	form.CheckField(validator.NotBlank(form.LastName), "lastName","pls provide your last name")
	form.CheckField(validator.NotBlank(form.Email),"email","Pls provide your email address")
	form.CheckField(validator.Matches(form.Email,validator.EmailRX),"email","Pls provide a valid email address")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		data.User = user
		 app.render(w,http.StatusUnprocessableEntity,"updateprofile.html",data)
		 return
	}

	if userId != app.sessionManager.GetInt(r.Context(),"authenticatedUserID"){
		app.clientError(w,http.StatusForbidden)
		return
	}


	err = app.users.UpdateUserProfile(userId,form.FirstName,form.LastName,form.Email)

	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail){
			form.AddNonFieldError(form.Email + " already in use")
			data := app.newTemplateData(r)
			data.Form = form
			data.User = user
			app.render(w,http.StatusUnprocessableEntity,"updateprofile.html",data)
			return
		}else{
			app.serverError(w,err)
			return
		}
	
	}

	app.sessionManager.Put(r.Context(),"flash","Profile updated successfully")

	http.Redirect(w,r,fmt.Sprintf("/profile/%d",userId),http.StatusSeeOther)

}


func (app *application) ChangePasswordForm (w http.ResponseWriter,r *http.Request){
	var form changePasswordForm

	data := app.newTemplateData(r)

	data.Form = form

	app.render(w,http.StatusOK,"changepassword.html", data)
}

func (app *application) ChangePassword (w http.ResponseWriter,r *http.Request){
	err := r.ParseForm()

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var form changePasswordForm

	err = app.decodePostForm(r, &form)
	if err != nil {
		app.serverError(w,err)
		return
	}

	form.CheckField(validator.NotBlank(form.Password),"password","password field is required")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w,http.StatusUnprocessableEntity,"changepassword.html",data)
		return
	}

	userId :=  app.sessionManager.GetInt(r.Context(),"authenticatedUserID")

	_, err = app.users.IsPasswordCorrect(userId, form.Password)

	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials){
			form.AddNonFieldError("incorrect password")
			data := app.newTemplateData(r)
			data.Form = form

			app.render(w,http.StatusUnprocessableEntity,"changepassword.html",data)
			return
		}
		app.serverError(w,err)

		return
	}


		app.sessionManager.Put(r.Context(), "canResetPassword", true)

		data := app.newTemplateData(r)
		data.Form = resetPasswordForm{}
		app.render(w,http.StatusOK,"resetpassword.html",data)
}


func (app *application) ResetPassword (w http.ResponseWriter, r *http.Request) {
	authenticatedUser := app.sessionManager.GetInt(r.Context(),"authenticatedUserID")    
		if authenticatedUser == 0 {
		app.clientError(w, http.StatusForbidden)
		return
	}

    // Check if the user has the session flag set
    canResetPassword := app.sessionManager.PopBool(r.Context(), "canResetPassword")
    if !canResetPassword {
		http.Redirect(w,r,"/change-password", http.StatusForbidden)
        return
    }




	err := r.ParseForm()
	if err != nil {
		app.serverError(w,err)
		return
	}

	var form resetPasswordForm

	err = app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Password),"password","Field cannot be blank")
	form.CheckField(validator.MinChars(form.Password,8),"password","password must be at least8 characters")
	form.CheckField(validator.NotBlank(form.RetypePassword),"reTypePassword","Pls re-enter your password")
	form.CheckField(validator.IsPasswordMatch(form.Password,form.RetypePassword),"reTypePassword","Password do not match")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w,http.StatusUnprocessableEntity,"resetpassword.html",data)
		return
	}



	 err = app.users.ResetPassword(authenticatedUser, form.Password)
	if err != nil {
		app.serverError(w,err)
		return
	}

	app.sessionManager.Put(r.Context(),"flash", "password changed successfully")

	http.Redirect(w,r,fmt.Sprintf("/profile/%d",authenticatedUser),http.StatusSeeOther)
}

func (app *application) ForgotPasswordForm (w http.ResponseWriter,r *http.Request){
	data := app.newTemplateData(r)
	data.Form = forgetPasswordForm{}

	app.render(w,http.StatusOK,"forgotpassword.html",data)
}

func (app *application) ForgotPassword (w http.ResponseWriter, r *http.Request){
	var form forgetPasswordForm

	err := app.decodePostForm(r,&form)

	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email),"email","Pls provide your registered email")
	form.CheckField(validator.Matches(form.Email,validator.EmailRX),"email","Pls enter a valid email address")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"forgotpassword.html",data)
		return
	}

	exist, userID, err := app.users.IsEmailExists(form.Email)

	if err != nil {
		app.serverError(w,err)
		return
	}

	if !exist {
		data := app.newTemplateData(r)
		form.AddNonFieldError("Email is not registered")
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"forgotpassword.html",data)
		return
	}

	err = app.users.SendResetCode(userID,form.Email)

	if err != nil {
		app.sendMailError(w, err)
		return
        // app.serverError(w, fmt.Errorf("failed to send password reset email: %w", err))
        // return
	}


	data := app.newTemplateData(r)
	data.Form = forgetPasswordForm{
		UserID: userID,
	}

	app.render(w,http.StatusOK,"verifyresetcode.html",data)

}


func (app *application) VerifyResetCode (w http.ResponseWriter, r *http.Request){
	var form forgetPasswordForm

	err := app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.ResetCode),"resetCode","pls enter the 6 digit pin sent to your registered email")
	form.CheckField(validator.MaxChars(form.ResetCode,6),"resetCode","invalid code")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"verifyresetcode.html", data)
		return
	}

	 err =	app.users.VerifyResetCode(form.UserID,(form.ResetCode))

	 if err != nil {
		data := app.newTemplateData(r)
		form.AddNonFieldError("wrong code, pls provide a valid code")
		data.Form = form
		app.render(w,http.StatusUnprocessableEntity,"verifyresetcode.html",data)
		return	
	 }

	app.sessionManager.Put(r.Context(),"forgotCanReset", true)

	http.Redirect(w,r, fmt.Sprintf("/forgot-reset-password/%d",form.UserID), http.StatusSeeOther)
}


func (app *application) ForgotResetPasswordForm (w http.ResponseWriter,r *http.Request){
	allowed := app.sessionManager.GetBool(r.Context(),"forgotCanReset")

	if !allowed {
		http.Redirect(w,r,"/forgot-password",http.StatusSeeOther)
		return
	}
	param := httprouter.ParamsFromContext(r.Context())
	userId, err := strconv.Atoi(param.ByName("userId"))

	if err != nil {
		app.serverError(w,err)
		return
	}

	data := app.newTemplateData(r)

	data.Form = resetPasswordForm{
		UserID: userId,
	}

	app.render(w,http.StatusOK,"forgotresetpassword.html", data)

}

func (app *application) ForgotResetPassword (w http.ResponseWriter,r *http.Request) {
	param  := httprouter.ParamsFromContext(r.Context())

	userId, err := strconv.Atoi(param.ByName("userId"))

	if err != nil {
		app.serverError(w,err)
		return
	}


	var form resetPasswordForm

	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w,http.StatusBadRequest)
		return
	}

	

	form.CheckField(validator.NotBlank(form.Password),"password","pls provide your new password")
	form.CheckField(validator.MinChars(form.Password,8),"password","password must be at least 8 characters")
	form.CheckField(validator.NotBlank(form.RetypePassword),"reTypePassword","pls re-enter your password")
	form.CheckField(validator.IsPasswordMatch(form.Password,form.RetypePassword),"reTypePassword","password does not match")

	if !form.Valid(){
		data := app.newTemplateData(r)
		data.Form =form
		app.render(w,http.StatusUnprocessableEntity,"forgotresetpassword.html", data)
		return
	}

	err = app.users.ResetPassword(userId,form.Password)

	if err != nil {
		app.serverError(w,err)
		return
	}
	
	app.sessionManager.PopBool(r.Context(), "forgotCanReset")
	app.sessionManager.Put(r.Context(), "flash", "password reset successfully, enter new password to login")

	http.Redirect(w,r,"/login",http.StatusSeeOther)
}




