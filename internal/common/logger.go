package common

type Style string

var (
	Reset = "\033[0m"

	RedFg     Style = "\033[38;5;196m"
	GreenFg   Style = "\033[38;5;46m"
	YellowFg  Style = "\033[38;5;226m"
	BlueFg    Style = "\033[38;5;27m"
	MagentaFg Style = "\033[38;5;201m"
	CyanFg    Style = "\033[38;5;51m"
	GrayFg    Style = "\033[38;5;245m"
	WhiteFg   Style = "\033[38;5;15m"

	RedBg     Style = "\033[48;5;196m"
	GreenBg   Style = "\033[48;5;46m"
	YellowBg  Style = "\033[48;5;226m"
	BlueBg    Style = "\033[48;5;27m"
	MagentaBg Style = "\033[48;5;201m"
	CyanBg    Style = "\033[48;5;51m"
	GrayBg    Style = "\033[48;5;245m"
	WhiteBg   Style = "\033[48;5;15m"

	Bold      Style = "\033[1m"
	Underline Style = "\033[4m"
)

func ColorStringFg(str string, colors ...Style) string {
	result := str

	for _, color := range colors {
		result = string(color) + result
	}

	return result + Reset
}
