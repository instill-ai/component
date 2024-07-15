package jira

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	Verbose             = 1
	DefaultVerboseLevel = 1
	DevelopVerboseLevel = 1
	StaticVerboseLevel  = 2
)

type DebugSession struct {
	SessionID     string   `json:"session-id"`
	Title         string   `json:"title"`
	Messages      []string `json:"messages"`
	halfBannerLen int
	indentLevel   int
	maxDepth      int
	verboseLevel  int
}

// Session Logger is only verbose when package verbose is greater than the verbose level specified here
func (d *DebugSession) SessionStart(name string, verboseLevel int) {
	d.verboseLevel = verboseLevel
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.SessionID = name
	d.halfBannerLen = 20
	halfBanner := strings.Repeat("=", d.halfBannerLen)
	if d.Messages == nil {
		d.Messages = []string{}
	}
	d.Title = fmt.Sprintf("%s %s %s", halfBanner, name, halfBanner)
	d.Messages = append(d.Messages, d.Title)
	d.indentLevel = 0
	d.maxDepth = 5 * Verbose
}

func (d *DebugSession) AddMessage(msg ...string) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	parseMsg := strings.Join(msg, " ")
	d.Messages = append(d.Messages,
		fmt.Sprintf("[%s] %s%s", d.SessionID, strings.Repeat("\t", d.indentLevel), parseMsg))
}

// addMapMessage adds a map message to the debug session
// if the map is empty, it will simply add "Map: {}"
func (d *DebugSession) AddMapMessage(name string, m interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	if name == "" {
		name = "Map"
	}
	d.AddMessage(name + ": {")
	defer d.AddMessage("}")

	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		d.AddMessage("Not a map")
		return
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
	} else {
		d.AddMessage("Not parseable as a map. Type: ", v.Kind().String())
		return
	}
	d.addInternalMapMessage(mapVal, 0)
}

func (d *DebugSession) AddRawMessage(m interface{}) {
	defer d.flush()
	d.Messages = append(d.Messages,
		fmt.Sprintf("[%s] %s%v", d.SessionID, strings.Repeat("\t", d.indentLevel), m))
}

func (d *DebugSession) addInternalMapMessage(m map[string]interface{}, depth int) {
	d.indentLevel++
	defer func() {
		d.indentLevel--
	}()
	if depth > d.maxDepth {
		d.AddMessage("...")
		return
	}
	for k, v := range m {
		switch v := v.(type) {
		case map[string]interface{}:
			d.AddMessage(k + ": {")
			d.addInternalMapMessage(v, depth+1)
			d.AddMessage("}")
		case []interface{}:
			d.AddMessage(k + ": {")
			d.addInternalSliceMessage(v, depth+1)
			d.AddMessage("}")
		default:
			d.AddMessage(fmt.Sprintf("%s: %v", k, v))
		}
	}
}

func (d *DebugSession) addInternalSliceMessage(s []interface{}, depth int) {
	d.indentLevel++
	defer func() {
		d.indentLevel--
	}()
	if depth > d.maxDepth {
		d.AddMessage("...")
		return
	}
	for _, v := range s {
		switch v := v.(type) {
		case map[string]interface{}:
			d.AddMessage("-")
			d.addInternalMapMessage(v, depth+1)
		case []interface{}:
			d.AddMessage("-")
			d.addInternalSliceMessage(v, depth+1)
		default:
			d.AddMessage(fmt.Sprintf("- %v", v))
		}
	}
}

func (d *DebugSession) SessionEnd() {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	defer func() {
		d.indentLevel = 0
	}()
	endHalfBanner := strings.Repeat("=", d.halfBannerLen-2)
	endBanner := fmt.Sprintf("%s %s end %s", endHalfBanner, d.SessionID, endHalfBanner)
	d.Messages = append(d.Messages, endBanner)
}

func (d *DebugSession) flush() {
	if Verbose < d.verboseLevel {
		return
	}
	for _, msg := range d.Messages {
		fmt.Println(msg)
	}
	d.Messages = []string{}
}
