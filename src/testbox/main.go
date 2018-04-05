package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"challenges"

	"github.com/rs/cors"
)

type languageDetail struct {
	Boilerplate   string `json:"boilerplate"`
	CommentPrefix string `json:"commentPrefix"`
}

// CodeSubmission describes a snippet of executable code, the language it's in, and any stdins that should be piped in.
// CodeSubmission optionally contains the id of the challenge it should be tested against
type CodeSubmission struct {
	Language    string        `json:"language"`
	Code        string        `json:"code"`
	Stdins      []string      `json:"stdins"`
	ChallengeID challenges.ID `json:"challengeId,omitempty`
}

func (c CodeSubmission) String() string {
	str := "<CodeSubmission> "
	if c.ChallengeID > -1 {
		str += fmt.Sprintf("ChallengeID: %d", c.ChallengeID)
	}
	str += "\n"
	str += fmt.Sprintf("Lang: %s\n", c.Language)
	str += fmt.Sprintf("Stdins: %s\n", c.Stdins)
	str += fmt.Sprintf("Code: Hidden\n")
	return str
}

// ExecutionResult contains the outputs (stdouts) of submitted code, as well as an compiler messages (Message).
// ExecutionResult may contain a grade for the code if the code was tested against a challenge
type ExecutionResult struct {
	Stdouts []string `json:"stdouts"`
	Graded  []grade  `json:"graded,omitempty"`
	Message message  `json:"message"`
}

type grade struct {
	Case   challenges.TestCase `json:"case"`
	Actual string              `json:"actual"`
	Grade  string              `json:"grade"`
}

func (g grade) String() string {
	return fmt.Sprintf("<grade>Case Expects: '%s' Actual: '%s' Grade: %s", g.Case.Expect, g.Actual, g.Grade)
}

type message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type apiResponse struct {
	ErrorMessage string        `json:"error,omitempty"`
	ID           challenges.ID `json:"id,omitempty"`
	Result       string        `json:"result,omitempty"`
}

var languages map[string]languageDetail
var cbAddress string

// // main for production, no cors support as admin page is served here...
// func main() {
// 	// Read settings. Compilebox address/port
// 	port, portOk := os.LookupEnv("COMPILEBOX_PORT")
// 	address, addressOk := os.LookupEnv("COMPILEBOX_ADDRESS")
// 	if !portOk || !addressOk {
// 		log.Fatal("Missing compilebox environment variables, please make sure service is available")
// 	}
// 	cbAddress = address + ":" + port

// 	// Check to ensure the compilebox is up by trying to fill langs variable
// 	fmt.Printf("Requesting language list from compilebox (%s)...\n", cbAddress)
// 	languages = fetchLanguages()

// 	// routes for admin of challenge db
// 	http.HandleFunc("/challenges/id/", getChallenge)
// 	http.HandleFunc("/challenges/insert/", insertChallenge)
// 	http.HandleFunc("/challenges/delete/", deleteChallenge)
// 	http.HandleFunc("/challenges/update/", updateChallenge)

// 	// routes for requesting challenges
// 	http.HandleFunc("/challenges/rand/", getRandChallenge)
// 	http.HandleFunc("/challenges/all/", getAllChallenges)
// 	// http.HandleFunc("/challenges/search/", searchChallenges)

// 	// get list of supported languages:
// 	http.HandleFunc("/languages/", getLangs)

// 	// routes for submitting code
// 	http.HandleFunc("/submit/", submitToChallenge)
// 	http.HandleFunc("/stdout/", getStdout)

// 	// front page, should have admin login and project info
// 	// http.HandleFunc("/", frontPage)

// 	challenges.OpenDB()
// 	defer challenges.CloseDB()

// 	port = getEnv("TESTBOX_PORT", "31336")
// 	fmt.Println("testbox listening on " + port)
// 	log.Fatal(http.ListenAndServe(":"+port, nil))
// }

// main for testing challenge admin separtely, supports cors
func main() {
	// Read settings. Compilebox address/port
	port, portOk := os.LookupEnv("COMPILEBOX_PORT")
	address, addressOk := os.LookupEnv("COMPILEBOX_ADDRESS")
	if !portOk || !addressOk {
		log.Fatal("Missing compilebox environment variables, please make sure service is available")
	}
	cbAddress = address + ":" + port

	// Check to ensure the compilebox is up by trying to fill langs variable
	fmt.Printf("Requesting language list from compilebox (%s)...\n", cbAddress)
	languages = fetchLanguages()

	mux := http.NewServeMux()
	// routes for admin of challenge db
	mux.HandleFunc("/challenges/id/", getChallenge)
	mux.HandleFunc("/challenges/insert/", insertChallenge)
	mux.HandleFunc("/challenges/delete/", deleteChallenge)
	mux.HandleFunc("/challenges/update/", updateChallenge)

	// routes for requesting challenges
	mux.HandleFunc("/challenges/rand/", getRandChallenge)
	mux.HandleFunc("/challenges/all/", getAllChallenges)
	// mux.HandleFunc("/challenges/search/", searchChallenges)

	// get list of supported languages:
	mux.HandleFunc("/languages/", getLangs)

	// routes for submitting code
	mux.HandleFunc("/submit/", submitToChallenge)
	mux.HandleFunc("/stdout/", getStdout)

	// front page, should have admin login and project info
	// uncomment next two lines to enable open access to db
	// fs := http.FileServer(http.Dir("front_end/build/"))
	// mux.Handle("/", fs)

	challenges.OpenDB()
	defer challenges.CloseDB()

	handler := cors.Default().Handler(mux)
	port = getEnv("TESTBOX_PORT", "31336")
	fmt.Println("testbox listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// TODO standardize error messaging
func getChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")

	var id challenges.ID
	decodeSubmission(r, &id)

	fmt.Printf("Challenge id: %s\n", id)

	c, err := challenges.GetById(id)

	resp := new(apiResponse)
	if err != nil {
		resp.ErrorMessage = err.Error()
	}
	resp.pack(c)

	jsonWrite(resp, w)
}

func insertChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")

	var c challenges.Challenge
	decodeSubmission(r, &c)

	fmt.Printf("Challenge struct: %v\n", c)

	id, err := challenges.Insert(&c)

	resp := apiResponse{ID: id}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	jsonWrite(resp, w)
}

func updateChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received update request")

	var c challenges.Challenge
	decodeSubmission(r, &c)

	fmt.Printf("Challenge struct: %v\n", c)

	// TODO this is broken... this note is not helpful. broken how?
	err := challenges.Update(c.ID, &c)
	resp := apiResponse{}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	jsonWrite(resp, w)
}

func deleteChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received delete request")

	var id challenges.ID
	decodeSubmission(r, &id)

	// fmt.Printf("Challenge struct: %v\n", c)

	err := challenges.Delete(id)
	resp := apiResponse{}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	jsonWrite(resp, w)
}

func getRandChallenge(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetRandom()
	resp := new(apiResponse)
	resp.pack(c)
	// if err != nil {
	// 	resp.ErrorMessage = err.Error()
	// }

	jsonWrite(resp, w)
}

func getAllChallenges(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetAll()
	resp := new(apiResponse)
	resp.pack(c)

	jsonWrite(resp, w)
}

func getLangs(w http.ResponseWriter, r *http.Request) {
	log.Println("Received languages list request")
	resp := new(apiResponse)
	resp.pack(languages)
	jsonWrite(resp, w)
}

func getStdout(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Received stdout-only code submission...")

	// decode submission

	var submission CodeSubmission
	decodeSubmission(r, &submission)

	fmt.Printf("with stdin: %s\n", submission.Stdins[0])

	// send to compilebox
	result, err := execCodeSubmission(submission)
	resp := new(apiResponse)
	if err != nil {
		resp.ErrorMessage = err.Error()
	}
	resp.pack(result)

	// encode and write result back to client
	jsonWrite(resp, w)
}

func submitToChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received code submission for challenge...")

	var submission CodeSubmission
	decodeSubmission(r, &submission)

	// fmt.Printf("Code targets challenge %d", submission.ChallengeID)
	fmt.Println(submission)

	// attach the appropriate challenge's stdins to the submission
	c, err := challenges.GetById(submission.ChallengeID)
	if err != nil {
		panic(err) // TODO handle error more gracefully and pass along to user
	}

	stdins, _ := c.GetIOSplit()
	submission.Stdins = stdins

	// send to compilebox
	result, err := execCodeSubmission(submission)

	// compare stdouts to challenges stdouts
	// fmt.Printf("about to grade, expecting %v\n", expectedStdouts)
	// result.Graded = gradeResults(stdins, expectedStdouts, result.Stdouts)
	result.Graded = gradeResults(c.Cases, result.Stdouts)
	// fmt.Printf("Suspect challenge: %v", c)
	// fmt.Printf("Done grading results: %v", result.Graded)
	// encode and write result back to client
	resp := new(apiResponse)
	if err != nil {
		resp.ErrorMessage = err.Error()
	}
	resp.pack(result)
	jsonWrite(resp, w)
}

// TODO consider that grading in strings should be done client side and that testbox side it should be bools?
// type gradeMap map[challenges.TestCase]string
func (r *apiResponse) pack(i interface{}) {
	buf, _ := json.Marshal(i)
	r.Result = string(buf)
}

func gradeResults(cases challenges.CaseList, actual []string) []grade {
	graded := make([]grade, len(cases))
	for i, c := range cases {
		graded[i] = grade{Case: c, Actual: actual[i]}
		result := cases[i].Expect == actual[i]

		if !result {
			graded[i].Grade = "Fail"
			continue
		}

		graded[i].Grade = "Pass"
	}
	return graded
}

func fetchLanguages() map[string]languageDetail {
	r, err := http.Get(cbAddress + "/languages/")

	if err != nil {
		log.Fatal("Unable to contact compilebox, please ensure service is available")
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var l map[string]languageDetail
	err = decoder.Decode(&l)
	if err != nil {
		panic(err)
	}

	supportedLangs := make([]string, 0, len(l))
	for k := range l {
		supportedLangs = append(supportedLangs, fmt.Sprintf("%s", k))
	}
	fmt.Printf("Supporting: %s\n", strings.Join(supportedLangs, ", "))

	return l
}

func execCodeSubmission(s CodeSubmission) (ExecutionResult, error) {
	// encode submission
	jsonBytes, _ := json.MarshalIndent(s, "", "   ")
	buf := bytes.NewBuffer(jsonBytes)

	// send code to compilebox
	r, err := http.Post(cbAddress+"/eval/", "application/json", buf)
	if err != nil {
		fmt.Printf("error posting to compilebox: %s", err.Error())
	}

	// decode response
	decoder := json.NewDecoder(r.Body)
	var result ExecutionResult
	derr := decoder.Decode(&result)
	defer r.Body.Close()
	if derr != nil {
		fmt.Printf("error decoding compilebox response: %s", derr.Error())
	}

	// return result
	return result, err
}

func jsonWrite(v interface{}, w http.ResponseWriter) {
	// encode variable
	// buf, _ := json.MarshalIndent(v, "", "   ")
	buf, _ := json.Marshal(v)

	// write encoded variable to client
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func decodeSubmission(r *http.Request, s interface{}) {
	decoder := json.NewDecoder(r.Body)
	// var submission CodeSubmission
	err := decoder.Decode(&s)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}
	// return submission
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Environment variable %s not found, setting to %s\n", key, fallback)
	os.Setenv(key, fallback)
	return fallback
}
