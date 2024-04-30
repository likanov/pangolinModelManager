package propListValue

import (
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
)

func DoCreatePropListValue(session security.UserSession, pangolin string, propListValue *PropListValue) {
	resource := "/api/v1/PropListValue"
	resp, resBody := restClient.SendPostRequest(propListValue, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("PropListValue created successfully :", string(resBody))
	} else {
		fmt.Println("Error during PropListValue creation")
		fmt.Println(string(resBody))
	}
}

type PropListValue struct {
	Name        string      `json:"name,omitempty"`
	PropId      string      `json:"propId,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	Id          string      `json:"id"`
}
