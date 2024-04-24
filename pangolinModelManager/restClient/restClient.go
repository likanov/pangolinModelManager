package restClient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pangolinModelManager/security"
	"strings"
)

func SendPostRequest(et interface{}, userSession security.UserSession, pangolinUIUrl string, resource string) (*http.Response, []byte) {
	marshal, err := json.Marshal(et)
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, pangolinUIUrl+resource, strings.NewReader(string(marshal)))
	if err != nil {
		fmt.Println(err)
	}

	return getHTTPResponse(req, userSession, err, client)
}

func SendPatchRequest(et interface{}, userSession security.UserSession, pangolinUIUrl string, resource string) (*http.Response, []byte) {
	marshal, err := json.Marshal(et)
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPatch, pangolinUIUrl+resource, strings.NewReader(string(marshal)))
	if err != nil {
		fmt.Println(err)
	}

	return getHTTPResponse(req, userSession, err, client)
}

func SendGetRequest(userSession security.UserSession, pangolinUIUrl string, resource string) (*http.Response, []byte) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pangolinUIUrl+resource, nil)
	if err != nil {
		fmt.Println(err)
	}

	return getHTTPResponse(req, userSession, err, client)
}

func getHTTPResponse(req *http.Request, userSession security.UserSession, err error, client *http.Client) (*http.Response, []byte) {

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+userSession.Token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return resp, resBody
}
