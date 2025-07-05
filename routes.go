package main

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

func initializeRoutes(router *mux.Router) {
	fs := http.FileServer(http.Dir("static/"))
	s := http.StripPrefix("/static/", fs)
	router.PathPrefix("/static/").Handler(s)

	handleBlogPost := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(blogPostPageHandler)))
	handleRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(redirectHandler)))
	handleFilterPage := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(filtersPageHandler)))
	handleRSS := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(generateRSSHandler)))
	handleTagRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(tagRedirectHandler)))
	handlePaginationRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(paginationRedirectHandler)))
	handleBlogIdRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(blogIdRedirectHandler)))
	handleYearRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(yearRedirectHandler)))
	handleHome := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(homePageHandler)))
	handleAllPosts := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(allPostsHandler)))
	handleRunning := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(runningHandler)))

	router.Handle("/", handleHome)
	router.Handle("/rss", handleRSS)
	router.Handle("/blog/{id:[0-9]+}", handleBlogIdRedirect)
	router.Handle("/blog/{id:[0-9]+}/{slug}", handleBlogPost)
	router.Handle("/page/{page:[0-9]+}", handlePaginationRedirect)
	router.Handle("/tagged/{tag}", handleTagRedirect)
	router.Handle("/posted/{year:[0-9]+}", handleYearRedirect)
	router.Handle("/posted/{year:[0-9]+}/page/{page:[0-9]+}", handlePaginationRedirect)
	router.Handle("/blog/page/{page}", handleRedirect)
	router.Handle("/blog/tagged/{tag}", handleRedirect)
	router.Handle("/blog/tagged/{tag}/page/{page}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}/page/{page:[0-9]+}", handleRedirect)
	router.Handle("/blog/filters", handleFilterPage)
	router.Handle("/about", getMarkdownTemplateHandler("About", "about.markdown", "/about"))
	router.Handle("/contact", getMarkdownTemplateHandler("Contact", "contact.markdown", "/contact"))
	router.Handle("/projects", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(projectsHandler)),
	))
	router.Handle("/interviews", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(interviewsHandler)),
	))
	router.Handle("/advice", getMarkdownTemplateHandler("Advice", "advice.markdown", "/advice"))
	router.Handle("/books", getMarkdownTemplateHandler("Books", "books.markdown", "/books"))
	router.Handle("/resume", getMarkdownTemplateHandler("Resume", "resume.markdown", "/resume"))
	router.Handle("/running", handleRunning)
	router.Handle("/all", handleAllPosts)
}
