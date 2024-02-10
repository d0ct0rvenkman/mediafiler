package logfmt

import (
	"fmt"
	"strings"

	logrus "github.com/sirupsen/logrus"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

type NonDebugFormatter struct {
}

func (f *NonDebugFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// HH:MM:SS L xxx msg
	// L: one letter log level
	// xxx: "indent" field from event to pad pretty-printed messages
	indent := ""

	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	if field, ok := data["indent"]; ok {
		indent = field.(string)
	}

	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = blue
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	case logrus.InfoLevel:
		levelColor = gray
	default:
		levelColor = gray
	}

	lineString := fmt.Sprintf("%02d:%02d:%02d %s :: %s",
		entry.Time.Hour(),
		entry.Time.Minute(),
		entry.Time.Second(),
		strings.ToUpper(string(entry.Level.String()[0])),
		indent+entry.Message,
	)

	lineBytes := []byte(fmt.Sprintf("\x1b[%dm%s\x1b[0m", levelColor, lineString))

	return append(lineBytes, '\n'), nil
}
