package cmd

type AnsiCode struct {
	Reset     string
	Bold      string
	Underline string
	Blink     string
	Reverse   string
	Conceal   string
	Black     string
	Red       string
	Green     string
	Yellow    string
	Blue      string
	Magenta   string
	Cyan      string
	White     string
	Gray      string
	BgBlack   string
	BgRed     string
	BgGreen   string
	BgYellow  string
	BgBlue    string
	BgMagenta string
	BgCyan    string
	BgWhite   string
	BgGray    string
}

var Ansi = AnsiCode{
	Reset:     "\033[0m",
	Bold:      "\033[1m",
	Underline: "\033[4m",
	Blink:     "\033[5m",
	Reverse:   "\033[7m",
	Conceal:   "\033[8m",
	Black:     "\033[30m",
	Red:       "\033[31m",
	Green:     "\033[32m",
	Yellow:    "\033[33m",
	Blue:      "\033[34m",
	Magenta:   "\033[35m",
	Cyan:      "\033[36m",
	White:     "\033[37m",
	Gray:      "\033[90m",
	BgBlack:   "\033[40m",
	BgRed:     "\033[41m",
	BgGreen:   "\033[42m",
	BgYellow:  "\033[43m",
	BgBlue:    "\033[44m",
	BgMagenta: "\033[45m",
	BgCyan:    "\033[46m",
	BgWhite:   "\033[47m",
	BgGray:    "\033[100",
}

func (a AnsiCode) GetColor(colorName string) string {
	switch colorName {
	case "reset":
		return a.Reset
	case "bold":
		return a.Bold
	case "underline":
		return a.Underline
	case "blink":
		return a.Blink
	case "reverse":
		return a.Reverse
	case "conceal":
		return a.Conceal
	case "black":
		return a.Black
	case "red":
		return a.Red
	case "green":
		return a.Green
	case "yellow":
		return a.Yellow
	case "blue":
		return a.Blue
	case "magenta":
		return a.Magenta
	case "cyan":
		return a.Cyan
	case "white":
		return a.White
	case "gray":
		return a.Gray
	case "bgblack":
		return a.BgBlack
	case "bgred":
		return a.BgRed
	case "bggreen":
		return a.BgGreen
	case "bgyellow":
		return a.BgYellow
	case "bgblue":
		return a.BgBlue
	case "bgmagenta":
		return a.BgMagenta
	case "bgcyan":
		return a.BgCyan
	case "bgwhite":
		return a.BgWhite
	case "bggray":
		return a.BgGray
	default:
		return ""
	}
}
