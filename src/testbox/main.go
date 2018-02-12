package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// type languageDetail struct {
// 	Boilerplate   string `json:"boilerplate"`
// 	CommentPrefix string `json:"commentPrefix"`
// }
type languageDetail struct {
	// Compiler           string `json:"compiler"`
	// SourceFile         string `json:"sourceFile"`
	// OptionalExecutable string `json:"optionalExecutable"`
	// CompilerFlags      string `json:"compilerFlags"`
	Boilerplate   string `json:"boilerplate"`
	CommentPrefix string `json:"commentPrefix"`
	// Disabled           string `json:"disabled"`
}

var languages map[string]languageDetail
var cbPort, cbAddress string

func main() {
	// Read settings. Compilebox address/port and this address/port
	var portOk, addressOk bool
	cbPort, portOk = os.LookupEnv("COMPILEBOX_PORT")
	cbAddress, addressOk = os.LookupEnv("COMPILEBOX_ADDRESS")
	if !portOk || !addressOk {
		log.Fatal("Missing compilebox environment variables, please make sure service is available")
	}

	// Check to ensure the compilebox is up by trying to fill langs variable
	fmt.Printf("Requesting language list from compilebox (%s:%s)...\n", cbAddress, cbPort)
	populateLanguages()

	/* Serve REST API endpoints for

	- various challenge searches

	- simple code submission (just pass along to testbox)

	- code submission checked against challenge

	*/

	port := getEnv("TESTBOX_PORT", "31336")

	// http.HandleFunc("/get_challenge/", getChallenge)
	// http.HandleFunc("/submit/", submitTest)
	// http.HandleFunc("/stdout/", getStdout)
	// http.HandleFunc("/languages/", getLangs)

	fmt.Println("testbox listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	/* Serve administrative interface for challenge collection

	 */
}

func cbPath(path string) string {
	return cbAddress + ":" + cbPort + path
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	fmt.Printf("Environment variable %s not found, setting to %s\n", key, fallback)
	os.Setenv(key, fallback)
	return fallback
}

func populateLanguages() {
	r, err := http.Get(cbPath("/languages/"))

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
