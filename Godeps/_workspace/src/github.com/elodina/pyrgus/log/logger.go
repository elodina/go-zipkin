/* Licensed to Elodina Inc. under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package log

import (
	"fmt"

	log "github.com/elodina/go-zipkin/Godeps/_workspace/src/gopkg.in/cihub/seelog.v2"
	"runtime"
	"strings"
)

var DefaultLogLevel LogLevel = InfoLevel
var DefaultLogFormat string = "%Date/%Time [%LEVEL] %Msg%n"

var Logger = NewDefaultLogger()

// Log is a logger interface. Lets you plug-in your custom logging library instead of using built-in one.
type Log interface {
	//Formats a given message according to given params to log with level Trace.
	Trace(message interface{}, params ...interface{})

	//Formats a given message according to given params to log with level Debug.
	Debug(message interface{}, params ...interface{})

	//Formats a given message according to given params to log with level Info.
	Info(message interface{}, params ...interface{})

	//Formats a given message according to given params to log with level Warn.
	Warn(message interface{}, params ...interface{})

	//Formats a given message according to given params to log with level Error.
	Error(message interface{}, params ...interface{})

	//Formats a given message according to given params to log with level Critical.
	Critical(message interface{}, params ...interface{})
}

// LogLevel represents a logging level.
type LogLevel string

const (
	// TraceLevel is used for debugging to find problems in functions, variables etc.
	TraceLevel LogLevel = "trace"

	// DebugLevel is used for detailed system reports and diagnostic messages.
	DebugLevel LogLevel = "debug"

	// InfoLevel is used for general information about a running application.
	InfoLevel LogLevel = "info"

	// WarnLevel is used to indicate small errors and failures that should not happen normally but are recovered automatically.
	WarnLevel LogLevel = "warn"

	// ErrorLevel is used to indicate severe errors that affect application workflow and are not handled automatically.
	ErrorLevel LogLevel = "error"

	// CriticalLevel is used to indicate fatal errors that may cause data corruption or loss.
	CriticalLevel LogLevel = "critical"
)

func NewDefaultLogger() Log {
	return NewConsoleLogger(DefaultLogLevel, DefaultLogFormat)
}

// ConsoleLogger is an implementation of Log interface that outputs to console.
type ConsoleLogger struct {
	logger log.LoggerInterface
}

// NewConsoleLogger creates a new ConsoleLogger that is configured to write messages to console with given minimum log level
// and given message format.
func NewConsoleLogger(level LogLevel, format string) *ConsoleLogger {
	var config = fmt.Sprintf(`<seelog minlevel="%s">
    <outputs formatid="main">
        <console />
    </outputs>

    <formats>
        <format id="main" format="%s"/>
    </formats>
</seelog>`, level, format)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(config))
	return &ConsoleLogger{logger}
}

// Trace formats a given message according to given params to log with level Trace.
func (dl *ConsoleLogger) Trace(message interface{}, params ...interface{}) {
	dl.logger.Tracef(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

// Debug formats a given message according to given params to log with level Debug.
func (dl *ConsoleLogger) Debug(message interface{}, params ...interface{}) {
	dl.logger.Debugf(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

// Info formats a given message according to given params to log with level Info.
func (dl *ConsoleLogger) Info(message interface{}, params ...interface{}) {
	dl.logger.Infof(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

// Warn formats a given message according to given params to log with level Warn.
func (dl *ConsoleLogger) Warn(message interface{}, params ...interface{}) {
	dl.logger.Warnf(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

// Error formats a given message according to given params to log with level Error.
func (dl *ConsoleLogger) Error(message interface{}, params ...interface{}) {
	dl.logger.Errorf(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

// Critical formats a given message according to given params to log with level Critical.
func (dl *ConsoleLogger) Critical(message interface{}, params ...interface{}) {
	dl.logger.Criticalf(fmt.Sprintf("%s %s", caller(), fmt.Sprint(message)), params...)
}

func caller() string {
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		f := runtime.FuncForPC(pc)
		fileTokens := strings.Split(file, "/")
		file = fileTokens[len(fileTokens)-1]

		funcTokens := strings.Split(f.Name(), "/")
		fun := funcTokens[len(funcTokens)-1]
		return fmt.Sprintf("[%s:%d|%s]", file, line, fun[strings.Index(fun, ".")+1:])
	}

	return "???"
}
