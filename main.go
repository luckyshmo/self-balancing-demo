package main

import (
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

func getAttemptsNumberFromHeader(req *http.Request) int {
	attempts := req.Header.Get("attempts")
	attemptsNumber, err := strconv.Atoi(attempts)
	if err != nil {
		log.Print("Probably zero attempts")
	}
	return attemptsNumber
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	req.Header.Set("attempts", fmt.Sprint(getAttemptsNumberFromHeader(req)+1))
	log.Println("Redirect to " + redirectUrl)

	proxy.ServeHTTP(res, req)
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

func handleRequestOrRedirect(res http.ResponseWriter, req *http.Request) {

	attemptsNum := getAttemptsNumberFromHeader(req)
	log.Println(fmt.Sprintf("attempts: %d", attemptsNum))
	if attemptsNum > 5 {
		http.Error(res, "Service not available", http.StatusServiceUnavailable)
		return
	}

	if !Limiter.Allow() {
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
