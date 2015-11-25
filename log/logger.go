package log

import (
	"fmt"

	log "gopkg.in/cihub/seelog.v2"
)

// Logger used by this client. Defaults to build-in logger with Info log level.
var Logger log.LoggerInterface = NewDefaultLogger("info")

func NewDefaultLogger(logLevel string) log.LoggerInterface {
	var config = fmt.Sprintf(`<seelog minlevel="%s">
    <outputs formatid="main">
        <console />
    </outputs>

    <formats>
        <format id="main" format="%%Date/%%Time [%%LEVEL] %%Msg%%n"/>
    </formats>
</seelog>`, logLevel)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(config))
	return logger
}