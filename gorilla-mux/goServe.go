package main

/** imports **/

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"regexp"
	"strconv"
)
/****/

/** data structures **/

type Saying struct {
	Id         int
	Company_id int
	Predictor  string
	Prediction string
}

type Company struct {
	Id        int
	Saying_id int
	CEO       string
	Name      string
	Address1  string
	Address2  string
}

type SayingCompany struct {
	Who   string
	Which string
	What  string
}
/****/

/** globals **/

var sayingsList = []*Saying{}
var companiesList = []*Company{}
/****/

/** request handlers **/

// GET /
// GET /home
func HomeH(response http.ResponseWriter, request *http.Request) {
	// home.html is static rather than templated HTML
	html := readFile("home.html")
	response.Write([]byte(html))

	log("home")
}

// POST /companies
func CompaniesH(response http.ResponseWriter, request *http.Request) {
	t := getTemplate("companies.html")
	t.Execute(response, companiesList)

	log("companies")
}

// POST /predictions
func PredictionsH(response http.ResponseWriter, request *http.Request) {
	t := getTemplate("predictions.html")
	t.Execute(response, sayingsList)

	log("sayings")
}

// GET /prediction
func PredictionH(response http.ResponseWriter, request *http.Request) {
	// Extract the user-provided index.
	request.ParseForm()
	form := request.Form
	id := form["saying"][0] // 1st and only member of a list

	flag, _ := regexp.MatchString("^[0-9]+$", id)
	if flag {
		sendResponse(response, id)
	} else {
		msg := "<html><body><h3>"
		msg += "Bad request! Please enter integers only in the text field."
		msg += "</h3><p><a href = '/home'>Home</a></p></body></html>"
		response.Write([]byte(msg))
	}

	log("saying")
}

// GET /predictionD/{id:[0-9]+}
func PredictionD(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]

	sendResponse(response, id)
	log("sayingD")
}

func sendResponse(response http.ResponseWriter, id string) {
	i, _ := strconv.Atoi(id) // Convert string to int.
	i = i % 16 // Ensure no out of bounds problems.

	// Set up template data.
	params := new(SayingCompany)
	params.Who = sayingsList[i].Predictor
	params.What = sayingsList[i].Prediction
	params.Which = companiesList[i].Name
	
	t := getTemplate("prediction.html")
	t.Execute(response, params)
}
/****/

/** primary functions **/

func main() {
	// Get lists of predictions and companies from the
	// data store (in this case, text files).
	sayingsList, companiesList = readData()

	flag := false // true for dump of lists
	if flag {
		dumpSayings(sayingsList)
		dumpCompanies(companiesList)
	}

	startServer()
}

func startServer() {
   router := mux.NewRouter()

	// Dispatch map
   router.HandleFunc("/", HomeH).Methods("GET")
	router.HandleFunc("/home", HomeH).Methods("GET")

	router.HandleFunc("/predictions", PredictionsH).Methods("POST")
   router.HandleFunc("/prediction", PredictionH).Methods("POST") 
	router.HandleFunc("/predictionD/{id:[0-9]+}", PredictionD).Methods("POST")

   router.HandleFunc("/companies", CompaniesH).Methods("POST")

	// Enable the router.
   http.Handle("/", router)

	// Start the server.
	fmt.Println("\nListening on port 8080...")
	http.ListenAndServe(":8080", router);
}
/****/

/** utilities **/

func createCompanies(cs string) []*Company {
	var records = splitString(cs, "\n")
	if len(records) < 1 {
		notifyAndMaybeDie("Need > 0 companies.", true)
	}

   var id = 1
	var company *Company
	var companies = []*Company{}

	for i, record := range records {
		n := i % 4
		switch {
		case n == 0:
			company = new(Company)
			company.Id = id
			company.Saying_id = id
			id++
			company.CEO = record
		case n == 1:
			company.Name = record
		case n == 2:
			company.Address1 = record
		case n == 3:
			company.Address2 = record
			companies = append(companies, company)
		}
	}
	return companies
}

func createSayings(inputs string) []*Saying {
	var ss = splitString(inputs, "\n")
	if len(ss) < 1 {
		notifyAndMaybeDie("Need >= 1 sayings.", true)
	}

	var sayings = []*Saying{}
	for i, saying := range ss {
		parts := splitString(saying, "!")
		if len(parts) == 2 {
			sayingStruct := new(Saying)
			sayingStruct.Id = i + 1
			sayingStruct.Company_id = i + 1
			sayingStruct.Predictor = parts[0]
			sayingStruct.Prediction = parts[1]
			sayings = append(sayings, sayingStruct)
		}
	}
	return sayings
}

func dumpCompanies(companies []*Company) {
	msg := fmt.Sprintf("\nThere are %v companies: ", len(companies))
	fmt.Println(msg);

	for _, company := range companies {
		c := fmt.Sprintf("\t%s\n\t%s\n\t%s\n\t%s", 
			company.Name, 
			company.CEO, 
			company.Address1, 
			company.Address2)
		fmt.Println(c)
	}
}

func dumpSayings(sayings []*Saying) {
	msg := fmt.Sprintf("\nThere are %v sayings: ", len(sayings))
	fmt.Println(msg);

	for _, saying := range sayings {
		s := fmt.Sprintf("%s says: %s", saying.Predictor, saying.Prediction)
		fmt.Println(s)
	}
}

func getTemplate(filename string) *template.Template {
	records := readFile(filename)

	tmpl, err := template.New("records").Parse(records)
	if err != nil { panic(err) }

	return tmpl
}

func notifyAndMaybeDie(msg string, die bool) {
	fmt.Println("\n!!! " + msg);
	if die {
		os.Exit(-1)
	}
}

func readData() ([]*Saying, []*Company) {
	sayings := readFile("sayings.db")
	companies := readFile("companies.db")
	return createSayings(sayings), createCompanies(companies)
}

func readFile(file_name string) string {
	records, err := ioutil.ReadFile(file_name)
	if err != nil {
		notifyAndMaybeDie("Cannot read " + file_name + ". Exiting.", true)
	}
	return string(records)
}

func splitString(in string, delimiter string) []string {
	return strings.Split(in, delimiter)
}

func log(msg string) {
	fmt.Println(msg)
}
/****/