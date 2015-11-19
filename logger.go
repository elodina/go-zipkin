package zipkin

import (
	"fmt"
	log "github.com/cihub/seelog"
)

var Logger log.LoggerInterface = NewDefaultLogger("info")

func NewDefaultLogger(level string) *log.LoggerInterface {
	var config = fmt.Sprintf(`<seelog minlevel="%s">
    <outputs formatid="main">
        <console />
    </outputs>

    <formats>
        <format id="main" format="%%Date/%%Time [%%LEVEL] %%Msg%%n"/>
    </formats>
</seelog>`, level)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(config))
	return logger
}