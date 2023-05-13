package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Init the API Key from the environment variables
var ASSEMBLY_AI_API_KEY = os.Getenv("ASSEMBLY_AI_API_KEY")

// Make a post request to the AssemblyAI API and retrieve the id of the transcript
func requestTranscript(url string, client http.Client) (string, error) {
	values := map[string]string{"audio_url": url}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.assemblyai.com/v2/transcript", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", ASSEMBLY_AI_API_KEY)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//Decode the response
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}
	if data["error"] != nil {
		return "", fmt.Errorf(data["error"].(string))
	}
	//Get the id of the transcript
	id := data["id"].(string)
	return id, nil
}

func getTranscript(id string, client http.Client) (string, error) {
	timeout := time.After(20 * time.Second)
	for time.Now().Before(<-timeout) {
		//Make a get request to the AssemblyAI API and retrieve the transcript
		req, err := http.NewRequest("GET", "https://api.assemblyai.com/v2/transcript/"+id, nil)
		if err != nil {
			log.Println("Error creating request:", err)
			return "", err
		}
		req.Header.Set("authorization", ASSEMBLY_AI_API_KEY)
		req.Header.Set("content-type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending request:", err)
			return "", err
		}
		defer resp.Body.Close()
		//Decode the response
		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Println("Error decoding response:", err)
			return "", err
		}
		if data["status"] == "completed" {
			return data["text"].(string), nil
		}
		time.Sleep(1 * time.Second)
	}
	return "", fmt.Errorf("Timeout")
}

func ToText(url string) (string, error) {
	client := &http.Client{}
	id, err := requestTranscript(url, *client)
	if err != nil {
		log.Println("Error requesting transcript:", err)
		return "", err
	}
	text, err := getTranscript(id, *client)
	if err != nil {
		log.Println("Error getting transcript:", err)
		return "", err
	}

	return text, nil
}
