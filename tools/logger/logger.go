package logger

import (
	"fmt"
	"strings"
)

func (d *Session) Indent() func() *Session {
	d.AddIndent()
	return d.RemoveIndent
}

func (d *Session) Separator() {
	if Verbose < d.verboseLevel {
		return
	}
	d.messages = append(d.messages, strings.Repeat("=", d.halfBannerLen*2+len(d.sessionID)+2))
}

func (d *Session) flush() {
	if Verbose < d.verboseLevel {
		return
	}
	if len(d.messages) == 0 {
		return
	}
	for _, msg := range d.messages {
		fmt.Println(msg)
	}
	d.messages = []string{}
}

func (d *Session) autoPrint(msg ...interface{}) {
	switch len(msg) {
	case 0:
		return
	case 1:
		d.addMessage(fmt.Sprintf("%v", msg[0]))
	case 2:
		key, ok := msg[0].(string)
		if !ok {
			_, funcName, line, _ := getCallerDetails(2)
			key = fmt.Sprintf("%s:%d", funcName, line)
			d.addMapMessage(key, msg)
		} else {
			d.addMapMessage(key, msg[1])
		}
	default:
		key, ok := msg[0].(string)
		if !ok {
			_, funcName, line, _ := getCallerDetails(2)
			key = fmt.Sprintf("%s:%d", funcName, line)
			d.addMapMessage(key, msg)
		} else {
			d.addMapMessage(key, msg[1:])
		}
	}
}

// Info logs messages with black color
func (d *Session) Info(msg ...interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.header = setColor(d.rawHeader, "default")
	d.autoPrint(msg...)
}

// Alias for Info
func (d *Session) Message(msg ...interface{}) {
	d.Info(msg...)
}

// Success logs messages with green color
func (d *Session) Success(msg ...interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.header = setColor(d.rawHeader, "green")
	d.autoPrint(msg...)
}

// Warn logs messages with yellow color
func (d *Session) Warn(msg ...interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.header = setColor(d.rawHeader, "yellow")
	d.autoPrint(msg...)
}

// Error logs messages with red color
func (d *Session) Error(msg ...interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.header = setColor(d.rawHeader, "red")
	d.autoPrint(msg...)
}

func (d *Session) Raw(msg ...interface{}) {
	if Verbose < d.verboseLevel {
		return
	}
	defer d.flush()
	d.header = setColor(d.rawHeader, "")
	d.addRawMessage(msg...)
}
