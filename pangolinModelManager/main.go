package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	spreadsheetID = "1ANHnLYMldOaWvGcdUrLLEEngxAVPPLDzZ__q0n5HeF8"
	readRange     = "Sheet1!A:C"
	credentials   = "golang-api-419608-80318434846a.json"
)

var sheetsService *sheets.Service

func main() {
	// Load the Google Sheets API credentials from your JSON file.
	creds, err := ioutil.ReadFile(credentials)
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(creds, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to create JWT config: %v", err)
	}

	client := config.Client(context.Background())
	sheetsService, err = sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Google Sheets service: %v", err)
	}

	ctx := context.Background()

	//connectionToPangolinParams := getConnectionToPangolinParams(spreadsheetID, ctx)
	//userSession := getToken(connectionToPangolinParams)

	getAllSheets(spreadsheetID, ctx)

	//propGroupList := getAllPropGroupList(userSession, connectionToPangolinParams.NfviPangolin)
	//fmt.Println(propGroupList)
	//
	//allEntityTypes := getAllEntityTypes(userSession, connectionToPangolinParams.NfviPangolin)
	//fmt.Println(allEntityTypes)
	//
	//types := getAllPropTypes(userSession, connectionToPangolinParams.NfviPangolin)
	//fmt.Println(types)
	//
	//dimensions := getAllDimensions(userSession, connectionToPangolinParams.NfviPangolin)
	//fmt.Println(dimensions)

	//propGroup := PropGroup{Name: "TestGroup123222", Description: "test import v1", Params: nil}
	//createGroup(userSession, connectionToPangolinParams.NfviPangolin, &propGroup)

	//entityType := EntityType{ParentId: "", Name: "TestType123222", Description: "test import v1", Params: nil}
	//doCreateEntityType(userSession, connectionToPangolinParams.NfviPangolin, &entityType)
	//fmt.Println(entityType)

	//props := getAllProps(userSession, connectionToPangolinParams.NfviPangolin)
	//fmt.Println(props)

	//propRequest := PropRequest{Name: "TestProp12322222", Description: "test import v1", Params: nil, PropGroupId: "b977e79b-166d-422d-84f3-de431599235c", PropTypeId: "00010000-0000-4000-8000-000000000000"}
	//doCreateProp(userSession, connectionToPangolinParams.NfviPangolin, propRequest)

	//prop := getPropDetailsById(userSession, connectionToPangolinParams.NfviPangolin, "147afa7a-c8dc-4def-ab00-3f66f34ad74a")
	//fmt.Println(prop)

	//var dimension Dimension
	//var entityType EntityType
	//var prop Prop
	//
	//doCreateEntityTypeProp(userSession, connectionToPangolinParams.NfviPangolin, dimension, entityType, prop)

}

func getAllSheets(id string, ctx context.Context) []SheetData {
	respSpreadSheetData, _ := sheetsService.Spreadsheets.Get(id).Context(ctx).Do()
	sheetList := respSpreadSheetData.Sheets

	var sheetData []SheetData
	for key, sheet := range sheetList {
		if key == 0 {
			continue
		}

		sheetRange := sheet.Properties.Title + "!" + "A:H"
		resp2, err := sheetsService.Spreadsheets.Values.Get(id, sheetRange).Context(ctx).Do()
		if err != nil {
			fmt.Println(err)
		}

		valuesMap := make(map[string]string)
		for key, value := range resp2.Values {
			if len(value) > 1 {
				valuesMap[value[0].(string)] = value[1].(string)
			} else {
				valuesMap[value[0].(string)] = ""
			}
			if key == 4 {
				break
			}
		}

		var SheetData SheetData
		SheetData.EntityTypeSheet = EntityTypeSheet{valuesMap["ParentId"], valuesMap["Name"], valuesMap["Description"], valuesMap["Params"], valuesMap["Id"]}

		var propSheets = []PropSheet{}
		for key, value := range resp2.Values {
			if key <= 6 {
				continue
			}

			var PropListValueItems = []PropListValue{}
			if value[5] != nil && value[5] != "" {
				split := strings.Split(value[5].(string), ",")
				for _, v := range split {
					PropListValueItems = append(PropListValueItems, PropListValue{Name: v[1 : len(v)-1]})
				}
			}

			PropSheet := PropSheet{
				Name:        value[0].(string),
				Description: value[1].(string),
				PropTypeId:  value[2].(string),
				GroupName:   value[3].(string),
				Params:      value[4].(string),
				ListValues:  PropListValueItems,
				Dimension:   value[7].(string),
			}

			propSheets = append(propSheets, PropSheet)
		}
		SheetData.PropSheetItems = propSheets

		sheetData = append(sheetData, SheetData)

	}
	return sheetData
}

func doCreatePropListValue(session UserSession, pangolin string, propListValue PropListValue) {
	resource := "/api/v1/PropListValue"
	resp, resBody := sendPostRequest(propListValue, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("PropListValue created successfully :", string(resBody))
	} else {
		fmt.Println("Error during PropListValue creation")
		fmt.Println(string(resBody))
	}
}

func doCreateEntityTypeProp(session UserSession, pangolin string, dimension Dimension, entityType EntityType, prop Prop) {
	resource := "/api/v1/EntityTypeProp"
	values := map[string]string{"entityTypeId": entityType.Id, "propId": prop.Id, "dimensionId": dimension.Id}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(pangolin+resource, "application/json", bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == 201 {
		fmt.Println("EntityTypeProp created successfully")
	} else {
		var errres map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errres)
		fmt.Printf("error during EntityTypeProp creation %v %v \n", resp.StatusCode, errres)
		panic(resp.Body)
	}
}

func getAllDimensions(session UserSession, pangolin string) []Dimension {
	resource := "/api/v1/Dimension/list?pageNumber=0&pageSize=1000"
	resp, resBody := sendGetRequest(session, pangolin, resource)
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

func getPropDetailsById(session UserSession, pangolin string, propId string) Prop {
	resource := "/api/v1/Prop/" + propId
	resp, resBody := sendGetRequest(session, pangolin, resource)
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

func doCreateProp(session UserSession, pangolin string, request PropRequest) Prop {
	resource := "/api/v1/Prop"
	resp, resBody := sendPostRequest(request, session, pangolin, resource)
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

func getAllProps(session UserSession, pangolin string) []Prop {
	resource := "/api/v1/Prop/list?pageNumber=0&pageSize=1000"
	resp, resBody := sendGetRequest(session, pangolin, resource)
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

func getAllPropTypes(session UserSession, pangolin string) []PropType {
	resource := "/api/v1/PropType/list?pageNumber=0&pageSize=10000"
	resp, resBody := sendGetRequest(session, pangolin, resource)
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

func doCreateEntityType(session UserSession, pangolin string, e *EntityType) {
	resource := "/api/v1/EntityType"
	resp, resBody := sendPostRequest(e, session, pangolin, resource)
	if resp.StatusCode == 201 {
		fmt.Println("Type created successfully :", string(resBody))
		e.Id = string(resBody)
	} else {
		fmt.Println("Error during Type creation")
		fmt.Println(string(resBody))
		fmt.Errorf("Error during Type creation")
		panic(string(resBody))
	}
}

func getAllEntityTypes(userSession UserSession, pangolinUIUrl string) []EntityType {
	resource := "/api/v1/EntityType/list?pageNumber=0&pageSize=10000"
	resp, resBody := sendGetRequest(userSession, pangolinUIUrl, resource)

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

func doCreatePropGroup(session UserSession, pangolin string, group *PropGroup) {
	resource := "/api/v1/PropGroup"
	resp, resBody := sendPostRequest(group, session, pangolin, resource)
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

func getAllPropGroupList(userSession UserSession, pangolinUIUrl string) []PropGroup {
	resource := "/api/v1/PropGroup/list?pageNumber=0&pageSize=10000"
	resp, resBody := sendGetRequest(userSession, pangolinUIUrl, resource)

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

func getConnectionToPangolinParams(id string, ctx context.Context) PangolinParams {
	respSpreadSheetData, _ := sheetsService.Spreadsheets.Get(id).Context(ctx).Do()
	sheetList := respSpreadSheetData.Sheets

	sheetRange := sheetList[0].Properties.Title + "!" + "A:H"
	resp2, err := sheetsService.Spreadsheets.Values.Get(id, sheetRange).Context(ctx).Do()
	if err != nil {
		fmt.Println(err)
	}

	valuesMap := make(map[string]string)
	for _, value := range resp2.Values {
		valuesMap[value[0].(string)] = value[1].(string)
	}

	const LoginKey = "login"
	const PassKey = "pass"
	const PangolinServiceKey = "nfvi-pangolin"
	const PangolinSecurityKey = "nvfi-pangolin-security"

	return PangolinParams{valuesMap[LoginKey], valuesMap[PassKey], valuesMap[PangolinServiceKey], valuesMap[PangolinSecurityKey]}
}

func getToken(params PangolinParams) UserSession {
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

func ReadData(w http.ResponseWriter, r *http.Request) {
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(r.Context()).Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respSpreadSheetData, _ := sheetsService.Spreadsheets.Get(spreadsheetID).Context(r.Context()).Do()
	sheetList := respSpreadSheetData.Sheets

	for _, sheet := range sheetList {
		fmt.Println(sheet.Properties.Title)
		sheetRange := sheet.Properties.Title + "!" + "A:G"
		resp2, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, sheetRange).Context(r.Context()).Do()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data2, _ := json.Marshal(resp2.Values)
		fmt.Println(string(data2))
	}

	data, _ := json.Marshal(resp.Values)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func sendPostRequest(et interface{}, userSession UserSession, pangolinUIUrl string, resource string) (*http.Response, []byte) {
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

func sendGetRequest(userSession UserSession, pangolinUIUrl string, resource string) (*http.Response, []byte) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pangolinUIUrl+resource, nil)
	if err != nil {
		fmt.Println(err)
	}

	return getHTTPResponse(req, userSession, err, client)
}

func getHTTPResponse(req *http.Request, userSession UserSession, err error, client *http.Client) (*http.Response, []byte) {

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

type EntityType struct {
	ParentId    string      `json:"parentId,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	Id          string      `json:"id,omitempty"`
}

type EntityTypeSheet struct {
	ParentId    string
	Name        string
	Description string
	Params      interface{}
	Id          string
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

type PropSheet struct {
	Name        string
	PropGroupId string
	PropTypeId  string
	Description string
	Params      interface{}
	GroupName   string
	ListValues  []PropListValue
	Dimension   string
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

type PropListValue struct {
	Name        string      `json:"name,omitempty"`
	PropId      string      `json:"propId,omitempty"`
	Description string      `json:"description,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	Id          string      `json:"id"`
}

type SheetData struct {
	EntityTypeSheet EntityTypeSheet
	PropSheetItems  []PropSheet
}
