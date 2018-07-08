package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ericMLindgren/testbox/internal/challenges"

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
	Stdouts []string `json:"stdouts,omitempty"`
	Grades  []string `json:"grades,omitempty"`
	Hints   []string `json:"hints,omitempty"`
	Message message  `json:"message"`
}

func (r *ExecutionResult) grade(c challenges.Challenge, scrub bool) {
	r.Grades = make([]string, len(r.Stdouts))
	r.Hints = make([]string, len(r.Stdouts))

	for i, c := range c.Cases {
		if c.Expect == r.Stdouts[i] {
			r.Grades[i] = "Pass"
		} else {
			r.Grades[i] = "Fail"
		}
		r.Hints[i] = c.Desc

	}
	if scrub { // avoid sending sensative info
		r.Stdouts = nil
	}
}

// type grade struct {
// 	Case   challenges.TestCase `json:"case"`
// 	Actual string              `json:"actual"`
// 	Grade  string              `json:"grade"`
// }

// func (g grade) String() string {
// 	return fmt.Sprintf("<grade>Case Expects: '%s' Actual: '%s' Grade: %s", g.Case.Expect, g.Actual, g.Grade)
// }

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
var fullAddress string

// // main for production, no cors support as admin page is served here...
// func main() {
// 	// Read settings. Compilebox address/port
// 	port, portOk := os.LookupEnv("COMPILEBOX_PORT")
// 	address, addressOk := os.LookupEnv("COMPILEBOX_ADDRESS")
// 	if !portOk || !addressOk {
// 		log.Fatal("Missing xaqt environment variables, please make sure service is available")
// 	}
// 	fullAddress = address + ":" + port

// 	// Check to ensure the xaqt is up by trying to fill langs variable
// 	fmt.Printf("Requesting language list from xaqt (%s)...\n", fullAddress)
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
	port := getEnv("XAQT_PORT", "31337")
	address := getEnv("XAQT_ADDRESS", "http://localhost")

	fullAddress = address + ":" + port

	// Check to ensure the xaqt is up by trying to fill langs variable
	fmt.Printf("Requesting language list from xaqt (%s)...\n", fullAddress)
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

func getChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received getChallenge")
	resp := new(apiResponse)

	var id challenges.ID
	err := decodeSubmission(r, &id)
	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {

		// If we got a valid submission
		fmt.Printf(" Challenge id: %s\n", id)

		c, err := challenges.GetById(id)

		if err != nil {
			resp.ErrorMessage = err.Error()
		} else {
			resp.pack(c)
		}
	}

	jsonWrite(resp, w)
}

func insertChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")
	resp := new(apiResponse)

	var c challenges.Challenge
	err := decodeSubmission(r, &c)

	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {
		// Submission is valid:
		fmt.Printf("Challenge struct: %v\n", c)

		id, err := challenges.Insert(&c)

		resp.ID = id
		if err != nil {
			resp.ErrorMessage = err.Error()
		}
	}

	jsonWrite(resp, w)
}

func updateChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received update request")
	resp := new(apiResponse)

	var c challenges.Challenge
	err := decodeSubmission(r, &c)

	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {
		fmt.Printf("Challenge struct: %v\n", c)

		err := challenges.Update(c.ID, &c)

		if err != nil {
			resp.ErrorMessage = err.Error()
		}
	}

	jsonWrite(resp, w)
}

func deleteChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received delete request")
	resp := new(apiResponse)

	var id challenges.ID
	err := decodeSubmission(r, &id)

	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {
		// fmt.Printf("Challenge struct: %v\n", c)
		err := challenges.Delete(id)
		if err != nil {
			resp.ErrorMessage = err.Error()
		}
	}

	jsonWrite(resp, w)
}

func getRandChallenge(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetRandom()
	resp := new(apiResponse)
	resp.pack(c)

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
	resp := new(apiResponse)

	// decode submission
	var submission CodeSubmission
	err := decodeSubmission(r, &submission)

	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {

		fmt.Printf(" with stdin: %s\n", submission.Stdins[0])

		// send to xaqt
		result, err := execCodeSubmission(submission)

		if err != nil {
			resp.ErrorMessage = err.Error()
		} else {
			resp.pack(result)
		}
	}

	jsonWrite(resp, w)
}

func submitToChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received code submission for challenge...")
	resp := new(apiResponse)

	// Decode client submission
	var submission CodeSubmission
	err := decodeSubmission(r, &submission)
	if err != nil {
		resp.ErrorMessage = err.Error()
	} else {

		// attach the appropriate challenge's stdins to the submission
		c, err := challenges.GetById(submission.ChallengeID)
		if err != nil {
			fmt.Printf("Error retrieving challenge: %s\n", err)
			resp.ErrorMessage = "Testbox internal failure: challenge retrieval"
			jsonWrite(resp, w)
			return
		}

		// pack challenge stdins into submission
		stdins, _ := c.GetIOSplit()
		submission.Stdins = stdins

		// send to xaqt
		result, err := execCodeSubmission(submission)

		if err != nil {
			// if trouble exec'ing, just send error
			resp.ErrorMessage = err.Error()
			jsonWrite(resp, w)
		} else {
			// grade + pack results
			result.grade(c, true)
			resp.pack(result)
		}
	}

	jsonWrite(resp, w)
}

// TODO consider that grading in strings should be done client side and that testbox side it should be bools?
// type gradeMap map[challenges.TestCase]string
func (r *apiResponse) pack(i interface{}) {
	buf, _ := json.Marshal(i)
	r.Result = string(buf)
}

func fetchLanguages() map[string]languageDetail {
	r, err := http.Get(fullAddress + "/languages/")

	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to contact xaqt, please ensure service is available:\n%s", err))
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var l map[string]languageDetail
	err = decoder.Decode(&l)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to decode xaqt response to language list request:\n%s", err))
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

	// send code to xaqt
	r, err := http.Post(fullAddress+"/evaluate/", "application/json", buf)
	if err != nil {
		fmt.Printf("error posting to xaqt: %s\n", err.Error())
		return ExecutionResult{}, err
	}

	// decode response
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var result ExecutionResult
	err = decoder.Decode(&result)

	if err != nil {
		fmt.Printf("error decoding xaqt response: %s\n", err.Error())
		return ExecutionResult{}, err
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

func decodeSubmission(r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&s)

	if err != nil {
		return err
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Environment variable %s not found, setting to %s\n", key, fallback)
	os.Setenv(key, fallback)
	return fallback
}
