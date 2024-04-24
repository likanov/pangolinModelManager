package prop

import (
	"encoding/json"
	"fmt"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
	"strings"
)

func getPropDetailsById(session security.UserSession, pangolin string, propId string) Prop {
	resource := "/api/v1/Prop/" + propId
	resp, resBody := restClient.SendGetRequest(session, pangolin, resource)
	if resp.StatusCode == 200 {
		var prop Prop
		err := json.Unmarshal(resBody, &prop)
		if err != nil {
			return Prop{}
		}
		return prop
	} else {
		fmt.Println("Error during Prop retrieval")
		fmt.Println(string(resBody))
		return Prop{}
	}
}

func DoCreateProp(session security.UserSession, pangolin string, request PropRequest) Prop {
	resource := "/api/v1/Prop"
	resp, resBody := restClient.SendPostRequest(request, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("Prop created successfully :", string(resBody))
		propId := strings.ReplaceAll(string(resBody), "\"", "")
		return getPropDetailsById(session, pangolin, propId)
	} else {
		fmt.Println("Error during Prop creation")
		fmt.Println(string(resBody))
		return Prop{}
	}
}

func GetAllProps(session security.UserSession, pangolin string) []Prop {
	resource := "/api/v1/Prop/list?pageNumber=0&pageSize=1000"
	resp, resBody := restClient.SendGetRequest(session, pangolin, resource)
	if resp.StatusCode == 200 {
		var getAllPropsResponse GetAllPropsResponse
		err := json.Unmarshal(resBody, &getAllPropsResponse)
		if err != nil {
			return nil
		}
		var props []Prop
		for _, propGroupBuffer := range getAllPropsResponse.Content {
			props = append(props, Prop{Name: propGroupBuffer.Name, Description: propGroupBuffer.Description, Params: propGroupBuffer.Params, Id: propGroupBuffer.Id})
		}
		return props
	} else {
		fmt.Println("Error during Prop list retrieval")
		fmt.Println(string(resBody))
		return nil
	}
}

type Prop struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Params      *struct {
		ParamClass         string      `json:"param_class,omitempty"`
		CalcState          string      `json:"calc_state,omitempty"`
		Query              string      `json:"query,omitempty"`
		IntervalHours      interface{} `json:"interval_hours,omitempty"`
		ExternalColumnName string      `json:"external_column_name,omitempty"`
		View               string      `json:"view,omitempty"`
		Filters            []struct {
			EntityTypeId string `json:"entity_type_id"`
		} `json:"filters,omitempty"`
		Key            int    `json:"key,omitempty"`
		TimestampMask  string `json:"timestamp_mask,omitempty"`
		HideTimePicker bool   `json:"hide_time_picker,omitempty"`
		IntervalDays   string `json:"interval_days,omitempty"`
		Mode           string `json:"mode,omitempty"`
	} `json:"params"`
	PropTypeId  *string `json:"propTypeId"`
	PropGroupId *string `json:"propGroupId"`
	Id          string  `json:"id"`
	PropType    *struct {
		Name        string      `json:"name"`
		Description *string     `json:"description"`
		Params      interface{} `json:"params"`
		Id          string      `json:"id"`
	} `json:"propType"`
	PropGroup *struct {
		Name        string      `json:"name"`
		Description *string     `json:"description"`
		Params      interface{} `json:"params"`
		Id          string      `json:"id"`
	} `json:"propGroup"`
}

type GetAllPropsResponse struct {
	Content []struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Params      *struct {
			ParamClass         string      `json:"param_class,omitempty"`
			CalcState          string      `json:"calc_state,omitempty"`
			Query              string      `json:"query,omitempty"`
			IntervalHours      interface{} `json:"interval_hours,omitempty"`
			ExternalColumnName string      `json:"external_column_name,omitempty"`
			View               string      `json:"view,omitempty"`
			Filters            []struct {
				EntityTypeId string `json:"entity_type_id"`
			} `json:"filters,omitempty"`
			Key            int    `json:"key,omitempty"`
			TimestampMask  string `json:"timestamp_mask,omitempty"`
			HideTimePicker bool   `json:"hide_time_picker,omitempty"`
			IntervalDays   string `json:"interval_days,omitempty"`
			Mode           string `json:"mode,omitempty"`
		} `json:"params"`
		PropTypeId  *string `json:"propTypeId"`
		PropGroupId *string `json:"propGroupId"`
		Id          string  `json:"id"`
		PropType    *struct {
			Name        string      `json:"name"`
			Description *string     `json:"description"`
			Params      interface{} `json:"params"`
			Id          string      `json:"id"`
		} `json:"propType"`
		PropGroup *struct {
			Name        string      `json:"name"`
			Description *string     `json:"description"`
			Params      interface{} `json:"params"`
			Id          string      `json:"id"`
		} `json:"propGroup"`
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

type PropRequest struct {
	Name        string      `json:"name,omitempty"`
	PropGroupId string      `json:"propGroupId,omitempty"`
	PropTypeId  string      `json:"propTypeId,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
}
