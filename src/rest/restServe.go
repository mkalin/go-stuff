package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Saying struct {
	Id         int
	Predictor  string   
	Prediction string   
}

// In effect, an in-memory data store. (Next is a real DB.)
type GlobalState struct {
	sayings   map[int]*Saying
   sayingId  int
   shutDown  bool
   minLen    int
	indent1   string
	indent2   string
	lock      sync.RWMutex
}
var gState *GlobalState

//** request handlers
// GET /sayingsXML
func SayingsXML(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	xmlDoc, err := xml.MarshalIndent(gState.ListifySayings(), gState.indent1, gState.indent2)
	sendResponse(response, xmlDoc, err)
	log.Println("/sayingsXML")
}

// GET /sayingXML/{id:[0-9]+}
func SayingXML(response http.ResponseWriter, request *http.Request) {	
	if gState.shutDown { return }

	// Extract and convert ID parameter. (Gorilla catches non-numeric Id.)
	n := mux.Vars(request)["id"]
	id, _ := strconv.Atoi(n)
	
	saying := readSaying(id)
	xmlDoc, err := xml.MarshalIndent(saying, gState.indent1, gState.indent2)
	sendResponse(response, xmlDoc, err)
	log.Println("/sayingXML/" + n)
}

// GET /sayingsJSON
func SayingsJSON(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	jsonDoc, err := json.MarshalIndent(gState.ListifySayings(), gState.indent1, gState.indent2)
	sendResponse(response, jsonDoc, err)
	log.Println("/sayingsJSON")
}

// GET /sayingJSON/{id:[0-9]+}
func SayingJSON(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	// Extract and convert ID parameter. (Gorilla catches non-numeric Id.)
	n := mux.Vars(request)["id"]
	id, _ := strconv.Atoi(n)

	saying := readSaying(id)
	jsonDoc, err := json.MarshalIndent(saying, gState.indent1, gState.indent2)
	sendResponse(response, jsonDoc, err)
	log.Println("/sayingXML/" + n)
}

// GET /sayingsPlain
func SayingsPlain(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	sendResponse(response, []byte(gState.StringifySayings()), nil)
	log.Println("/sayingsPlain")
}

// GET /sayingPlain/{id:[0-9]+}
func SayingPlain(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	// Extract and convert ID parameter. (Gorilla catches non-numeric Id.)
	n := mux.Vars(request)["id"]
	id, _ := strconv.Atoi(n)
	
	saying := readSaying(id)
	sendResponse(response, []byte(saying.ToString()), nil)
	log.Println("/sayingPlain/" + n)
}

// POST /saying
func SayingCreate(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	minLen := gState.minLen
	prediction := request.FormValue("prediction")
	predictor := request.FormValue("predictor")
	
	if len(prediction) < minLen || len(predictor) < minLen {
		err := errors.New("Prediction/predictor must be >= " + string(minLen) + " chars.")
		sendResponse(response, []byte(""), err)
		return
	}

	saying := new(Saying)
	saying.Prediction = prediction
	saying.Predictor = predictor

	// Insert into sayings list.
	gState.lock.Lock()
	saying.Id = gState.sayingId
	gState.sayings[saying.Id] = saying
	gState.sayingId++
	gState.lock.Unlock()

	msg := fmt.Sprintf("New Saying %d created\n.", saying.Id)
	sendResponse(response, []byte(msg), nil)
	log.Println("/sayingCreate")
}

// PUT /saying
func SayingEdit(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	// Id provided?
	id, err := strconv.Atoi(request.FormValue("id"))
	if (err != nil) {
		err := errors.New("No Id provided")
		sendResponse(response, ([]byte("")), err)
		return
	}

	// Need Prediction, Predictor, or both.
	minLen := gState.minLen
	prediction := request.FormValue("prediction")
	predictor := request.FormValue("predictor")
	if len(prediction) < minLen && len(predictor) < minLen {
		err := errors.New("Prediction/predictor must be >= " + string(minLen) + " chars.")
		sendResponse(response, []byte(""), err)
		return
	}

	// Update the Saying.
	gState.lock.Lock()
	if len(prediction) >= minLen {
		gState.sayings[id].Prediction = prediction
	}
	if len(predictor) >= minLen {
		gState.sayings[id].Predictor = predictor
	}
	gState.lock.Unlock()

	sendResponse(response, []byte("Saying " + request.FormValue("id") + " updated\n."), nil)
	log.Println("/sayingEdit/" + request.FormValue("id"))
}

// DELETE /saying/{id:[0-9]+}
func SayingDelete(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	// Extract and convert ID parameter. (Gorilla catches non-numeric Id.)
	n := mux.Vars(request)["id"]
	id, _ := strconv.Atoi(n)

	gState.lock.Lock()
	delete(gState.sayings, id)
	gState.lock.Unlock()

	sendResponse(response, []byte("Saying " + n + " deleted\n."), nil)
	log.Println("/sayingDelete/" + n)
}

// GET /reload (for test purposes only, hence not synchronized)
func Reload(response http.ResponseWriter, request *http.Request) {
	if gState.shutDown { return }

	initialize()
	sendResponse(response, []byte("Reloaded data."), nil)
	log.Println("/reload")
}

// Set up Gorilla router.
func startServer() {
   router := mux.NewRouter()

   router.HandleFunc("/sayingsXML", SayingsXML).Methods("GET")
	router.HandleFunc("/sayingXML/{id:[0-9]+}", SayingXML).Methods("GET")
	router.HandleFunc("/sayingsJSON", SayingsJSON).Methods("GET")
	router.HandleFunc("/sayingJSON/{id:[0-9]+}", SayingJSON).Methods("GET")
	router.HandleFunc("/sayingsPlain", SayingsPlain).Methods("GET")
	router.HandleFunc("/sayingPlain/{id:[0-9]+}", SayingPlain).Methods("GET")
	router.HandleFunc("/sayingCreate", SayingCreate).Methods("POST")
	router.HandleFunc("/sayingEdit", SayingEdit).Methods("PUT")
	router.HandleFunc("/sayingDelete/{id:[0-9]+}", SayingDelete).Methods("DELETE")
	router.HandleFunc("/reload", Reload).Methods("GET") // refresh the data

   http.Handle("/", router)

	fmt.Println("\nStarting server on port 9999...")
	http.ListenAndServe(":9999", router);
}

//** methods
func (s Saying) ToString() string {
   return fmt.Sprintf("%2d. %s says: %s", s.Id, s.Predictor, s.Prediction)
}

func (gs *GlobalState) SortSayings() []int {
	keys := []int{}

	gState.lock.RLock()
   for k, _ := range gState.sayings {
      keys = append(keys, k)
   }
	gState.lock.RUnlock()

   sort.Ints(keys)
	return keys
}

func (gs *GlobalState) Dumper(keys []int) {
	fmt.Println("\nPredictions:")	
	for _, k := range keys {
		fmt.Println(gState.sayings[k].ToString())
	}
}

func (gs *GlobalState) ListifySayings() []*Saying {
	list := []*Saying{}
	
	gState.lock.RLock()
	for _, v := range gs.sayings {
		list = append(list, v)
	}
	gState.lock.RUnlock()

	return list
}

func (gs *GlobalState) StringifySayings() string {
   var buffer bytes.Buffer

	gState.lock.RLock()
	keys := gs.SortSayings()
	for _, k := range keys {
		buffer.WriteString(gs.sayings[k].ToString() + "\n")
	}
	gState.lock.RUnlock()	

	return buffer.String()
}

//** utility functions
func readSaying(id int) *Saying {
   gState.lock.RLock()
   defer gState.lock.RUnlock()
	return gState.sayings[id]
}

func sendResponse(rw http.ResponseWriter, doc []byte, err error) {
	if err == nil {
		rw.Write(doc)
	} else {
		rw.Write([]byte(err.Error()))
	}
}

func readFile(file_name string) string {
	records, err := ioutil.ReadFile(file_name)
	if err != nil {
		log.Fatalln("Cannot read " + file_name + ". Exiting.")
	}
	return string(records)
}

func splitString(in string, delimiter string) []string {
	return strings.Split(in, delimiter)
}

func createSayings(inputs string) {
	var ss = splitString(inputs, "\n")
	if len(ss) < 1 {
		log.Fatalln("Need > 0 sayings.")
	}

	for _, saying := range ss {
		parts := splitString(saying, "!")
		if len(parts) == 2 {
			saying := new(Saying)
			saying.Id = gState.sayingId
			saying.Predictor = parts[0]
			saying.Prediction = parts[1]
			gState.sayings[gState.sayingId] = saying
			gState.sayingId++
		}
	}
}

func readAndDumpData() {
	createSayings(readFile("sayings.db"))
	gState.Dumper(gState.SortSayings())
}

func initialize() {
	gState = &GlobalState {
		sayings:   make(map[int]*Saying),
      sayingId:  1,
		indent1:   " ",
		indent2:   "  ",  
	   shutDown:  false,
      minLen:    6}
	readAndDumpData()
}

//** main
func main() {
	// Point var globalState to a GlobalState instance, which embeds
	// maps for Sayings together with auto-incremented counter for the Id.
   // The data are read from the file sayings.db.
	initialize()   

	// Create a Gorilla router that maps HTTP requests to handler functions
	// and start the HTTP server, which uses the router.
	go startServer()

	// Gracefully shut down by pausing to allow current requests to be
	// handled. No new requests are processed during shutdown.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM) // control-C
	log.Println(<-ch)

	gState.shutDown = true
	log.Println("Gracefully shutting down...")
	time.Sleep(time.Duration(5) * time.Second)
	os.Exit(0) // kill all goroutines
}
