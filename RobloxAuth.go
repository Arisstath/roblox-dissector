package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

// GetAuthTicket provides a ticket required for joining games in peer/Client.go
func GetAuthTicket(username string, password string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	authTicketClient := &http.Client{Jar: jar}
	csrfRequest, err := http.NewRequest("POST", "https://auth.roblox.com", nil)

	resp, err := authTicketClient.Do(csrfRequest)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var csrfToken string
	if resp.StatusCode == 403 {
		csrfToken = resp.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			return "", errors.New("empty CSRF token while trying to authenticate")
		}
	} else {
		return "", fmt.Errorf("wrong status code \"%d %s\" when trying to authenticate", resp.StatusCode, resp.Status)
	}

	authCredentials, err := json.Marshal(map[string]string{
		"ctype":    "Username",
		"cvalue":   username,
		"password": password,
	})
	if err != nil {
		return "", errors.New("failed to construct JSON credentials")
	}

	loginRequest, err := http.NewRequest("POST", "https://auth.roblox.com/v2/login", bytes.NewReader(authCredentials))
	if err != nil {
		return "", err
	}
	loginRequest.Header.Set("X-CSRF-Token", csrfToken)
	loginRequest.Header.Set("Content-Type", "application/json")

	resp, err = authTicketClient.Do(loginRequest)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		jsonResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()
		println("failed to auth: ", string(jsonResp))
		return "", errors.New("failed to authenticate")
	}

	ticketRequest, err := http.NewRequest("GET", "https://www.roblox.com/game-auth/getauthticket", nil)
	if err != nil {
		return "", err
	}
	ticketRequest.Header.Set("RBX-For-Gameauth", "true")
	ticketRequest.Header.Set("Referer", "https://www.roblox.com/home")

	resp, err = authTicketClient.Do(ticketRequest)
	if err != nil {
		return "", err
	}
	ticketResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		println("failed to get authticket: ", resp.Status, string(ticketResp))
		return "", errors.New("failed to authenticate")
	}

	return string(ticketResp), nil
}
