package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var Daemon zerolog.Logger
var Access zerolog.Logger

var m sync.Mutex

// Initialize the logging system with the given configuration.
func Initialize(c *Config) {
	m.Lock()
	defer m.Unlock()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFunc = time.Now().UTC
	zerolog.CallerMarshalFunc = formatCaller

	Daemon = newLogger(c.Daemon, os.Stdout)
	Access = newLogger(c.HttpAccess, os.Stderr)
}

func newLogger(c *LoggerConfig, writer io.Writer) zerolog.Logger {
	var l zerolog.Logger

	if !c.Enabled {
		return zerolog.Nop()
	}

	if c.PrettyPrint {
		l = zerolog.New(zerolog.ConsoleWriter{
			Out:          writer,
			TimeLocation: time.UTC,
			TimeFormat:   time.RFC3339Nano,
			FormatLevel:  formatLevel,
		})
	} else {
		l = zerolog.New(writer)
	}

	ctx := l.With()

	if c.IncludeCaller {
		ctx = ctx.Caller()
	}

	return ctx.Timestamp().Logger().Level(parseLevel(c.Level))
}

func formatLevel(i interface{}) string {
	var l zerolog.Level

	switch v := i.(type) {
	case zerolog.Level:
		l = v
	case string:
		l = parseLevel(v)
	}

	ul := strings.ToUpper(l.String())
	if c, ok := zerolog.LevelColors[l]; ok {
		return colorize(ul, c, false)
	}

	return ul
}

func formatCaller(pc uintptr, file string, line int) string {
	fileParts := strings.Split(file, string(os.PathSeparator))
	truncFile := path.Join(fileParts[len(fileParts)-3:]...)

	return fmt.Sprintf("%s:%d", truncFile, line)
}

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
