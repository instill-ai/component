package logger

import (
	"runtime"
	"strings"

	"github.com/fatih/color"
)

func sanitizeName(name string) string {
	funcNameList := strings.Split(name, "/")
	return funcNameList[len(funcNameList)-1]
}

// The argument skip is the number of stack frames to ascend, with 0 identifying the caller of Caller.
func getCallerDetails(skip int) (filename string, funcName string, line int, ok bool) {
	pc, filename, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0, false
	}

	sanitizedFuncName := sanitizeName(runtime.FuncForPC(pc).Name())
	if strings.Contains(sanitizedFuncName, ".") {
		funcNameList := strings.Split(sanitizedFuncName, ".")
		sanitizedFuncName = funcNameList[len(funcNameList)-1]
	}
	sanitizedFilename := sanitizeName(filename)
	return sanitizedFilename, sanitizedFuncName, line, true
}

func setColor(text string, c string) (coloredText string) {
	switch c {
	case "red":
		coloredText = color.RedString(text)
	case "green":
		coloredText = color.GreenString(text)
	case "yellow":
		coloredText = color.YellowString(text)
	case "blue":
		coloredText = color.BlueString(text)
	case "magenta":
		coloredText = color.MagentaString(text)
	case "cyan":
		coloredText = color.CyanString(text)
	case "white":
		coloredText = color.WhiteString(text)
	case "black":
		coloredText = color.BlackString(text)
	default:
		coloredText = text
	}
	return
}
