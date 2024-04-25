package datametrics

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func processData(requestData RequestData, url string) []byte {

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Fatal("Error marshalling JSON:", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Error creating HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request")
	}

	// Read the HTTP response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading HTTP response")
	}

	resp.Body.Close()

	return body
}
