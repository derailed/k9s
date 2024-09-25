// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"regexp"

	"github.com/rs/zerolog/log"
)

// LogChan represents a channel for logs.
type LogChan chan *LogItem

var ItemEOF = new(LogItem)

var springRegexp = regexp.MustCompile(`(?P<Date>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}) (?P<LogLevel>INFO|INF|WARN|WARNING|ERROR|SEVERE|ERR) (?P<LoggingClass>[a-zA-Z.-]*) (?P<Thread>\[[^\]]*\]) (?P<LogMessage>.*)`)

// var gigaspaceRegexp = regexp.MustCompile(`(.*)`)
var gigaspaceRegexp = regexp.MustCompile(`(?P<Date>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}) (?P<Repository>[a-zA-Z.-]*) (?P<LogLevel>INFO|INF|WARN|WARNING|ERROR|SEVERE|ERR) (?P<LoggingClass>\[[^\]]*]) (?P<LogMessage>.*)`)
var correlationIdRegexp = regexp.MustCompile(`\[correlationId=\S*] `)
var sessionIdRegexp = regexp.MustCompile(`\[sessionId=\S*] `)
var customerIdRegexp = regexp.MustCompile(`\[customerId=\S*] `)

// LogItem represents a container log line.
type LogItem struct {
	Pod, Container  string
	SingleContainer bool
	Bytes           []byte
	IsError         bool
}

// NewLogItem returns a new item.
func NewLogItem(bb []byte) *LogItem {
	return &LogItem{
		Bytes: bb,
	}
}

// NewLogItemFromString returns a new item.
func NewLogItemFromString(s string) *LogItem {
	return &LogItem{
		Bytes: []byte(s),
	}
}

// ID returns pod and or container based id.
func (l *LogItem) ID() string {
	if l.Pod != "" {
		return l.Pod
	}
	return l.Container
}

// GetTimestamp fetch log lime timestamp
func (l *LogItem) GetTimestamp() string {
	index := bytes.Index(l.Bytes, []byte{' '})
	if index < 0 {
		return ""
	}
	return string(l.Bytes[:index])
}

// Info returns pod and container information.
func (l *LogItem) Info() string {
	return l.Pod + "::" + l.Container
}

// IsEmpty checks if the entry is empty.
func (l *LogItem) IsEmpty() bool {
	return len(l.Bytes) == 0
}

// Size returns the size of the item.
func (l *LogItem) Size() int {
	return 100 + len(l.Bytes) + len(l.Pod) + len(l.Container)
}

// Render returns a log line as string.
func (l *LogItem) Render(paint string, logOptions *LogOptions, bb *bytes.Buffer) {
	index := bytes.Index(l.Bytes, []byte{' '})
	if logOptions.ShowTimestamp && index > 0 {
		bb.WriteString("[gray::b]")
		bb.Write(l.Bytes[:index])
		bb.WriteString(" ")
		if l := 30 - len(l.Bytes[:index]); l > 0 {
			bb.Write(bytes.Repeat([]byte{' '}, l))
		}
		bb.WriteString("[-::-]")
	}

	if l.Pod != "" {
		bb.WriteString("[" + paint + "::]" + l.Pod)
	}

	if !l.SingleContainer && l.Container != "" {
		// if len(l.Pod) > 0 {
		// 	bb.WriteString(" ")
		// }
		// bb.WriteString("[" + paint + "::b]" + l.Container + "[-::-] ")
	} else if len(l.Pod) > 0 {
		bb.WriteString("[-::] ")
	}

	if index > 0 {
		if logOptions.CleanLogs == "" || logOptions.CleanLogs == "Off" {
			bb.Write(l.Bytes[index+1:])
			return
		}
		// TODO
		regexp := springRegexp
		matches := regexp.FindSubmatch(l.Bytes[index+1:])
		if matches == nil {
			regexp = gigaspaceRegexp
			matches = gigaspaceRegexp.FindSubmatch(l.Bytes[index+1:])
			if matches == nil {
				log.Debug().Msgf("[Render] no matches for %q", l.Bytes[index+1:])
				bb.Write(l.Bytes[index+1:])
				return
			}
			log.Debug().Msgf("[Render] Found gigaspaceMatches for %q", l.Bytes[index+1:])
			log.Debug().Msgf("[Render] %q", matches[0])
		}

		date := matches[regexp.SubexpIndex("Date")]
		logLevel := matches[regexp.SubexpIndex("LogLevel")]
		loggingClass := matches[regexp.SubexpIndex("LoggingClass")]
		logMessage := matches[regexp.SubexpIndex("LogMessage")]
		logMessage = correlationIdRegexp.ReplaceAll(logMessage, []byte(""))
		logMessage = sessionIdRegexp.ReplaceAll(logMessage, []byte(""))
		logMessage = customerIdRegexp.ReplaceAll(logMessage, []byte(""))

		// log.Info().Msgf("[Render] %q", logMessage)
		bb.Write([]byte("[gray::b]"))
		bb.Write(date)
		bb.Write([]byte("[-::-]"))
		bb.Write([]byte(GetLogLevelColor(string(logLevel))))
		bb.Write(logLevel)
		bb.Write([]byte("[-::-] "))
		if logOptions.CleanLogs == "On (+)" {
			bb.Write([]byte("[yellow::b]"))
			bb.Write(loggingClass)
			bb.Write([]byte("[-::-] "))
		}
		bb.Write(logMessage)
		bb.Write([]byte("\n"))
	} else {
		bb.Write(l.Bytes)
	}
}

func GetLogLevelColor(logLevel string) string {
	if logLevel == "INFO" {
		return "[white::b] "
	} else if logLevel == "WARN" {
		return "[orange::b] "
	} else if logLevel == "ERROR" || logLevel == "SEVERE" || logLevel == "ERR" {
		return "[red::b] "
	} else {
		return "[white::b] "
	}
}
