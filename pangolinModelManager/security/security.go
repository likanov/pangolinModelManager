package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func GetToken(params PangolinParams) UserSession {
	resource := "/api/v1/security/auth/login"
	values := map[string]string{"userLogin": params.Login, "userPass": params.Password}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(params.NvfiPangolinSecurity+resource, "application/json", bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var res UserSession
		json.NewDecoder(resp.Body).Decode(&res)
		return res

	} else {
		var errres map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errres)
		fmt.Printf("error during get token execution %v %v \n", resp.StatusCode, errres)
		panic(resp.Body)
	}
}

type UserSession struct {
	Login       string `json:"login"`
	Token       string `json:"token"`
	LandingPage string `json:"landingPage"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type PangolinParams struct {
	Login                string
	Password             string
	NfviPangolin         string
	NvfiPangolinSecurity string
}
