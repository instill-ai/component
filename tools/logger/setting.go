package logger

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	Verbose             = 1
	DevelopVerboseLevel = 1
	StaticVerboseLevel  = 2
)

type Session struct {
	sessionID     string
	title         string
	messages      []string
	active        bool
	halfBannerLen int
	indentLevel   int
	maxDepth      int
	verboseLevel  int
	rawHeader     string
	header        string
	indentSymbol  string
}

// Session Logger is only verbose when package verbose is greater than the verbose level specified here.
//
// If no name is provided, the name will be <filename>:<function name> of the caller
func (d *Session) SessionStart(name string, verboseLevel int) (self *Session) {
	d.verboseLevel = verboseLevel
	if Verbose < verboseLevel {
		return d
	}
	defer d.flush()
	if d.active {
		d.SessionEnd()
	}

	if name == "" {
		name = "Unknown"
		sanitizedFilename, sanitizedFuncName, _, ok := getCallerDetails(1)
		if ok {
			name = fmt.Sprintf("%s:%s", sanitizedFilename, sanitizedFuncName)
		}
	}
	/************ Set Default Value ************/
	d.halfBannerLen = 20
	d.indentLevel = 0
	d.maxDepth = 5 * Verbose
	d.active = true
	d.messages = []string{}
	d.indentSymbol = "  "
	/*******************************************/

	halfBanner := strings.Repeat("=", d.halfBannerLen)
	d.sessionID = name

	d.rawHeader = fmt.Sprintf("[%s]", name)
	d.header = color.BlackString(d.rawHeader)
	d.title = fmt.Sprintf("%s %s %s", halfBanner, name, halfBanner)
	d.messages = append(d.messages, d.title)
	return d
}

func (d *Session) SessionEnd() (self *Session) {
	if Verbose < d.verboseLevel || !d.active {
		return d
	}
	defer d.flush()
	defer func() {
		d.indentLevel = 0
		d.active = false
	}()
	endHalfBanner := strings.Repeat("=", d.halfBannerLen-2)
	endBanner := fmt.Sprintf("%s %s end %s", endHalfBanner, d.sessionID, endHalfBanner)
	d.messages = append(d.messages, endBanner)
	return d
}

func (d *Session) IncrementIndent() (self *Session) {
	d.indentLevel++
	return d
}
func (d *Session) DecrementIndent() (self *Session) {
	d.indentLevel--
	return d
}

func (d *Session) SetMaxDepth(depth int) (self *Session) {
	d.maxDepth = depth
	return d
}

func (d *Session) SetIndentSymbol(symbol string) (self *Session) {
	d.indentSymbol = symbol
	return d
}
