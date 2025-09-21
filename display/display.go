package display

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"

	"kugo/game"
)

const (
	Reset   = "\033[2J\033[1;1H\033[?25l"
	Restore = "\033[2J\033[1;1H\033[?25h"
)

// Error handling
func (d *Display) RecoverPanic() {
	if captured := recover(); captured != nil {
		err := fmt.Errorf("%v", captured)
		d.chanErr <- err
	}
}

// Terminal Functions

func MoveCursor(row int, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func AddStr(row, col int, str string) {
	MoveCursor(row, col)
	fmt.Print(str)
}

func WrapDisplay(f func() error) error {
	// Put terminal in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// Restore the user's terminal on exit
	defer RestoreDisplay(oldState)
	if err != nil {
		return err
	}
	fmt.Print(Reset)
	err = f() // <-- Where the game is actually running
	if err != nil {
		return err
	}
	return nil
}

func RestoreDisplay(s *term.State) {
	fmt.Print(Restore)
	term.Restore(int(os.Stdin.Fd()), s)
}

type Display struct {
	allPlayers    []*game.Player
	activePlayers []*game.Player
	validTargets  []*game.Player
	current       *game.Player
	victor        *game.Player
	actionLog     *game.ActionLog
	state         game.State
	clock         *time.Ticker
	row           int
	chanErr       chan error
}

func NewDisplay(chanErr chan error) *Display {
	var fps float64 = 1000.0 / 30
	return &Display{clock: time.NewTicker(time.Duration(fps) * time.Millisecond), chanErr: chanErr}
}

// Draw Functions
func (d *Display) resetScreen() {
	fmt.Print("\033[2J\033[1;1H")
	d.row = 1
}

func (d *Display) checkAudience() bool {
	for _, p := range d.activePlayers {
		if !p.IsLocal {
			continue
		}
		return true
	}
	AddStr(d.row, 5, "Biding your time...")
	d.row++
	return false
}

func (d *Display) DrawDisplay(ctx context.Context) {
	defer d.RecoverPanic()
	for {
		if d.state.Phase == game.EndGame {
			break
		}
		d.resetScreen()
		d.drawHeader()
		d.drawPlayers()
		d.drawActionLog()
		d.drawLocalHand()
		d.drawMenu()
		<-d.clock.C
	}
	d.resetScreen()
	d.drawHeader()
	d.drawVictoryScreen()
}

func (d *Display) UpdateDisplay(info *game.DisplayData) {
	d.allPlayers = info.AllPlayers
	d.activePlayers = info.ActivePlayers
	d.validTargets = info.ValidTargets
	d.current = info.Current
	d.actionLog = info.ActionLog
	d.state = info.State
	if d.state.Phase != game.EndGame {
		return
	}
	for _, p := range d.allPlayers {
		if !p.IsAlive() {
			continue
		}
		d.victor = p
	}
}

func (d *Display) drawHeader() {
	AddStr(d.row, 16, "=== KUGO ===")
	d.row += 2
}

func (d *Display) drawPlayers() {
	coinString := "\033[31m~DEAD~\033[0m"
	for _, player := range d.allPlayers {
		marker := "    "
		currentIdx := d.current.Index
		if player.Index == currentIdx {
			marker = ">>> "
		}
		handString := getHandString(player)
		if player.IsAlive() {
			coinString = fmt.Sprintf("%2d", player.Coins)
		}
		playerString := fmt.Sprintf("%s%-12s%s      %s", marker, player.Name, coinString, handString)
		AddStr(d.row, 1, playerString)
		d.row++
	}
	d.row++
}

func (d *Display) drawActionLog() {
	if len(d.actionLog.Items) == 0 {
		return
	}
	for _, msg := range d.actionLog.Items {
		AddStr(d.row, 1, msg)
		d.row++
	}
	d.row++
}

func (d *Display) drawLocalHand() {
	var pHand string
	for _, p := range d.allPlayers {
		if !p.IsLocal {
			continue
		}
		switch len(p.CardsHeld) {
		case 1:
			pHand = fmt.Sprintf("Your hand: [%s]", p.CardsHeld[0])
		case 2:
			pHand = fmt.Sprintf(
				"Your hand: [%s | %s]",
				p.CardsHeld[0],
				p.CardsHeld[1],
			)
		case 3:
			pHand = fmt.Sprintf(
				"Your hand: [%s | %s | %s]",
				p.CardsHeld[0],
				p.CardsHeld[1],
				p.CardsHeld[2],
			)
		case 4:
			pHand = fmt.Sprintf(
				"Your hand: [%s | %s | %s | %s]",
				p.CardsHeld[0],
				p.CardsHeld[1],
				p.CardsHeld[2],
				p.CardsHeld[3],
			)
		}
		AddStr(d.row, 5, pHand)
		d.row += 2
		return
	}
}

func getHandString(p *game.Player) string {
	if len(p.CardsLost) == 2 {
		return fmt.Sprintf("[%s | %s]", p.CardsLost[0].Short(), p.CardsLost[1].Short())
	}
	if len(p.CardsLost) == 0 {
		return fmt.Sprint("[??? | ???]")
	}
	return fmt.Sprintf("[%s | ???]", p.CardsLost[0].Short())
}

func (d *Display) drawMenu() {
	if !d.checkAudience() {
		return
	}
	switch d.state.Phase {
	case game.SelectAction:
		d.drawActionMenu()
	case game.SelectTarget:
		d.drawTargetMenu()
	case game.MakeChallenge, game.ChallengeBlock:
		d.drawChallengeMenu()
	case game.MakeBlock:
		d.drawBlockMenu()
	case game.ChallengeReveal, game.BlockReveal:
		d.drawRevealMenu()
	case game.ChallengeLoss, game.BlockLoss:
		d.drawLossMenu()
	case game.ResolveAction:
		if d.state.Action == game.Assassinate || d.state.Action == game.Coup {
			d.drawLossMenu()
			return
		}
	case game.ExchangeMiddle:
		d.drawReturnTwo()
	case game.ExchangeFinal:
		d.drawReturnOne()
	default:
		panic("Unreachable code! (DrawMenu)")
	}
}

func (d *Display) drawActionMenu() {
	AddStr(d.row, 5, "[1] Income (+1 coin)")
	d.row++
	AddStr(d.row, 5, fmt.Sprintf("[2] Foreign Aid (+2 coins; blocked by %s)", game.Duke))
	d.row++
	AddStr(d.row, 5, "[3] Coup (-7 coins; target loses influence)")
	d.row++
	AddStr(d.row, 5, fmt.Sprintf("\033[37m[4] Assassinate (-3 coins; target loses influences; blocked by %s)", game.Contessa))
	d.row++
	AddStr(d.row, 5, fmt.Sprintf("\033[32m[5] Exchange (Draw 2 cards, then return 2 cards)\033[0m"))
	d.row++
	AddStr(d.row, 5, fmt.Sprintf("\033[36m[6] Steal (Take up to 2 coins from target; blocked by %s \033[36mor %s)", game.Ambassador, game.Captain))
	d.row++
	AddStr(d.row, 5, "\033[35m[7] Tax (+3 coins)\033[0m")
	d.row += 2
	AddStr(d.row, 1, "The time has come to act:")
	d.row++
}

func (d *Display) drawTargetMenu() {
	AddStr(d.row, 5, "And you will act upon?")
	d.row++
	var counter int
	for _, p := range d.validTargets {
		counter++
		AddStr(d.row, 5, fmt.Sprintf("[%d] %s", counter, p.Name))
		d.row++
	}
}

func (d *Display) drawChallengeMenu() { // Saved for later: pInfo []*Player) {
	AddStr(d.row, 5, "Are they bluffing?")
	d.row++
	AddStr(d.row, 5, "[1] Challenge")
	d.row++
	AddStr(d.row, 5, "[0] Pass")
	d.row++
}

func (d *Display) drawBlockMenu() { // Saved for later: pInfo []*Player) {}
	AddStr(d.row, 5, "Will you block?")
	d.row++
	if d.state.Action != game.Steal {
		AddStr(d.row, 5, "[1] Block")
		d.row++
	} else {
		AddStr(d.row, 5, fmt.Sprintf("[1] Block with %s", game.Ambassador))
		d.row++
		AddStr(d.row, 5, fmt.Sprintf("[2] Block with %s", game.Captain))
		d.row++
	}
	AddStr(d.row, 5, "[0] Pass")
	d.row++
}

func (d *Display) drawRevealMenu() {
	AddStr(d.row, 5, "Show the world the truth. Reveal a card:")
	d.row++
	for i, card := range d.activePlayers[0].CardsHeld {
		AddStr(d.row, 5, fmt.Sprintf("[%d] Reveal %s", i+1, card))
		d.row++
	}
}

func (d *Display) drawLossMenu() {
	AddStr(d.row, 5, "Who has disappointed you? Choose a card to lose:")
	d.row++
	for i, card := range d.activePlayers[0].CardsHeld {
		AddStr(d.row, 5, fmt.Sprintf("[%d] Lose %s", i+1, card))
		d.row++
	}
}

func (d *Display) drawReturnTwo() {
	AddStr(d.row, 5, "Who do you no longer need? (Returned 0 of 2)")
	d.row++
	for i, card := range d.activePlayers[0].CardsHeld {
		AddStr(d.row, 5, fmt.Sprintf("[%d] Return %s", i+1, card))
		d.row++
	}
}

func (d *Display) drawReturnOne() {
	AddStr(d.row, 5, "Who do you no longer need? (Returned 1 of 2)")
	d.row++
	for i, card := range d.activePlayers[0].CardsHeld {
		AddStr(d.row, 5, fmt.Sprintf("[%d] Return %s", i+1, card))
		d.row++
	}
	AddStr(d.row, 5, fmt.Sprint("[0] Cancel"))
	d.row++
}

func (d *Display) drawVictoryScreen() {
	AddStr(d.row, 0, fmt.Sprintf("The game is over, and %s is the victor!", d.victor))
}
