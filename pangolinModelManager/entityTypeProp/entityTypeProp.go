package entityTypeProp

import (
	"encoding/json"
	"fmt"
	"io"
	"pangolinModelManager/dimension"
	"pangolinModelManager/entityType"
	"pangolinModelManager/prop"
	"pangolinModelManager/restClient"
	"pangolinModelManager/security"
)

func DoCreateEntityTypeProp(session security.UserSession, pangolin string, dimension dimension.Dimension, entityType entityType.EntityType, prop prop.Prop) {
	resource := "/api/v1/EntityTypeProp"
	values := map[string]string{"entityTypeId": entityType.Id, "propId": prop.Id, "dimensionId": dimension.Id}

	resp, resBody := restClient.SendPostRequest(values, session, pangolin, resource)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == 201 {
		fmt.Println("EntityTypeProp created successfully", string(resBody))
	} else {
		var errres map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errres)
		fmt.Printf("error during EntityTypeProp creation %v %v \n", resp.StatusCode, errres)
		panic(resp.Body)
	}
}
