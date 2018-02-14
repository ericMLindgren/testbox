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
	ChallengeId string   `json:"challengeId`
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
	// langs := make(map[string]sandbox.Language)

	// for k, v := range box.LanguageMap {
	// 	langs[k] = sandbox.Language{Boilerplate: v.Boilerplate, CommentPrefix: v.CommentPrefix}
	// }

	// add boilerplate and comment info
	// log.Println(langs)

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
	decoder := json.NewDecoder(r.Body)
	var submission CodeSubmission
	err := decoder.Decode(&submission)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("with stdin: %s\n", submission.Stdins[0])
	// send to compilebox
	result := execCodeSubmission(submission)

	// encode result
	buf, _ := json.MarshalIndent(result, "", "   ")

	// write result to client
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
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

func submitToChallenge(w http.ResponseWriter, r *http.Request) {
	log.Print("Received code submission for challenge...")
	decoder := json.NewDecoder(r.Body)
	var submission CodeSubmission
	err := decoder.Decode(&submission)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	log.Printf("Code targets challenge %s", submission.ChallengeId)
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Environment variable %s not found, setting to %s\n", key, fallback)
	os.Setenv(key, fallback)
	return fallback
}
