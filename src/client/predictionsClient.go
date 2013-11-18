package main

import (
	"fmt"
	"strings"
	"net/http"
)

const (
	baseUrl    = "http://localhost:"
	portNumber = "8080"
	uri        = "/cliches2/"
)

func main() { // entry point
	url := buildUrl()
	fmt.Println(url)

	response, _ := http.Get("http://localhost:8080/cliches2/")
	defer response.Body.Close()
//	if (err != nil) {
//		fmt.Println(err)
//	} else { 
		fmt.Println(response) 
	//
	
}

func getXmlAll() {

}

func getXmlOne() {

}

func getJsonAll() {

}

func getJsonOne() { }
func post() { }
func put() { }
func remove() { }

func runTests() {
	getXmlAll()
	getXmlOne()
	getJsonAll()
	getJsonOne()
	post()
	getXmlAll()
	put()
	getXmlAll()
	remove()
	getXmlAll()
}

func buildUrl() string {
	var parts []string 
	parts = append(parts, baseUrl, portNumber, uri)
	return strings.Join(parts, "")
}