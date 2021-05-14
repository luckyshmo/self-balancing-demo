package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/time/rate"
)

var Limiter *rate.Limiter
var port string
var redirectUrl string

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	attempts := getAttemptsFromContext(req)
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	att := req.Header.Get("att")
	attN, _ := strconv.Atoi(att)
	req.Header.Set("att", fmt.Sprint(attN+1))
	// Note that ServeHttp is non blocking and uses a go routine under the hood

	ctx := context.WithValue(req.Context(), "att", attempts+1)
	log.Println("Redirect to " + redirectUrl)

	proxy.ServeHTTP(res, req.WithContext(ctx))
}

func handle(req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}
	str := string(body)

	log.Println(str)
}

func getAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value("att").(int); ok {
		log.Println("Att found")
		return attempts
	}
	return 1
}

func handleRequestOrRedirect(res http.ResponseWriter, req *http.Request) {

	att := req.Header.Get("att")
	attN, _ := strconv.Atoi(att)
	log.Println(fmt.Sprintf("attN: - %d", attN))
	if attN > 5 {
		http.Error(res, "Service not available", http.StatusServiceUnavailable)
		return
	}

	attempts := getAttemptsFromContext(req)
	log.Println(fmt.Sprintf("Att: - %d", attempts))
	if attempts > 5 {
		http.Error(res, "Service not available", http.StatusServiceUnavailable)
		return
	}
	if Limiter.Allow() == false {
		serveReverseProxy(redirectUrl, res, req)
	} else {
		log.Println("Req handle at: " + req.Host)
		handle(req)
	}
}

func setupLimiter(maxReq string) {

	rMax, err := strconv.Atoi(maxReq)
	if err != nil {
		log.Println("Use deafult max req count")
		rMax = 1
	}
	Limiter = rate.NewLimiter(1, rMax) //rMax req in 1 sec
}

func main() {
	maxReq := os.Getenv("R_MAX")

	flag.StringVar(&redirectUrl, "redirect", "", "url for riderrect")
	flag.StringVar(&maxReq, "rn", "", "req number to handle")
	flag.StringVar(&port, "port", "", "app port")

	flag.Parse()

	setupLimiter(maxReq)

	if port == "" {
		log.Fatal("App port is empty")
	}
	if redirectUrl == "" {
		log.Fatal("Redirect url is empty")
	}

	http.HandleFunc("/", handleRequestOrRedirect)
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		panic(err)
	}

}
