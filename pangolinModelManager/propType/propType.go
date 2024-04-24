package propType

import (
	"encoding/json"
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
)

func GetAllPropTypes(session security.UserSession, pangolin string) []PropType {
	resource := "/api/v1/PropType/list?pageNumber=0&pageSize=10000"
	resp, resBody := restClient.SendGetRequest(session, pangolin, resource)
	if resp.StatusCode == 200 {
		var getAllPropTypesResponse GetAllPropTypesResponse
		err := json.Unmarshal(resBody, &getAllPropTypesResponse)
		if err != nil {
			return nil
		}
		var types []PropType
		for _, propGroupBuffer := range getAllPropTypesResponse.Content {
			types = append(types, PropType{Name: propGroupBuffer.Name, Description: propGroupBuffer.Description, Params: propGroupBuffer.Params, Id: propGroupBuffer.Id})
		}
		return types
	} else {
		fmt.Println("Error during PropType list retrieval")
		fmt.Println(string(resBody))
		return nil
	}
}

type PropType struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      interface{} `json:"params"`
	Id          string      `json:"id"`
}

type GetAllPropTypesResponse struct {
	Content []struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Params      interface{} `json:"params"`
		Id          string      `json:"id"`
	} `json:"content"`
	Pageable struct {
		Sort struct {
			Sorted   bool `json:"sorted"`
			Unsorted bool `json:"unsorted"`
			Empty    bool `json:"empty"`
		} `json:"sort"`
		PageSize   int  `json:"pageSize"`
		PageNumber int  `json:"pageNumber"`
		Offset     int  `json:"offset"`
		Unpaged    bool `json:"unpaged"`
		Paged      bool `json:"paged"`
	} `json:"pageable"`
	TotalPages       int  `json:"totalPages"`
	TotalElements    int  `json:"totalElements"`
	Last             bool `json:"last"`
	NumberOfElements int  `json:"numberOfElements"`
	First            bool `json:"first"`
	Size             int  `json:"size"`
	Number           int  `json:"number"`
	Sort             struct {
		Sorted   bool `json:"sorted"`
		Unsorted bool `json:"unsorted"`
		Empty    bool `json:"empty"`
	} `json:"sort"`
	Empty bool `json:"empty"`
}
