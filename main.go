package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_PORT       = 1337
	DEFAULT_PASSPHRASE = ""
)

var (
	logger            = ligneous.NewLogger()
	PASSPHRASE string = DEFAULT_PASSPHRASE
	PORT       int    = DEFAULT_PORT
)

func init() {
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.StringVar(&DB_FILE, "db", DEFAULT_DB_FILE, "Database file")
	flag.StringVar(&PASSPHRASE, "e", DEFAULT_PASSPHRASE, "Passphrase for encryption")
	flag.Parse()
}

func main() {
	var err error
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	db, err := NewDatabase(DB_FILE, PASSPHRASE)
	if nil != err {
		panic(err)
	}

	db.CreateTable("pages")

	signal_queue := make(chan os.Signal)
	signal.Notify(signal_queue, syscall.SIGTERM)
	signal.Notify(signal_queue, syscall.SIGINT)
	go func() {
		sig := <-signal_queue
		logger.Warnf("caught sig: %+v", sig)
		logger.Warn("Gracefully shutting down...")
		db.Close()
		// logger.Warn("Shutting down...")
		time.Sleep(250 * time.Millisecond)
		os.Exit(0)
	}()

	// wiki engine
	wiki := &WikiEngine{db: db}
	router.PathPrefix("/").Handler(wiki)
	//.end

	router.Use(LoggingMiddleWare, SetHeadersMiddleWare, CORSMiddleWare)

	logger.Infof("Magic happens on port %v...", PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), router)
	if nil != err {
		panic(err)
	}
}
