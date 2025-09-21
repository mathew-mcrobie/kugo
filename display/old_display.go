// Package display provides all display functionality required.String() for drawing
// the terminal UI for kugo. This is a low-level system using ANSI escape
// codes and is powered.String() primarily by the x/term library.
package display

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const reset string = "\033[2J\033[1;1H\033[?25l"
const restore string = "\033[2J\033[1;1H\033[?25h"

func Wrap(f func()) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer restoreTerm(oldState)
	f()
}

func restoreTerm(previousState *term.State) {
	fmt.Print(restore)
	term.Restore(int(os.Stdin.Fd()), previousState)
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func addString(row, col int, s string) {
	moveCursor(row, col)
	fmt.Print(s)
}

func resetScreen() {
	fmt.Print(reset)
}

func DrawState() {
	resetScreen()
	row := 1
	addString(row, 10, "=== COUP ===")
	row += 2
	addString(row, 1, fmt.Sprintf("    You: 4 coins - [%sAMB%s | %sCON%s]", green.String(), normal.String(), red.String(), normal.String()))
	row += 1
	addString(row, 1, "    Alice: DEAD")
	row += 1
	addString(row, 1, fmt.Sprintf(">>> Bob: 1 coins - [??? | %sASS%s]", blue.String(), normal.String()))
	row += 2
	addString(row, 1, fmt.Sprintf("Your hand: [%sAmbassador%s | %sContessa%s]", green.String(), normal.String(), red.String(), normal.String()))
	row += 2
	addString(row, 1, "Select your action:")
}
