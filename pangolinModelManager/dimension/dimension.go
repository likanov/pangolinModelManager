package dimension

import (
	"encoding/json"
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
)

func GetAllDimensions(session security.UserSession, pangolin string) []Dimension {
	resource := "/api/v1/Dimension/list?pageNumber=0&pageSize=1000"
	resp, resBody := restClient.SendGetRequest(session, pangolin, resource)
	if resp.StatusCode == 200 {
		var getAllDimensionsResponse GetAllDimensionsResponse
		err := json.Unmarshal(resBody, &getAllDimensionsResponse)
		if err != nil {
			return nil
		}
		var dimensions []Dimension
		for _, propGroupBuffer := range getAllDimensionsResponse.Content {
			dimensions = append(dimensions, Dimension{ParentId: propGroupBuffer.ParentId, Name: propGroupBuffer.Name, Description: propGroupBuffer.Description, Params: propGroupBuffer.Params, Id: propGroupBuffer.Id})
		}
		return dimensions
	} else {
		fmt.Println("Error during Dimension list retrieval")
		fmt.Println(string(resBody))
		return nil
	}
}

type Dimension struct {
	ParentId    string      `json:"parentId"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      interface{} `json:"params"`
	Id          string      `json:"id"`
}

type GetAllDimensionsResponse struct {
	Content []struct {
		ParentId    string      `json:"parentId"`
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
