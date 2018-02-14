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
)

type languageDetail struct {
	Boilerplate   string `json:"boilerplate"`
	CommentPrefix string `json:"commentPrefix"`
}

type CodeSubmission struct {
	Language    string   `json:"language"`
	Code        string   `json:"code"`
	Stdins      []string `json:"stdins"`
	ChallengeId string   `json:"challengeId,omitempty`
}

type ExecutionResult struct {
	Stdouts []string          `json:"stdouts"`
	Graded  map[string]string `json:"graded,omitempty"`
	Message message           `json:"message"`
}

// Message ...
type message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var languages map[string]languageDetail
var cbAddress string

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
	populateLanguages()

	/* Serve REST API endpoints for

	- various challenge searches

	- simple code submission (just pass along to testbox)

	- code submission checked against challenge

	*/

	http.HandleFunc("/get_challenge/", getChallenge)
	http.HandleFunc("/submit/", submitToChallenge)
	http.HandleFunc("/stdout/", getStdout)
	http.HandleFunc("/languages/", getLangs)
	// http.HandleFunc("/", frontPage)

	port = getEnv("TESTBOX_PORT", "31336")
	fmt.Println("testbox listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	/* Serve administrative interface for challenge collection

	 */
}

func getLangs(w http.ResponseWriter, r *http.Request) {
	log.Println("Received languages list request")

	buf, _ := json.MarshalIndent(languages, "", "   ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getChallenge(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetRandom()

	buf, _ := json.MarshalIndent(c, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getStdout(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Received stdout-only code submission...")

	// decode submission
	submission := decodeSubmission(r)

	fmt.Printf("with stdin: %s\n", submission.Stdins[0])

	// send to compilebox
	result := execCodeSubmission(submission)

	// encode and write result back to client
	jsonWrite(result, w)
}

func submitToChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Received code submission for challenge...")

	submission := decodeSubmission(r)

	fmt.Printf("Code targets challenge %s", submission.ChallengeId)
	fmt.Println(submission)

	// attach the appropriate challenge's stdins to the submission
	stdins, expectedStdouts := challenges.GetById(submission.ChallengeId).GetIOSplit()
	submission.Stdins = stdins

	// send to compilebox
	result := execCodeSubmission(submission)

	// compare stdouts to challenges stdouts
	fmt.Printf("about to grade, expecting %v\n", expectedStdouts)
	result.Graded = gradeResults(stdins, expectedStdouts, result.Stdouts)

	// encode and write result back to client
	jsonWrite(result, w)
}

// TODO consider that grading in strings should be done client side and that testbox side it should be bools?
func gradeResults(ins, exp, actual []string) map[string]string {
	graded := make(map[string]string)
	for i, e := range exp {
		// if we've run out of results, fail. This should never happen but would cause index error if it did
		// if i > len(actual)-1 {
		// 	graded[e] = "Fail"
		// 	continue
		// }
		thisIn := ins[i]

		if e != actual[i] {
			graded[thisIn] = "Fail"
			continue
		}

		graded[thisIn] = "Pass"
	}
	return graded
}

func populateLanguages() {
	r, err := http.Get(cbAddress + "/languages/")

	if err != nil {
		log.Fatal("Unable to contact compilebox, please ensure service is available")
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	// b := make([]byte, 256)
	// _, _ = r.Body.Read(b)
	// fmt.Printf("response: %s", b)

	err = decoder.Decode(&languages)
	if err != nil {
		panic(err)
	}

	supportedLangs := make([]string, 0, len(languages))
	for k := range languages {
		supportedLangs = append(supportedLangs, fmt.Sprintf("%s", k))
	}
	fmt.Printf("Supporting: %s\n", strings.Join(supportedLangs, ", "))
}

func execCodeSubmission(s CodeSubmission) ExecutionResult {
	// encode submission
	jsonBytes, _ := json.MarshalIndent(s, "", "   ")
	buf := bytes.NewBuffer(jsonBytes)

	// send code to compilebox
	r, err := http.Post(cbAddress+"/eval/", "application/json", buf)
	if err != nil {
		panic(err)
	}

	// decode response
	decoder := json.NewDecoder(r.Body)
	var result ExecutionResult
	err = decoder.Decode(&result)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	// return result
	return result
}

func jsonWrite(v interface{}, w http.ResponseWriter) {
	// encode variable
	buf, _ := json.MarshalIndent(v, "", "   ")

	// write encoded variable to client
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func decodeSubmission(r *http.Request) CodeSubmission {
	decoder := json.NewDecoder(r.Body)
	var submission CodeSubmission
	err := decoder.Decode(&submission)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}
	return submission
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Environment variable %s not found, setting to %s\n", key, fallback)
	os.Setenv(key, fallback)
	return fallback
}
