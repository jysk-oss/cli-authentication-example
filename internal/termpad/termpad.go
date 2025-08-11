package termpad

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

var termpadUrl = "https://termpad.your-domain.com"

// Post a string to Termpad
func PostData(token string, input string) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", termpadUrl, bytes.NewBufferString(input))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(body))
}

// Get a string from Termpad by identifier.
func GetData(token string, input string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", termpadUrl+"/raw/"+input, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(body))
}
