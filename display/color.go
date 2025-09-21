package display

type color int

const (
	black color = iota
	red
	green
	yellow
	blue
	magenta
	cyan
	white
	normal
)

var colorCodeANSI = map[color]string{
	black:   "\033[30m",
	red:     "\033[31m",
	green:   "\033[32m",
	yellow:  "\033[33m",
	blue:    "\033[34m",
	magenta: "\033[35m",
	cyan:    "\033[36m",
	white:   "\033[37m",
	normal:  "\033[0m",
}

func (c color) String() string {
	return colorCodeANSI[c]
}
