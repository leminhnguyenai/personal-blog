package simplelog

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"
)

// The output that all messages be sent to
var OUTPUT = os.Stdout

type Style string

var (
	Reset string = "\033[0m"

	// Foreground colors
	RedFg        Style = "\033[38;5;196m" // bright red
	GreenFg      Style = "\033[38;5;82m"  // softer green (was 46)
	MutedGreenFg Style = "\033[38;5;102m"
	YellowFg     Style = "\033[38;5;220m" // darker yellow (was 226)
	BlueFg       Style = "\033[38;5;33m"  // navy blue (was 27)
	MagentaFg    Style = "\033[38;5;201m"
	CyanFg       Style = "\033[38;5;51m"
	GrayFg       Style = "\033[38;5;244m" // darker gray
	WhiteFg      Style = "\033[38;5;15m"  // white (unchanged)
	OrangeFg     Style = "\033[38;5;208m" // 🟠 orange

	// Background colors
	RedBg        Style = "\033[48;5;196m"
	GreenBg      Style = "\033[48;5;28m"
	MutedGreenBg Style = "\033[48;5;102m"
	YellowBg     Style = "\033[48;5;220m"
	BlueBg       Style = "\033[48;5;33m"
	MagentaBg    Style = "\033[48;5;201m"
	CyanBg       Style = "\033[48;5;51m"
	GrayBg       Style = "\033[48;5;236m"
	WhiteBg      Style = "\033[48;5;15m"
	OrangeBg     Style = "\033[48;5;208m" // 🟠 orange

	// Styles
	Bold      Style = "\033[1m"
	Underline Style = "\033[4m"
)

type LogType string

const (
	FATAL LogType = "FATAL"
	ERROR LogType = "ERROR"
	WARN  LogType = "WARN"
	INFO  LogType = "INFO"
	DEBUG LogType = "DEBUG"
	TRACE LogType = "TRACE"
)

var LogLevelsMap = map[LogType]int{
	FATAL: 10,
	ERROR: 20,
	WARN:  30,
	INFO:  40,
	DEBUG: 50,
	TRACE: 60,
}

func printTime() {
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	fmt.Fprintf(OUTPUT, ColorString(currentTime+" ", CyanFg))
}

func printStack() {
	var traceStack string

	var currentStack []byte

	for _, trace := range debug.Stack() {
		if string(trace) == "\n" {
			traceStack += "     " + string(currentStack) + "\n"
			currentStack = []byte{}
			continue
		}

		currentStack = append(currentStack, trace)
	}

	fmt.Fprint(OUTPUT, ColorString("└──TRACESTACK:\n", Bold+MutedGreenFg)+traceStack)
}

func Fatalf(err error) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["FATAL"] {
		return
	}

	printTime()
	fmt.Fprint(OUTPUT, ColorString("[FATAL]", Bold+WhiteFg+RedBg)+" > "+err.Error())
	printStack()
	os.Exit(1)
}

func Errorf(err error) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["ERROR"] {
		return
	}

	printTime()
	fmt.Fprint(OUTPUT, ColorString("[ERROR]", Bold+RedFg)+" > "+err.Error())
	printStack()
}

func Warnf(err error) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["WARN"] {
		return
	}

	printTime()
	fmt.Fprint(OUTPUT, ColorString("[WARN]", Bold+YellowFg)+" > "+err.Error())
	printStack()
}

func Infof(msg any) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["INFO"] {
		return
	}

	printTime()

	switch reflect.TypeOf(msg) {
	case reflect.TypeOf(map[string]string{}):
		// TODO: Pretty print this in the future
		str := ColorString("{", OrangeFg)

		for key, val := range msg.(map[string]string) {
			str += ColorString(`"`+key+`": `, GreenFg) + val + ColorString(", ", OrangeFg)
		}

		str += ColorString("}\n", OrangeFg)

		fmt.Fprint(OUTPUT, ColorString("[INFO]", Bold+GreenFg)+" > "+str)
	default:
		fmt.Fprint(OUTPUT, ColorString("[INFO]", Bold+GreenFg)+" > "+fmt.Sprintf("%v", msg))
	}
}

func Debugf(msg any) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["DEBUG"] {
		return
	}

	printTime()

	switch reflect.TypeOf(msg) {
	case reflect.TypeOf(map[string]string{}):
		// TODO: Pretty print this in the future
		str := ColorString("{", OrangeFg)

		for key, val := range msg.(map[string]string) {
			str += ColorString(`"`+key+`": `, GreenFg) + val + ColorString(", ", OrangeFg)
		}

		str += ColorString("}\n", OrangeFg)

		fmt.Fprint(OUTPUT, ColorString("[DEBUG]", Bold+BlueFg)+" > "+str)
	default:
		fmt.Fprint(OUTPUT, ColorString("[DEBUG]", Bold+BlueFg)+" > "+fmt.Sprintf("%v", msg))
	}
}

func Tracef(msg any) {
	logLevel := LogLevelsMap[LogType(os.Getenv("LOG_LEVEL"))]
	if logLevel == 0 || logLevel > LogLevelsMap["TRACE"] {
		return
	}

	printTime()

	switch reflect.TypeOf(msg) {
	case reflect.TypeOf(map[string]string{}):
		// TODO: Pretty print this in the future
		str := ColorString("{", OrangeFg)

		for key, val := range msg.(map[string]string) {
			str += ColorString(`"`+key+`": `, GreenFg) + val + ColorString(", ", OrangeFg)
		}

		str += ColorString("}\n", OrangeFg)

		fmt.Fprint(OUTPUT, ColorString("[TRACE]", Bold+GrayFg)+" > "+str)
	default:
		fmt.Fprint(OUTPUT, ColorString("[TRACE]", Bold+GrayFg)+" > "+fmt.Sprintf("%v", msg))
	}
}

func ColorString(str string, colors ...Style) string {
	result := str

	for _, color := range colors {
		result = string(color) + result
	}

	return result + Reset
}
