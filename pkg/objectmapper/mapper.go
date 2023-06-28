package objectmapper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	defaultTag = "om"
	comma      = ","
)

type ObjectMapper struct {
	Tag                   string
	DefaultValueSeparator string
}

func (om *ObjectMapper) Map(source, destination interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recovered in Map()", r)
		}
	}()
	destValue := reflect.ValueOf(destination).Elem()
	destType := destValue.Type()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("destination is not a struct: %v", destValue)
	}

	if om.Tag == "" {
		om.Tag = defaultTag
	}
	if om.DefaultValueSeparator == "" {
		om.DefaultValueSeparator = comma
	}
	sourceMap, err := convertToMap(source)
	if err != nil {
		return fmt.Errorf("invalid source: %v err: %s", source, err)
	}

	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := destValue.Field(i)
		destTag := destField.Tag.Get(om.Tag)
		if destTag == "" {
			continue
		}
		defaultValue := ""
		tagParts := strings.Split(destTag, om.DefaultValueSeparator)
		if len(tagParts) > 1 {
			defaultValue = tagParts[1]
		}
		srcValue, err := getSrcValueByTag(sourceMap, tagParts[0])
		if err != nil {
			fmt.Println("Error getting source value for field", destField.Name, ":", err)
		}
		if srcValue == nil && defaultValue != "" {
			srcValue = defaultValue
		}
		convertedValue, err := convertValue(srcValue, destField.Type)
		if err != nil {
			//fallback to default value if conversion fails
			convertedValue, err = convertValue(defaultValue, destField.Type)
			if err != nil {
				fmt.Println("Error converting value for field", destField.Name, ":", err)
			}
		}
		if destFieldValue.CanSet() && convertedValue.IsValid() {
			destFieldValue.Set(convertedValue)
		}
	}
	return nil
}

func getSrcValueByTag(sourceMap map[string]interface{}, destTag string) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recovered in getSrcValueByTag()", r)
		}
	}()
	if destTag == "." {
		return sourceMap, nil
	}
	tags := strings.Split(destTag, ".")
	var srcValue interface{}
	srcValue = sourceMap
	var ok bool

	for _, tag := range tags {
		srcMap, _ := srcValue.(map[string]interface{})
		if strings.Contains(tag, "[") && strings.Contains(tag, "]") {
			// Handle arrays
			arrayName, indexStr := getArrayNameAndIndexStr(tag)
			array, ok := srcMap[arrayName]
			arrayValue := reflect.ValueOf(array)
			if !ok {
				return nil, fmt.Errorf("unable to find value for tag: %s", destTag)
			}
			if !isArrayValue(arrayValue) {
				// element is not an array, simply return value
				return srcValue, nil
			}
			if strings.Contains(indexStr, ":") {
				// fill in from index range of an array like [2:3], or use full array if [:]
				startIndex, endIndex := getStartAndEndIndex(indexStr, arrayValue.Len())
				subArray := make([]interface{}, 0, endIndex-startIndex+1)
				for i := startIndex; i <= endIndex; i++ {
					subArray = append(subArray, arrayValue.Index(i))
				}
				srcValue = subArray
			} else {
				// single element from array
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return nil, fmt.Errorf("invalid array index: %v", indexStr)
				}
				if index < 0 {
					index = 0
				}
				if index >= arrayValue.Len() {
					index = arrayValue.Len() - 1
				}
				srcValue = arrayValue.Index(index).Interface()
			}
		} else {
			srcValue, ok = srcMap[tag]
			if !ok {
				return nil, fmt.Errorf("unable to find value for tag: %s in sourceMap: %#v", destTag, sourceMap)
			}
		}
	}
	return srcValue, nil
}

func convertValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recovered in convertValue()", r)
		}
	}()
	if reflect.TypeOf(value) == targetType {
		return reflect.ValueOf(value), nil
	}
	stringValue := fmt.Sprintf("%v", value)

	switch targetType.Kind() {
	case reflect.Pointer:
		val, err := convertValue(value, targetType.Elem())
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(val)
		return ptr, err
	case reflect.String:
		return reflect.ValueOf(stringValue), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(stringValue, 10, targetType.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string to int: %v", err)
		}
		return reflect.ValueOf(intVal).Convert(targetType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(stringValue, 10, targetType.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string to uint: %v", err)
		}
		return reflect.ValueOf(uintVal).Convert(targetType), nil
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(stringValue, targetType.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string to float: %v", err)
		}
		return reflect.ValueOf(floatVal).Convert(targetType), nil
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string to bool: %v", err)
		}
		return reflect.ValueOf(boolVal).Convert(targetType), nil
	case reflect.Array:
	case reflect.Slice:
		var array reflect.Value
		if isArrayValue(reflect.ValueOf(value)) {
			// if value is slice or array, copy items from it
			arrayValue := reflect.ValueOf(value)
			array = reflect.MakeSlice(reflect.SliceOf(targetType.Elem()), arrayValue.Len(), arrayValue.Len())
			for i := 0; i < arrayValue.Len(); i++ {
				convertedValue, _ := convertValue(arrayValue.Index(i).Interface(), targetType.Elem())
				if convertedValue.IsValid() {
					array.Index(i).Set(convertedValue)
				}
			}
		} else {
			// insert at 0th index of array
			array = reflect.MakeSlice(reflect.SliceOf(targetType.Elem()), 1, 1)
			convertedValue, _ := convertValue(value, targetType.Elem())
			if convertedValue.IsValid() {
				array.Index(0).Set(convertedValue)
			}
		}
		return reflect.ValueOf(array.Interface()).Convert(targetType), nil
	default:
		// to support struct and maps
		val, _ := value.(reflect.Value)
		var jsonStr []byte
		if val.IsValid() {
			jsonStr, _ = json.Marshal(val.Interface())
		} else {
			mpVal, ok := value.(map[string]interface{})
			if ok {
				jsonStr, _ = json.Marshal(mpVal)
			} else {
				jsonStr = []byte(fmt.Sprintf("%v", value))
			}
		}
		obj := reflect.New(targetType).Interface()
		json.Unmarshal(jsonStr, &obj)
		return reflect.ValueOf(obj).Elem(), nil
	}
	return reflect.Value{}, fmt.Errorf("unsupported type conversion from %v to %v", reflect.TypeOf(value), targetType)
}

func isArrayValue(v reflect.Value) bool {
	kind := v.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

func convertToMap(source interface{}) (map[string]interface{}, error) {
	mp, ok := source.(map[string]interface{})
	if ok {
		return mp, nil
	}
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(sourceBytes, &mp)
	return mp, err
}

func getArrayNameAndIndexStr(tag string) (string, string) {
	bracketStart := strings.Index(tag, "[")
	bracketEnd := strings.Index(tag, "]")
	arrayName := tag[:bracketStart]
	indexStr := tag[bracketStart+1 : bracketEnd]
	return arrayName, indexStr
}

func getStartAndEndIndex(indexStr string, len int) (int, int) {
	startIndexStr := indexStr[:strings.Index(indexStr, ":")]
	endIndexStr := indexStr[strings.Index(indexStr, ":")+1:]
	startIndex, _ := strconv.Atoi(startIndexStr)
	endIndex, _ := strconv.Atoi(endIndexStr)
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndexStr == "" || endIndex >= len {
		endIndex = len - 1
	}
	return startIndex, endIndex
}
