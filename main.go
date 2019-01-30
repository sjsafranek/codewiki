package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_PORT = 1337
)

var (
	logger     = ligneous.NewLogger()
	PORT   int = DEFAULT_PORT
)

func init() {
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.StringVar(&CONTENT_DIRECTORY, "C", DEFAULT_CONTENT_DIRECTORY, "Wiki content directory")
	flag.Parse()
}

func main() {
	var err error
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// wiki engine
	wiki := &WikiEngine{}
	WIKI_DIRECTORY = fmt.Sprintf("%v/wiki/", CONTENT_DIRECTORY)
	err = os.MkdirAll(WIKI_DIRECTORY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	router.PathPrefix("/").Handler(wiki)
	//.end

	router.Use(LoggingMiddleWare, SetHeadersMiddleWare)

	logger.Infof("Magic happens on port %v...", PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), router)
	if nil != err {
		panic(err)
	}
}
