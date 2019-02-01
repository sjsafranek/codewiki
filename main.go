package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_PORT              = 1337
	DEFAULT_PASSPHRASE        = ""
	PROJECT            string = "CodeZombie"
	VERSION            string = "0.1.1"
)

var (
	logger            = ligneous.NewLogger()
	PASSPHRASE string = DEFAULT_PASSPHRASE
	PORT       int    = DEFAULT_PORT
)

func init() {
	var print_version bool = false
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.StringVar(&DB_FILE, "db", DEFAULT_DB_FILE, "Database file")
	flag.StringVar(&PASSPHRASE, "e", DEFAULT_PASSPHRASE, "Passphrase for encryption")
	flag.BoolVar(&print_version, "V", false, "Print version and exit")
	flag.Parse()

	if print_version {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}
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

	logger.Infof("%v-%v", PROJECT, VERSION)
	logger.Debug("GOOS: ", runtime.GOOS)
	logger.Debug("CPUS: ", runtime.NumCPU())
	logger.Debug("PID: ", os.Getpid())
	logger.Debug("Go Version: ", runtime.Version())
	logger.Debug("Go Arch: ", runtime.GOARCH)
	logger.Debug("Go Compiler: ", runtime.Compiler)
	logger.Debug("NumGoroutine: ", runtime.NumGoroutine())

	logger.Infof("Magic happens on port %v...", PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%v", PORT), router)
	if nil != err {
		panic(err)
	}
}
