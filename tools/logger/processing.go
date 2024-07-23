package logger

import (
	"fmt"
	"reflect"
	"strings"
)

func (d *Session) addRawMessage(m ...interface{}) {
	d.messages = append(d.messages,
		fmt.Sprintf("%s %s%v", d.header, strings.Repeat(d.indentSymbol, d.indentLevel), m))
}

func (d *Session) addMessage(msg ...interface{}) {
	parseMsg := ""
	for _, m := range msg {
		parseMsg += fmt.Sprintf(" %v", m)
	}
	d.messages = append(d.messages,
		fmt.Sprintf("%s %s%s", d.header, strings.Repeat(d.indentSymbol, d.indentLevel), parseMsg))
}

func (d *Session) checkMapOrSlice(value interface{}) (map[string]interface{}, []interface{}, reflect.Value) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, nil, v
	} else if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	mapVal := make(map[string]interface{})
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if v.MapIndex(key).IsValid() && v.MapIndex(key).CanInterface() {
				mapVal[fmt.Sprintf("%v", key)] = v.MapIndex(key).Interface()
			}
		}
	} else if v.Kind() == reflect.Struct {
		typeOfS := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsValid() || !v.Field(i).CanInterface() {
				continue
			}
			val := v.Field(i).Interface()
			paramName := typeOfS.Field(i).Name
			mapVal[paramName] = val
		}
	}

	if len(mapVal) > 0 {
		return mapVal, nil, v
	} else if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		vv := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			vv[i] = v.Index(i).Interface()
		}
		return nil, vv, v
	}
	return nil, nil, v
}

func (d *Session) addMapMessage(name string, m interface{}) {
	mapVal, sliceVal, v := d.checkMapOrSlice(m)
	if mapVal != nil {
		d.addMessage(name + ": {")
		d.addInternalMapMessage(mapVal, 0)
		d.addMessage("}")
	} else if sliceVal != nil {
		d.addMessage(name + ": [")
		d.addInternalSliceMessage(sliceVal, 0)
		d.addMessage("]")
	} else {
		d.addMessage(fmt.Sprintf("%s (kind: %v): %v", name, v.Kind(), v))
	}
}

func (d *Session) addInternalMapMessage(m map[string]interface{}, depth int) {
	defer d.Indent()()
	if depth > d.maxDepth {
		d.addMessage("...")
		return
	}
	for k, value := range m {
		switch value := value.(type) {
		case map[string]interface{}:
			d.addMessage(k + ": {")
			d.addInternalMapMessage(value, depth+1)
			d.addMessage("}")
		case []interface{}:
			d.addMessage(k + ": [")
			d.addInternalSliceMessage(value, depth+1)
			d.addMessage("]")
		default:
			mapVal, sliceVal, v := d.checkMapOrSlice(value)
			if mapVal != nil {
				d.addMessage(k + ": {")
				d.addInternalMapMessage(mapVal, depth+1)
				d.addMessage("}")
			} else if sliceVal != nil {
				d.addMessage(k + ": [")
				d.addInternalSliceMessage(sliceVal, depth+1)
				d.addMessage("]")
			} else {
				d.addMessage(fmt.Sprintf("%s (kind: %v): %v", k, v.Kind(), v))
			}
		}
	}
}

func (d *Session) addInternalSliceMessage(s []interface{}, depth int) {
	defer d.Indent()()
	if depth > d.maxDepth {
		d.addMessage("...")
		return
	}
	for _, value := range s {
		switch value := value.(type) {
		case map[string]interface{}:
			d.addMessage("-")
			d.addInternalMapMessage(value, depth+1)
		case []interface{}:
			d.addMessage("-")
			d.addInternalSliceMessage(value, depth+1)
		default:
			mapVal, sliceVal, v := d.checkMapOrSlice(value)
			if mapVal != nil {
				d.addMessage("- {")
				d.addInternalMapMessage(mapVal, depth+1)
				d.addMessage("}")
			} else if sliceVal != nil {
				d.addMessage("- [")
				d.addInternalSliceMessage(sliceVal, depth+1)
				d.addMessage("]")
			} else {
				d.addMessage(fmt.Sprintf("- (kind: %v): %v", v.Kind(), v))
			}
		}
	}
}
