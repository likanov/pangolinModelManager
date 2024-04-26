package entityType

import (
	"encoding/json"
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
	"strings"
)

func DoCreateEntityType(session security.UserSession, pangolin string, e *EntityType) {
	resource := "/api/v1/EntityType"
	resp, resBody := restClient.SendPostRequest(e, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("Type created successfully :", string(resBody))
		s := string(resBody)
		e.Id = strings.ReplaceAll(s, "\"", "")
	} else {
		fmt.Println("Error during Type creation")
		fmt.Println(string(resBody))
	}
}

func DoUpdateEntityType(session security.UserSession, pangolin string, e *EntityType) EntityType {
	if e.Id == "" {
		fmt.Println("EntityType Id is required for update")
		panic("EntityType Id is required for update")
	}

	resource := "/api/v1/EntityType/" + e.Id

	resp, resBody := restClient.SendPatchRequest(e, session, pangolin, resource)
	if resp.StatusCode == 200 {
		fmt.Println("Type created successfully :", string(resBody))
		e.Id = string(resBody)

		var EntityType EntityType
		err := json.Unmarshal(resBody, &EntityType)
		if err != nil {
			fmt.Println("Error during Type update")
			panic(string(resBody))
		}
		return EntityType

	} else {
		fmt.Println("Error during Type creation")
		fmt.Println(string(resBody))
		panic(string(resBody))
	}
}

func GetAllEntityTypes(userSession security.UserSession, pangolinUIUrl string) []EntityType {
	resource := "/api/v1/EntityType/list?pageNumber=0&pageSize=10000"
	resp, resBody := restClient.SendGetRequest(userSession, pangolinUIUrl, resource)

	if resp.StatusCode == 200 {
		var getAllEntityTypesResponse GetAllEntityTypesResponse
		err := json.Unmarshal(resBody, &getAllEntityTypesResponse)
		if err != nil {
			return nil
		}
		var types []EntityType
		for _, propGroupBuffer := range getAllEntityTypesResponse.Content {
			types = append(types, EntityType{ParentId: propGroupBuffer.ParentId, Name: propGroupBuffer.Name, Description: propGroupBuffer.Description, Params: propGroupBuffer.Params, Id: propGroupBuffer.Id})
		}
		return types
	} else {
		fmt.Println("Error during PropGroup list retrieval")
		fmt.Println(string(resBody))
		return nil
	}
}

type EntityType struct {
	ParentId    string      `json:"parentId,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	Id          string      `json:"id,omitempty"`
}

type GetAllEntityTypesResponse struct {
	Content []struct {
		ParentId    string `json:"parentId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Params      struct {
			ExternalTableName   string `json:"external_table_name,omitempty"`
			View                string `json:"view,omitempty"`
			KeyColumn           string `json:"key_column,omitempty"`
			GetUrlParameters    string `json:"get_url_parameters,omitempty"`
			ExternalUrlEndpoint string `json:"external_url_endpoint,omitempty"`
			DeleteUrlParameters string `json:"delete_url_parameters,omitempty"`
		} `json:"params"`
		Id string `json:"id"`
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
