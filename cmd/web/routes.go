package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/msdoshack/mycodedairy/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		app.notFound(w)
	})


	// SERVE STATIC FILES Without using the embeded stystem
	// fileServer := http.FileServer(http.Dir("./ui/static/"))
	// router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))
	
	// SERVER STATIC FILES WITH NEW GO EMBEDED SYSTEM
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// SESSION MIDDLEWARE FOR DYNAMIC ROUTE
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf,app.authenticate)
	

	// PUBLIC ROUTES

	router.Handler(http.MethodGet, "/login",dynamic.ThenFunc(app.signInForm))
	router.Handler(http.MethodPost,"/login",dynamic.ThenFunc(app.signIn))
	router.Handler(http.MethodGet,"/sign-up",dynamic.ThenFunc(app.signUpForm))
	router.Handler(http.MethodPost,"/sign-up", dynamic.ThenFunc(app.signUp))
	router.Handler(http.MethodGet,"/forgot-password",dynamic.ThenFunc(app.ForgotPasswordForm))
	router.Handler(http.MethodPost,"/forgot-password", dynamic.ThenFunc(app.ForgotPassword))
	router.Handler(http.MethodPost,"/verify-reset",dynamic.ThenFunc(app.VerifyResetCode))
	router.Handler(http.MethodGet,"/forgot-reset-password/:userId",dynamic.ThenFunc(app.ForgotResetPasswordForm))
	router.Handler(http.MethodPost,"/forgot-reset-password/:userId",dynamic.ThenFunc(app.ForgotResetPassword))

	// PROCTECTED ROUTES
	protected := dynamic.Append(app.requireAuth)
	router.Handler(http.MethodGet,"/", protected.ThenFunc(app.home))
	router.Handler(http.MethodGet,"/snippet/:id", protected.ThenFunc(app.singleSnippets))
	router.Handler(http.MethodGet,"/user/snippets", protected.ThenFunc(app.userSnippets))
	router.Handler(http.MethodGet,"/user/search", protected.ThenFunc(app.userSnippetsSearch))
	router.Handler(http.MethodPost,"/add-snippet", protected.ThenFunc(app.addSnippet))
	router.Handler(http.MethodGet,"/add-snippet", protected.ThenFunc(app.addSnippetForm))
	router.Handler(http.MethodGet,"/update-snippet/:id", protected.ThenFunc(app.updateSnippetForm))
	router.Handler(http.MethodPost,"/update-snippet/:id", protected.ThenFunc(app.updateSnippet))
	router.Handler(http.MethodGet,"/profile/:userId",protected.ThenFunc(app.profileGet))
	router.Handler(http.MethodGet,"/update-profile/:userId",protected.ThenFunc(app.UpdateProfileForm))
	router.Handler(http.MethodPost,"/update-profile/:userId", protected.ThenFunc(app.UpdateProfile))
	router.Handler(http.MethodGet,"/change-password",protected.ThenFunc(app.ChangePasswordForm))
	router.Handler(http.MethodPost,"/change-password",protected.ThenFunc(app.ChangePassword))
	router.Handler(http.MethodPost,"/reset-password",protected.ThenFunc(app.ResetPassword))

	router.Handler(http.MethodGet, "/logout", protected.ThenFunc(app.logout))
	router.Handler(http.MethodPost,"/del-snippet/:id", protected.ThenFunc(app.DeleteSnippet))

	// middleware for logger, headers etc that wraps the whole app
	standard := alice.New(app.recoverPanic, app.logRequest,secureHeaders)

	return standard.Then(router)
}