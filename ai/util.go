package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InsertMapToJSON inserts a map into a JSON object at a specified key.
func InsertMapToJSON(jsonB *[]byte, mapToInsert map[string]interface{}, targetKey string) error {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(*jsonB, &jsonMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	keys := strings.Split(targetKey, ".")

	currentMap := jsonMap
	for i, key := range keys {
		if i == len(keys)-1 {
			existingValue, exists := currentMap[key]
			if !exists {
				return fmt.Errorf("target key does not exist in the JSON object")
			}

			existingMap, ok := existingValue.(map[string]interface{})
			if !ok {
				return fmt.Errorf("target key is not a map")
			}

			for k, v := range mapToInsert {
				existingMap[k] = v
			}
		} else {
			nestedValue, exists := currentMap[key]
			if !exists {
				return fmt.Errorf("target key does not exist in the JSON object")
			}

			nestedMap, ok := nestedValue.(map[string]interface{})
			if !ok {
				return fmt.Errorf("target key is not a map")
			}

			currentMap = nestedMap
		}
	}

	modifiedJSON, err := json.Marshal(jsonMap)

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	*jsonB = modifiedJSON

	return nil
}
