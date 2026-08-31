package main

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

func initializeRoutes(router *mux.Router) {
	fs := http.FileServer(http.Dir("static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

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
	handleSitemap := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(generateSitemapHandler)))
	handleRobots := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(robotsHandler)))
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
		negroni.HandlerFunc(gzipHTML),
		negroni.Wrap(http.HandlerFunc(runningHandler)))
	handlePictures := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(picturesHandler)),
	)
	handleCheatsheets := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(cheatsheetsHandler)),
	)
	handleCheatsheet := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(cheatsheetHandler)),
	)

	router.Handle("/", handleHome)
	router.Handle("/rss", handleRSS)
	router.Handle("/sitemap.xml", handleSitemap).Methods("GET")
	router.Handle("/robots.txt", handleRobots).Methods("GET")

	// Test error endpoint (remove in production) - place early to avoid conflicts
	router.Handle("/test/{errorcode:[4-5][0-9][0-9]}", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(testErrorHandler)))).Methods("GET")
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
	router.Handle("/blog/{id:[^/]*[^0-9/][^/]*}", handleBlogIdRedirect)
	router.Handle("/blog/{id:[^/]*[^0-9/][^/]*}/{slug}", handleBlogPost)
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
	router.Handle("/books", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(booksHandler)),
	))
	router.Handle("/resume", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(resumePageHandler)),
	))
	router.Handle("/career", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(careerPageHandler)),
	))
	router.Handle("/running", handleRunning)
	router.Handle("/cheatsheets", handleCheatsheets)
	router.Handle("/cheatsheets/{slug}", handleCheatsheet)
	router.Handle("/api/runs-for-date", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(runsForDateHandler)),
	))
	router.Handle("/api/running-totals", negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(runningTotalsHandler)),
	))
	router.Handle("/all", handleAllPosts)
	router.Handle("/pictures", handlePictures)

	// 404 handler for unmatched routes
	router.NotFoundHandler = negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorPage(w, "", "404")
		})))
}
