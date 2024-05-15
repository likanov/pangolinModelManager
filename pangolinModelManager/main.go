package main

import (
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"pangolinModelManager/dimension"
	"pangolinModelManager/entityType"
	"pangolinModelManager/entityTypeProp"
	"pangolinModelManager/prop"
	"pangolinModelManager/propGroup"
	"pangolinModelManager/propListValue"
	"pangolinModelManager/propType"
	"pangolinModelManager/security"
	"pangolinModelManager/stringUtils"
	"slices"
	"strconv"

	"strings"
)

var (
	//go:embed golang-api-419608-80318434846a.json
	credentialsData []byte
)

var spreadsheetID string
var sheetsService *sheets.Service

var typeList []entityType.EntityType
var dimensionList []dimension.Dimension
var propGroupList []propGroup.PropGroup
var propTypesList []propType.PropType

var version string

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Fatalf("Please provide the command to run, use 'help' to get the list of commands")
	}
	if args[1] == "import" {
		if len(args) < 3 {
			log.Fatalf("Please provide the url to the google sheet")
		}
		buffer := args[2][:strings.LastIndex(args[2], "/")]
		spreadsheetID = buffer[strings.LastIndex(buffer, "/")+1:]
		fmt.Println(spreadsheetID)
		fmt.Println("Importing data from Google Sheets to Pangolin")
		doImport()
	} else if args[1] == "version" {
		fmt.Println("Version:1.2.0")
		fmt.Println("git info:", version)
	} else if args[1] == "help" {
		fmt.Println("import: Import the data from Google Sheets to Pangolin")
		fmt.Println("version: Get the version of the tool")
		fmt.Println("help: Get the help")
	} else {
		log.Fatalf("Command not found")
	}

}
func doImport() {
	//var f embed.FS
	//creds, err := f.ReadFile(credentials)
	//if err != nil {
	//	log.Fatalf("Unable to read credentials file: %v", err)
	//}

	config, err := google.JWTConfigFromJSON(credentialsData, sheets.SpreadsheetsScope)
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

	dimensionList = dimension.GetAllDimensions(userSession, connectionToPangolinParams.NfviPangolin)
	propGroupList = propGroup.GetAllPropGroupList(userSession, connectionToPangolinParams.NfviPangolin)
	typeList = entityType.GetAllEntityTypes(userSession, connectionToPangolinParams.NfviPangolin)
	propTypesList = propType.GetAllPropTypes(userSession, connectionToPangolinParams.NfviPangolin)

	allSheets := getAllSheets(spreadsheetID, ctx)

	//check if id is not empty
	for index, sheet := range allSheets {
		if sheet.EntityTypeSheet.Id == "" {

			typeSheet := sheet.EntityTypeSheet
			idx := slices.IndexFunc(typeList, func(c entityType.EntityType) bool { return c.Name == sheet.EntityTypeSheet.ParentId })
			if idx != -1 {
				typeSheet.ParentId = typeList[idx].Id
			} else {
				log.Fatalf("Parent type not found: %s", sheet.EntityTypeSheet.ParentId)
			}

			entityTypeInstance := getEntityTypeInstance(typeSheet)
			entityType.DoCreateEntityType(userSession, connectionToPangolinParams.NfviPangolin, &entityTypeInstance)
			allSheets[index].EntityTypeSheet.Id = entityTypeInstance.Id
			doUpdateCell(spreadsheetID, sheet.sheetTitle, 1, "B", entityTypeInstance.Id, ctx)
			typeList = entityType.GetAllEntityTypes(userSession, connectionToPangolinParams.NfviPangolin)

		} else if stringUtils.ISUUID(sheet.EntityTypeSheet.Id) {
			fmt.Println("Id is valid")

			typeSheet := sheet.EntityTypeSheet
			idx := slices.IndexFunc(typeList, func(c entityType.EntityType) bool { return c.Name == sheet.EntityTypeSheet.ParentId })
			if idx != -1 {
				typeSheet.ParentId = typeList[idx].Id
			} else {
				log.Fatalf("Parent type not found: %s", sheet.EntityTypeSheet.ParentId)
			}

			entityTypeInstance := getEntityTypeInstance(typeSheet)
			entityTypeInstance.Id = sheet.EntityTypeSheet.Id

			entityType.DoUpdateEntityType(userSession, connectionToPangolinParams.NfviPangolin, &entityTypeInstance)
			allSheets[index].EntityTypeSheet.Id = entityTypeInstance.Id
			typeList = entityType.GetAllEntityTypes(userSession, connectionToPangolinParams.NfviPangolin)
		}

	}

	for _, sheet := range allSheets {
		entityTypeId := sheet.EntityTypeSheet.Id
		fmt.Println("EntityTypeId: ", entityTypeId)

		for _, propSheet := range sheet.PropSheetItems {

			propGroupIdx := slices.IndexFunc(propGroupList, func(c propGroup.PropGroup) bool { return c.Name == propSheet.GroupName })
			if propGroupIdx != -1 {
				propSheet.PropGroupId = propGroupList[propGroupIdx].Id
			} else {
				propGroupInstance := propGroup.PropGroup{Name: propSheet.GroupName}
				propGroup.DoCreatePropGroup(userSession, connectionToPangolinParams.NfviPangolin, &propGroupInstance)
				propSheet.PropGroupId = propGroupInstance.Id
				propGroupList = propGroup.GetAllPropGroupList(userSession, connectionToPangolinParams.NfviPangolin)
			}

			propTypeIdx := slices.IndexFunc(propTypesList, func(c propType.PropType) bool { return c.Name == propSheet.PropTypeId })
			if propTypeIdx != -1 {
				propSheet.PropTypeId = propTypesList[propTypeIdx].Id
			} else {
				log.Fatal("PropType not found: ", propSheet.PropTypeId)
			}

			propRequestInstance := getPropRequestInstance(propSheet)

			var dimensionRequest = dimension.Dimension{}
			dimensionIdx := slices.IndexFunc(dimensionList, func(c dimension.Dimension) bool { return c.Name == propSheet.Dimension })
			if dimensionIdx != -1 {
				dimensionRequest = dimensionList[dimensionIdx]
			} else {
				log.Fatal("Dimension not found: ", propSheet.Dimension)
			}

			if propSheet.Id == "" {
				createProp := prop.DoCreateProp(userSession, connectionToPangolinParams.NfviPangolin, &propRequestInstance)
				entityTypeInstance := entityType.EntityType{Name: sheet.EntityTypeSheet.Name, Id: sheet.EntityTypeSheet.Id, ParentId: sheet.EntityTypeSheet.ParentId, Description: sheet.EntityTypeSheet.Description, Params: sheet.EntityTypeSheet.Params}
				entityTypeProp.DoCreateEntityTypeProp(userSession, connectionToPangolinParams.NfviPangolin, dimensionRequest, entityTypeInstance, createProp)
				doUpdateCell(spreadsheetID, sheet.sheetTitle, propSheet.RowIndex, "H", createProp.Id, ctx)

				//list values
				if propRequestInstance.PropTypeId == "000a0000-0000-4000-8000-000000000000" {
					for _, propListValueItem := range propSheet.ListValues {
						propListValueItem.PropId = createProp.Id
						propListValueInstance := propListValue.PropListValue{Name: propListValueItem.Name, PropId: propListValueItem.PropId}
						propListValue.DoCreatePropListValue(userSession, connectionToPangolinParams.NfviPangolin, &propListValueInstance)
					}
				}

			} else {
				fmt.Println("Prop Id is valid")
			}

		}
	}

}

func getPropRequestInstance(propSheet PropSheet) prop.PropRequest {
	propRequestInstance := prop.PropRequest{
		Name:        propSheet.Name,
		PropGroupId: propSheet.PropGroupId,
		PropTypeId:  propSheet.PropTypeId,
	}

	if propSheet.Params != "" {
		propRequestInstance.Params = propSheet.Params
	}

	if propSheet.Description != "" {
		propRequestInstance.Description = propSheet.Description
	}

	return propRequestInstance
}

func getEntityTypeInstance(typeSheet EntityTypeSheet) entityType.EntityType {

	if typeSheet.Name == "" {
		log.Fatalf("Name is required")
	}

	entityTypeInstance := entityType.EntityType{
		Name: typeSheet.Name,
	}
	if typeSheet.Params != "" {
		entityTypeInstance.Params = typeSheet.Params
	}
	if typeSheet.ParentId != "" {
		entityTypeInstance.ParentId = typeSheet.ParentId
	}
	if typeSheet.Description != "" {
		entityTypeInstance.Description = typeSheet.Description
	}

	return entityTypeInstance
}

func doUpdateCell(sheetId string, sheetTitle string, rowIndex int, columnName string, value string, ctx context.Context) {
	valueRange := sheets.ValueRange{}
	valueRange.Values = append(valueRange.Values, []interface{}{value})
	itoa := strconv.Itoa(rowIndex)
	_, err := sheetsService.Spreadsheets.Values.Update(sheetId, sheetTitle+"!"+columnName+itoa, &valueRange).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		fmt.Println(err)
	}
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

				r := csv.NewReader(strings.NewReader(value[5].(string)))
				r.LazyQuotes = true

				records, err := r.ReadAll()
				if err != nil {
					log.Fatalf("Error reading csv: %s", err)
				}

				for _, record := range records {
					for _, val := range record {
						trim := strings.Trim(val, " ")
						PropListValueItems = append(PropListValueItems, propListValue.PropListValue{Name: trim})
					}

				}
			}

			PropSheet := PropSheet{
				Name:        value[0].(string),
				Description: value[1].(string),
				PropTypeId:  value[2].(string),
				GroupName:   value[3].(string),
				Params:      value[4].(string),
				ListValues:  PropListValueItems,
				Dimension:   value[6].(string),
				RowIndex:    key + 1,
			}
			if len(value) > 7 {
				PropSheet.Id = value[7].(string)
			}

			propSheets = append(propSheets, PropSheet)
		}
		SheetData.PropSheetItems = propSheets
		SheetData.sheetTitle = sheet.Properties.Title

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
	Id          string
	Name        string
	PropGroupId string
	PropTypeId  string
	Description string
	Params      interface{}
	GroupName   string
	ListValues  []propListValue.PropListValue
	Dimension   string
	RowIndex    int
}

type SheetData struct {
	sheetTitle      string
	EntityTypeSheet EntityTypeSheet
	PropSheetItems  []PropSheet
}
