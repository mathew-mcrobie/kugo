package display

import(
	"fmt"
)

func highlightSelected(str string) string {
	return fmt.Sprintf("\033[7m%s\033[0m", str)
}

func highlight[T any](value T) string {
	return fmt.Sprintf("\033[7m%v\033[0m", value)
}

func (d *Display) DrawMainMenu() {
	d.buildString(d.row, 3, "select number of players by number keys")
	d.row += 2
	d.buildString(d.row, 12, "# Players: ")
	for i := range 4 {
		if i == d.Selection {
			d.buildString(d.row, 23 + i*2, fmt.Sprintf("%s", highlight(i+3)))
			continue
		}
		d.buildString(d.row, 23 + i*2, fmt.Sprintf("%d", i+3))
	}
	d.row += 4
	d.buildString(d.row, 12, "press Enter to begin")
	d.row += 2
	d.buildString(d.row, 8, "press 'q' at any time to quit")
}
