package main

import (
	"net/http"
	"time"
)

// type RequestStats struct {
// 	BytesOut int
// }
//
// var (
// 	requestStats RequestStats
// )

// https://ndersson.me/post/capturing_status_code_in_net_http/
// https://upgear.io/blog/golang-tip-wrapping-http-response-writer-for-middleware/
// https://www.reddit.com/r/golang/comments/7p35s4/how_do_i_get_the_response_status_for_my_middleware/
type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (self *statusWriter) WriteHeader(code int) {
	self.status = code
	self.ResponseWriter.WriteHeader(code)
}

func (self *statusWriter) Write(b []byte) (int, error) {
	if self.status == 0 {
		self.status = 200
	}
	n, err := self.ResponseWriter.Write(b)
	self.length += n
	return n, err
}

func LoggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// before
		logger.Debugf(" In %v %v %v", r.RemoteAddr, r.Method, r.URL)

		// Initialize the status to 200 in case WriteHeader is not called
		sw := statusWriter{w, 200, 0}

		// during
		next.ServeHTTP(&sw, r)

		// end
		logger.Debugf("Out %v %v %v [%v] %v - %v bytes", r.RemoteAddr, r.Method, r.URL, sw.status, time.Since(start), sw.length)

		// requestStats.BytesOut += sw.length
		// logger.Debug(requestStats.BytesOut)
	})
}

func SetHeadersMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}

func CORSMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// package main
//
// import (
// 	"fmt"
// 	"net/http"
// 	"strings"
// )
//
// // https://ndersson.me/post/capturing_status_code_in_net_http/
// // https://upgear.io/blog/golang-tip-wrapping-http-response-writer-for-middleware/
// type LoggingResponseWriter struct {
// 	http.ResponseWriter
// 	status int
// }
//
// func (self *LoggingResponseWriter) WriteHeader(code int) {
// 	self.status = code
// 	self.ResponseWriter.WriteHeader(code)
// }
//
// func LoggingMiddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// before
// 		logger.Debugf(" In %v %v %v", r.RemoteAddr, r.Method, r.URL)
//
// 		// during
// 		if strings.Contains(fmt.Sprintf("%v", r.URL), "/ws/") || strings.Contains(fmt.Sprintf("%v", r.URL), "/log") {
// 			next.ServeHTTP(w, r)
// 			return
// 		}
// 		//.end
//
// 		// Initialize the status to 200 in case WriteHeader is not called
// 		lrw := LoggingResponseWriter{w, 200}
//
// 		// during
// 		next.ServeHTTP(&lrw, r)
//
// 		// end
// 		logger.Debugf("Out %v %v %v [%v]", r.RemoteAddr, r.Method, r.URL, lrw.status)
// 	})
// }
//
// func SetHeadersMiddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Max-Age", "86400")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
// 		w.Header().Set("Access-Control-Allow-Credentials", "true")
//
// 		next.ServeHTTP(w, r)
// 	})
// }
