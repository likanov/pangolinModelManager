package propGroup

import (
	"encoding/json"
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
)

func DoCreatePropGroup(session security.UserSession, pangolin string, group *PropGroup) {
	resource := "/api/v1/PropGroup"
	resp, resBody := restClient.SendPostRequest(group, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("Group created successfully :", string(resBody))
		group.Id = string(resBody)
	} else {
		fmt.Println("Error during Group creation")
		fmt.Println(string(resBody))
		fmt.Errorf("Error during Group creation")
		panic(string(resBody))
	}
}

func GetAllPropGroupList(userSession security.UserSession, pangolinUIUrl string) []PropGroup {
	resource := "/api/v1/PropGroup/list?pageNumber=0&pageSize=10000"
	resp, resBody := restClient.SendGetRequest(userSession, pangolinUIUrl, resource)

	if resp.StatusCode == 200 {
		var getAllPropGroupsResponse AllPropGroupsResponse
		err := json.Unmarshal(resBody, &getAllPropGroupsResponse)
		if err != nil {
			return nil
		}
		var groups []PropGroup
		for _, propGroupBuffer := range getAllPropGroupsResponse.Content {
			groups = append(groups, PropGroup{Name: propGroupBuffer.Name, Description: propGroupBuffer.Description, Params: propGroupBuffer.Params, Id: propGroupBuffer.Id})
		}
		return groups
	} else {
		fmt.Println("Error during PropGroup list retrieval")
		fmt.Println(string(resBody))
		return nil
	}
}

type PropGroup struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	Id          string      `json:"id,omitempty"`
}

type AllPropGroupsResponse struct {
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
