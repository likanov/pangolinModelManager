package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"pangolinModelManager/dimension"
	"pangolinModelManager/entityType"
	"pangolinModelManager/propGroup"
	"pangolinModelManager/propListValue"
	"pangolinModelManager/propType"
	"pangolinModelManager/security"
	"pangolinModelManager/stringUtils"
	"slices"

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

	connectionToPangolinParams := getConnectionToPangolinParams(spreadsheetID, ctx)
	userSession := security.GetToken(connectionToPangolinParams)

	dimensions := dimension.GetAllDimensions(userSession, connectionToPangolinParams.NfviPangolin)
	fmt.Println(dimensions)

	list := propGroup.GetAllPropGroupList(userSession, connectionToPangolinParams.NfviPangolin)
	fmt.Println(list)

	types := entityType.GetAllEntityTypes(userSession, connectionToPangolinParams.NfviPangolin)
	fmt.Println(types)

	propTypes := propType.GetAllPropTypes(userSession, connectionToPangolinParams.NfviPangolin)
	fmt.Println(propTypes)

	allSheets := getAllSheets(spreadsheetID, ctx)

	for index, sheet := range allSheets {
		if sheet.EntityTypeSheet.ParentId != "" {
			idx := slices.IndexFunc(types, func(c entityType.EntityType) bool { return c.Name == sheet.EntityTypeSheet.ParentId })
			if idx != -1 {
				allSheets[index].EntityTypeSheet.ParentId = types[idx].Id
			} else {
				log.Fatalf("Parent type not found: %s", sheet.EntityTypeSheet.ParentId)
			}
		}
	}

	//check if id is not empty
	for index, sheet := range allSheets {
		if sheet.EntityTypeSheet.Id == "" {
			typeSheet := sheet.EntityTypeSheet
			EntityType := entityType.EntityType{ParentId: typeSheet.ParentId, Name: typeSheet.Name, Description: typeSheet.Description, Params: typeSheet.Params}
			entityType.DoCreateEntityType(userSession, connectionToPangolinParams.NfviPangolin, &EntityType)
			allSheets[index].EntityTypeSheet.Id = sheet.EntityTypeSheet.Id
		} else if stringUtils.ISUUID(sheet.EntityTypeSheet.Id) {
			//update entity type
			typeSheet := sheet.EntityTypeSheet
			EntityType := entityType.EntityType{ParentId: typeSheet.ParentId, Name: typeSheet.Name, Description: typeSheet.Description, Params: typeSheet.Params, Id: typeSheet.Id}
			updateEntityType := entityType.DoUpdateEntityType(userSession, connectionToPangolinParams.NfviPangolin, &EntityType)
			fmt.Println(updateEntityType)
		}

	}

	fmt.Println(allSheets)

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
		SheetData.EntityTypeSheet = EntityTypeSheet{valuesMap["Parent"], valuesMap["Name"], valuesMap["Description"], valuesMap["Params"], valuesMap["Id"]}

		var propSheets = []PropSheet{}
		for key, value := range resp2.Values {
			if key <= 6 {
				continue
			}

			var PropListValueItems = []propListValue.PropListValue{}
			if value[5] != nil && value[5] != "" {
				split := strings.Split(value[5].(string), ",")
				for _, v := range split {
					PropListValueItems = append(PropListValueItems, propListValue.PropListValue{Name: v[1 : len(v)-1]})
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

func getConnectionToPangolinParams(id string, ctx context.Context) security.PangolinParams {
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

	return security.PangolinParams{valuesMap[LoginKey], valuesMap[PassKey], valuesMap[PangolinServiceKey], valuesMap[PangolinSecurityKey]}
}

type EntityTypeSheet struct {
	ParentId    string
	Name        string
	Description string
	Params      interface{}
	Id          string
}

type PropSheet struct {
	Name        string
	PropGroupId string
	PropTypeId  string
	Description string
	Params      interface{}
	GroupName   string
	ListValues  []propListValue.PropListValue
	Dimension   string
}

type SheetData struct {
	EntityTypeSheet EntityTypeSheet
	PropSheetItems  []PropSheet
}
