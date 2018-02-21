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

// CodeSubmission describes a snippet of executable code, the language it's in, and any stdins that should be piped in.
// CodeSubmission optionally contains the id of the challenge it should be tested against
type CodeSubmission struct {
	Language    string        `json:"language"`
	Code        string        `json:"code"`
	Stdins      []string      `json:"stdins"`
	ChallengeId challenges.ID `json:"challengeId,omitempty`
}

// ExecutionResult contains the outputs (stdouts) of submitted code, as well as an compiler messages (Message).
// ExecutionResult may contain a grade for the code if the code was tested against a challenge
type ExecutionResult struct {
	Stdouts []string          `json:"stdouts"`
	Graded  map[string]string `json:"graded,omitempty"`
	Message message           `json:"message"`
}

type message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type apiResponse struct {
	ErrorMessage string        `json:"error,omitempty"`
	ID           challenges.ID `json:"id,omitempty"`
	Result       interface{}   `json:"result,omitempty"`
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
	languages = fetchLanguages()

	// routes for admin of challenge db
	http.HandleFunc("/challenges/id/", getChallenge)
	http.HandleFunc("/challenges/insert/", insertChallenge)
	http.HandleFunc("/challenges/delete/", deleteChallenge)
	http.HandleFunc("/challenges/update/", updateChallenge)

	// routes for requesting challenges
	http.HandleFunc("/challenges/rand/", getRandChallenge)
	http.HandleFunc("/challenges/all/", getAllChallenges)
	// http.HandleFunc("/challenges/search/", searchChallenges)

	// get list of supported languages:
	http.HandleFunc("/languages/", getLangs)

	// routes for submitting code
	http.HandleFunc("/submit/", submitToChallenge)
	http.HandleFunc("/stdout/", getStdout)

	// front page, should have admin login and project info
	// http.HandleFunc("/", frontPage)

	port = getEnv("TESTBOX_PORT", "31336")
	fmt.Println("testbox listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// TODO standardize error messaging
func getChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")

	var id challenges.ID
	decodeSubmission(r, &id)

	fmt.Printf("Challenge id: %s\n", id)

	c, err := challenges.GetById(id)

	resp := apiResponse{Result: c}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	buf, _ := json.MarshalIndent(resp, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
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

	buf, _ := json.MarshalIndent(resp, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func updateChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")

	var c challenges.Challenge
	decodeSubmission(r, &c)

	fmt.Printf("Challenge struct: %v\n", c)

	// TODO this is broken
	err := challenges.Update(c.ID, &c)
	if err != nil {
		// TODO this should get passed to client.
		panic(err)
	}

	temp := "How to handle responses?"

	buf, _ := json.MarshalIndent(temp, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func deleteChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received insert request")

	var id challenges.ID
	decodeSubmission(r, &id)

	// fmt.Printf("Challenge struct: %v\n", c)

	err := challenges.Delete(id)
	if err != nil {
		// TODO this should get passed to client.
		panic(err)
	}

	temp := "How to handle responses?"

	buf, _ := json.MarshalIndent(temp, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getRandChallenge(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetRandom()

	buf, _ := json.MarshalIndent(c, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getAllChallenges(w http.ResponseWriter, r *http.Request) {
	c := challenges.GetAll()

	buf, _ := json.MarshalIndent(c, "", "   ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getLangs(w http.ResponseWriter, r *http.Request) {
	log.Println("Received languages list request")

	buf, _ := json.MarshalIndent(languages, "", "   ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func getStdout(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Received stdout-only code submission...")

	// decode submission

	var submission CodeSubmission
	decodeSubmission(r, &submission)

	fmt.Printf("with stdin: %s\n", submission.Stdins[0])

	// send to compilebox
	result := execCodeSubmission(submission)

	// encode and write result back to client
	jsonWrite(result, w)
}

func submitToChallenge(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Received code submission for challenge...")

	var submission CodeSubmission
	decodeSubmission(r, &submission)

	fmt.Printf("Code targets challenge %s", submission.ChallengeId)
	fmt.Println(submission)

	// attach the appropriate challenge's stdins to the submission
	c, err := challenges.GetById(submission.ChallengeId)
	if err != nil {
		panic(err) // TODO handle error more gracefully and pass along to user
	}

	stdins, expectedStdouts := c.GetIOSplit()
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
