package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	Reset   = "\033[2J\033[1;1H\033[?25l"
	Restore = "\033[2J\033[1;1H\033[?25h"
)

var PassError = fmt.Errorf("Player passed")
var UserQuitError = fmt.Errorf("User requested quit")

func RoutePanic(chanErr chan<- error, funcName string) {
	if r := recover(); r != nil {
		chanErr <- fmt.Errorf("Panic in %s: %v", funcName, r)
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

func ResetScreen() {
	fmt.Print("\033[1J\033[1;1H")
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

// Draw Functions
func CheckAudience(row *int, pInfo []*Player) bool {
	for _, p := range pInfo {
		if p.Category == IsBot {
			continue
		}
		if p.IsFocused {
			return true
		}
		AddStr(*row, 5, "Biding your time...")
		*row++
		return false
	}
	panic("Unreachable code! (Check Audience)")
}

func DrawDisplay(game *GameState) {
	row := 1
	ResetScreen()
	DrawHeader(&row)
	DrawPlayers(&row, game.Players)
	DrawActionLog(&row, game.ActionLog.Items)
	DrawLocalHand(&row, game.Players)
	DrawMenu(&row, game)
}

func DrawHeader(row *int) {
	AddStr(*row, 16, "=== KUGO ===")
	*row += 2
}

func DrawPlayers(row *int, players []*Player) {
	for _, player := range players {
		marker := "    "
		if player.IsCurrent {
			marker = ">>> "
		}
		HandString := GetHandString(player)
		PlayerString := fmt.Sprintf("%s%-12s%2d coins      %s", marker, player.Name, player.Coins, HandString)
		AddStr(*row, 1, PlayerString)
		*row++
	}
	*row++
}

func DrawActionLog(row *int, actionLog []string) {
	for _, m := range actionLog {
		AddStr(*row, 1, m)
		*row++
	}
	*row++
}

func DrawLocalHand(row *int, players []*Player) {
	for _, p := range players {
		if p.Category != IsLocal {
			continue
		}
		AllCards := slices.Concat(p.CardsHeld, p.CardsLost)
		pHand := fmt.Sprintf("Your hand: [%s | %s]", AllCards[0].Full(), AllCards[1].Full())
		AddStr(*row, 5, pHand)
		*row += 2
		return
	}
}

func GetHandString(p *Player) string {
	if len(p.CardsLost) == 2 {
		return fmt.Sprintf("[%s | %s]", p.CardsLost[0].Short(), p.CardsLost[1].Short())
	}
	if p.Category == IsLocal {
		AllCards := slices.Concat(p.CardsHeld, p.CardsLost)
		return fmt.Sprintf("[%s | %s]", AllCards[0].Short(), AllCards[1].Short())
	}
	if len(p.CardsHeld) == 2 {
		return fmt.Sprint("[??? | ???]")
	}
	return fmt.Sprintf("[%s | ???]", p.CardsLost[0].Short())
}

func DrawMenu(row *int, gs *GameState) {
	if !CheckAudience(row, gs.Players) {
		return
	}
	switch gs.CurrentMenu {
	case ActionMenu:
		DrawActionMenu(row)
	case TargetMenu:
		DrawTargetMenu(row, gs)
	case ChallengeMenu:
		DrawChallengeMenu(row)
	case BlockMenu:
		DrawBlockMenu(row, gs)
	case RevealMenu:
		DrawRevealMenu(row, gs.Players)
	case LossMenu:
		DrawLossMenu(row, gs.Players)
	case ReturnTwo:
		DrawReturnTwo(row, gs.Players)
	case ReturnOne:
		DrawReturnOne(row, gs.Players)
	default:
		panic("Unreachable code! (DrawMenu)")
	}
}

func DrawActionMenu(row *int) {
	AddStr(*row, 5, "[1] Income (+1 coin)")
	*row++
	AddStr(*row, 5, fmt.Sprintf("[2] Foreign Aid (+2 coins; blocked by %s)", Duke.Full()))
	*row++
	AddStr(*row, 5, "[3] Coup (-7 coins; target loses influence)")
	*row++
	AddStr(*row, 5, fmt.Sprintf("\033[90m[4] Assassinate (-3 coins; target loses influences; blocked by %s)", Contessa.Full()))
	*row++
	AddStr(*row, 5, fmt.Sprintf("\033[32m[5] Exchange (Draw 2 cards, then return 2 cards)\033[0m"))
	*row++
	AddStr(*row, 5, fmt.Sprintf("\033[36m[6] Steal (Take up to 2 coins from target; blocked by %s \033[36mor %s)", Ambassador.Full(), Captain.Full()))
	*row++
	AddStr(*row, 5, "\033[35m[7] Tax (+3 coins)\033[0m")
	*row += 2
	AddStr(*row, 1, "The time has come to act:")
	*row++
}

func DrawTargetMenu(row *int, gs *GameState) {
	AddStr(*row, 5, "And you will act upon?")
	*row++
	var counter int
	for _, p := range gs.Players {
		if p.IsCurrent || !p.IsAlive() {
			continue
		}
		counter++
		AddStr(*row, 5, fmt.Sprintf("[%d] %s", counter, p.Name))
		*row++
	}
}

func DrawChallengeMenu(row *int) { // Saved for later: pInfo []*Player) {
	AddStr(*row, 5, "Are they bluffing?")
	*row++
	AddStr(*row, 5, "[1] Challenge")
	*row++
	AddStr(*row, 5, "[0] Pass")
	*row++
}

func DrawBlockMenu(row *int, gs *GameState) { // Saved for later: pInfo []*Player) {}
	AddStr(*row, 5, "Will you block?")
	*row++
	if gs.CurrentAction != Steal {
		AddStr(*row, 5, "[1] Block")
		*row++
	} else {
		AddStr(*row, 5, fmt.Sprintf("[1] Block with %s", Ambassador.Full()))
		*row++
		AddStr(*row, 5, fmt.Sprintf("[2] Block with %s", Captain.Full()))
		*row++
	}
	AddStr(*row, 5, "[0] Pass")
	*row++
}

func DrawRevealMenu(row *int, pInfo []*Player) {
	var hand []Card
	for _, p := range pInfo {
		if p.Category == IsBot {
			continue
		}
		hand = p.CardsHeld
		break
	}
	AddStr(*row, 5, "Show the world who you really are!")
	*row++
	for i, card := range hand {
		AddStr(*row, 5, fmt.Sprintf("[%d] Reveal %s", i+1, card.Full()))
		*row++
	}
}

func DrawLossMenu(row *int, pInfo []*Player) {
	var hand []Card
	for _, p := range pInfo {
		if p.Category == IsBot {
			continue
		}
		hand = p.CardsHeld
		break
	}
	AddStr(*row, 5, "Who can you do without?")
	*row++
	for i, card := range hand {
		AddStr(*row, 5, fmt.Sprintf("[%d] Lose %s", i+1, card.Full()))
		*row++
	}
}

func DrawReturnTwo(row *int, pInfo []*Player) {
	var currentP *Player
	for _, p := range pInfo {
		if p.Category == IsBot {
			continue
		}
		currentP = p
		break
	}
	AddStr(*row, 5, "Who do you no longer need? (Returned 0 of 2)")
	*row++
	for i, card := range currentP.CardsHeld {
		AddStr(*row, 5, fmt.Sprintf("[%d] Return %s", i+1, card.Full()))
		*row++
	}
}

func DrawReturnOne(row *int, pInfo []*Player) {
	var currentP *Player
	for _, p := range pInfo {
		if p.Category == IsBot {
			continue
		}
		currentP = p
		break
	}
	AddStr(*row, 5, "Who do you no longer need? (Returned 1 of 2)")
	*row++
	for i, card := range currentP.CardsHeld {
		AddStr(*row, 5, fmt.Sprintf("[%d] Return %s", i+1, card.Full()))
		*row++
	}
	AddStr(*row, 5, fmt.Sprint("[0] Cancel"))
	*row++
}

// Types for InputHandler returns.

type GameUpdate interface {
	Update(gs *GameState)
}

type ActionSelection struct {
	ActionNumber int
}

func (as ActionSelection) Update(gs *GameState) {
	var logStr string
	gs.CurrentAction = Action(as.ActionNumber)
	if gs.CurrentAction == Income {
		gs.CurrentPlayer.AddCoins(1)
		logStr = fmt.Sprintf("[A] %s gains Income (+1 coin)", gs.CurrentPlayer.Name)
		gs.ActionLog.Enqueue(logStr)
		AdvanceState(gs, false, false)
		return
	}
	logStr = fmt.Sprintf("[A] %s intends to use %s....", gs.CurrentPlayer.Name, gs.CurrentAction)
	gs.ActionLog.Enqueue(logStr)
	AdvanceState(gs, false, false)
}

type TargetSelection struct {
	PlayerIndex int
}

func (ts TargetSelection) Update(gs *GameState) {
	var logStr string
	gs.TargetPlayer = gs.Players[ts.PlayerIndex]
	gs.TargetPlayer.IsTarget = true
	logStr = fmt.Sprintf("    ...targeting %s", gs.TargetPlayer.Name)
	gs.ActionLog.Enqueue(logStr)
	AdvanceState(gs, false, false)
}

type ChallengeSelection struct {
	PlayerIndex int
	Decision    bool
}

func (cs ChallengeSelection) Update(gs *GameState) {
	var logStr string
	gs.Players[cs.PlayerIndex].ChoiceMade = true
	if !cs.Decision {
		gs.Passes++
	}
	if !cs.Decision && gs.Passes < len(gs.FocusedPlayers) {
		return
	}
	if gs.Passes >= len(gs.FocusedPlayers) {
		AdvanceState(gs, false, false)
		return
	}
	gs.Challenger = gs.Players[cs.PlayerIndex]
	gs.Challenger.IsChallenger = true
	logStr = fmt.Sprintf("[!] %s is challenging!", gs.Challenger.Name)
	gs.ActionLog.Enqueue(logStr)
	AdvanceState(gs, false, false)
}

type BlockSelection struct {
	PlayerIndex int
	Decision    int
}

func (bs BlockSelection) Update(gs *GameState) {
	var logStr string
	gs.Players[bs.PlayerIndex].ChoiceMade = true
	if bs.Decision == 0 {
		gs.Passes++
	}
	if bs.Decision == 0 && gs.Passes < len(gs.FocusedPlayers) {
		return
	}
	if gs.Passes >= len(gs.FocusedPlayers) {
		if gs.CurrentAction == ForeignAid {
			gs.CurrentPlayer.AddCoins(2)
			logStr = fmt.Sprintf("    %s uses Foreign Aid (+2 coins)", gs.CurrentPlayer.Name)
			gs.ActionLog.Enqueue(logStr)
		}
		AdvanceState(gs, false, false)
		return
	}
	gs.Blocker = gs.Players[bs.PlayerIndex]
	gs.Blocker.IsBlocker = true
	if bs.Decision == 1 && gs.CurrentAction == Steal {
		gs.BlockType = Ambassador
		logStr = fmt.Sprintf("[B] %s is blocking with %s!", gs.Blocker.Name, Ambassador.Full())
	} else if bs.Decision == 2 && gs.CurrentAction == Steal {
		gs.BlockType = Captain
		logStr = fmt.Sprintf("[B] %s is blocking with %s!", gs.Blocker.Name, Captain.Full())
	} else {
		logStr = fmt.Sprintf("[B] %s is blocking!", gs.Blocker.Name)
	}

	gs.ActionLog.Enqueue(logStr)
	AdvanceState(gs, false, false)
}

type RevealSelection struct {
	CardIndex int
}

func (rs RevealSelection) Update(gs *GameState) {
	var logStr string
	successful := true
	focus := gs.FocusedPlayers[0]
	revealed := focus.RevealCard(rs.CardIndex)
	gs.ActionLog.Enqueue(fmt.Sprintf("    %s reveals... ", focus.Name))

	switch gs.CurrentAction {
	case ForeignAid:
		if revealed == Duke {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else {
			successful = true
			focus.LoseCard(rs.CardIndex)
			logStr = fmt.Sprintf("[X] ...%s: challenge succeeds! %s loses the %s", revealed.Full(), focus.Name, revealed.Full())
		}
	case Assassinate:
		if gs.Blocker == nil && revealed == Assassin {
			successful = false
		} else if gs.Blocker != nil && revealed == Contessa {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else {
			successful = true
			focus.LoseCard(rs.CardIndex)
			logStr = fmt.Sprintf("[X] ...%s: challenge succeeds! %s loses the %s", revealed.Full(), focus.Name, revealed.Full())
		}
	case Exchange:
		if revealed == Ambassador {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else {
			successful = true
			focus.LoseCard(rs.CardIndex)
			logStr = fmt.Sprintf("[X] ...%s: challenge succeeds! %s loses the %s", revealed.Full(), focus.Name, revealed.Full())
		}
	case Steal:
		if revealed == Ambassador && gs.BlockType == Ambassador {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else if revealed == Captain && gs.BlockType == Captain {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else {
			successful = true
			focus.LoseCard(rs.CardIndex)
			logStr = fmt.Sprintf("[X] ...%s: challenge succeeds! %s loses the %s", revealed.Full(), focus.Name, revealed.Full())
		}
	case Tax:
		if revealed == Duke {
			successful = false
			logStr = fmt.Sprintf("    ...%s: challenge failed! %s swaps %s for something new", revealed.Full(), focus.Name, revealed.Full())
		} else {
			successful = true
			focus.LoseCard(rs.CardIndex)
			logStr = fmt.Sprintf("[X] ...%s: challenge succeeds! %s loses the %s", revealed.Full(), focus.Name, revealed.Full())
		}
	}
	if !successful {
		gs.Deck = append(gs.Deck, focus.CardsHeld[rs.CardIndex])
		focus.CardsHeld = slices.Delete(focus.CardsHeld, rs.CardIndex, rs.CardIndex+1)
		idx := gs.RNG.IntN(len(gs.Deck))
		focus.CardsHeld = append(focus.CardsHeld, gs.Deck[idx])
	}
	gs.ActionLog.Enqueue(logStr)

	AdvanceState(gs, successful, false)
}

type LossSelection struct {
	CardIndex int
}

func (ls LossSelection) Update(gs *GameState) {
	gs.FocusedPlayers[0].LoseCard(ls.CardIndex)
	lostLen := len(gs.FocusedPlayers[0].CardsLost)
	lost := gs.FocusedPlayers[0].CardsLost[lostLen-1]
	gs.ActionLog.Enqueue(fmt.Sprintf("[X] %s loses %s", gs.FocusedPlayers[0].Name, lost.Full()))
	AdvanceState(gs, false, false)
}

type ReturnTwoSelection struct {
	CardIndex int
}

func (rt ReturnTwoSelection) Update(gs *GameState) {
	var hand = gs.CurrentPlayer.CardsHeld
	if gs.DrawnCards == nil {
		gs.ActionLog.Enqueue(fmt.Sprintf("    %s draws 2 cards...", gs.CurrentPlayer.Name))
		for range 2 {
			idx := gs.RNG.IntN(len(gs.Deck))
			gs.DrawnCards = append(gs.DrawnCards, gs.Deck[idx])
			gs.Deck = slices.Delete(gs.Deck, idx, idx+1)
		}
		gs.CurrentPlayer.CardsHeld = slices.Concat(hand, gs.DrawnCards)
		hand = gs.CurrentPlayer.CardsHeld
	}
	gs.ReturnedCards = append(gs.ReturnedCards, hand[rt.CardIndex])
	gs.CurrentPlayer.CardsHeld = slices.Delete(hand, rt.CardIndex, rt.CardIndex+1)
	AdvanceState(gs, false, false)
}

type ReturnOneSelection struct {
	CardIndex int
}

func (ro ReturnOneSelection) Update(gs *GameState) {
	var hand = gs.CurrentPlayer.CardsHeld
	if ro.CardIndex == 0 {
		gs.CurrentPlayer.CardsHeld = append(hand, gs.ReturnedCards[0])
		gs.ReturnedCards = nil
		AdvanceState(gs, false, true)
		return
	}
	gs.ReturnedCards = append(gs.ReturnedCards, hand[ro.CardIndex])
	gs.CurrentPlayer.CardsHeld = slices.Delete(hand, ro.CardIndex, ro.CardIndex+1)
	for _, card := range gs.ReturnedCards {
		gs.Deck = append(gs.Deck, card)
	}
	gs.ActionLog.Enqueue(fmt.Sprint("    ...and returns 2 cards."))
	AdvanceState(gs, false, false)
}

// Input Functions

// GetBotInputs continuously generates pseudorandom input data, verifies
// it is appropriate for the situation, and sends the information as a
// rune via the passed output channel for processing.
func GetBotInputs(game *GameState, i int, chanOut chan<- rune, chanErr chan<- error) {
	defer RoutePanic(chanErr, "GetBotInputs")
	duration := time.Duration(1200 + game.RNG.IntN(2000))
	time.Sleep(duration * time.Millisecond)

	for {
		switch game.CurrentMenu {
		case ActionMenu:
			key := '1' + rune(game.RNG.IntN(7))
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		case TargetMenu:
			key := '1' + rune(game.RNG.IntN(len(game.LivingPlayers)-1))
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		case ChallengeMenu, BlockMenu:
			key := '0'
			chance := game.RNG.IntN(100)
			if chance < 15 { key = '1' }
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		case LossMenu, RevealMenu:
			key := '1' + rune(game.RNG.IntN(len(game.Players[i].CardsHeld)))
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		case ReturnTwo:
			key := '1' + rune(game.RNG.IntN(len(game.Players[i].CardsHeld)))
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		case ReturnOne:
			key := '0' + rune(game.RNG.IntN(1 + len(game.Players[i].CardsHeld)))
			ok, err := verifyInput(game, key, i)
			if err != nil {
				chanErr <- err
				return
			}
			if !ok {
				continue
			}
			chanOut <- key
			return
		}
	}
}

// GetClientInputs continuously reads keystrokes from stdin, verfies there
// are no serious errors, and then sends the byte for processing via the
// the passed output channel as a rune.
func GetClientInputs(game *GameState, i int, chanOut chan<- rune, chanErr chan<- error) {
	defer RoutePanic(chanErr, "GetClientInputs")

	var buf = make([]byte, 1)
	for {
		log.Printf("pre-buffer: %v", buf[:])
		n, err := os.Stdin.Read(buf)
		log.Printf("buffer: %v", buf[:])
		if err != nil {
			chanErr <- err
			return
		}
		if n > 1 {
			chanErr <- fmt.Errorf("Invalid input (> 1 byte)")
			return
		}
		if buf[0] == 'q' {
			chanErr <- UserQuitError
			return
		}
		ok, err := verifyInput(game, rune(buf[0]), i)
		if err != nil {
			chanErr <- err
			return
		}
		if !ok {
			continue
		}
		break
	}
	chanOut <- rune(buf[0])
}

func AsyncInputHandler(
	game *GameState,
	chansIn []chan rune,
	chanOut chan<- GameUpdate,
	chanErr chan<- error,
) {
	defer RoutePanic(chanErr, "AsyncInputHandler")

	for {
		select {
		case pSig := <-chansIn[0]:
			if !game.Players[0].IsFocused || game.Players[0].ChoiceMade {
				continue
			}
			log.Printf("channel 0")
			chanOut <- getUpdate(game, pSig, 0)
		case pSig := <-chansIn[1]:
			if !game.Players[1].IsFocused || game.Players[1].ChoiceMade {
				continue
			}
			log.Printf("channel 1")
			chanOut <- getUpdate(game, pSig, 1)
		case pSig := <-chansIn[2]:
			if !game.Players[2].IsFocused || game.Players[2].ChoiceMade {
				continue
			}
			log.Printf("channel 2")
			chanOut <- getUpdate(game, pSig, 2)
		case pSig := <-chansIn[3]:
			if !game.Players[3].IsFocused || game.Players[3].ChoiceMade {
				continue
			}
			log.Printf("channel 3")
			chanOut <- getUpdate(game, pSig, 3)
		case pSig := <-chansIn[4]:
			if !game.Players[4].IsFocused || game.Players[4].ChoiceMade {
				continue
			}
			log.Printf("channel 4")
			chanOut <- getUpdate(game, pSig, 4)
		}
	}
}

// verifyInput performs simple checks on input runes to ensure that the
// input is valid for the state of the game and the source player.
func verifyInput(game *GameState, signal rune, pIdx int) (bool, error) {
	switch game.CurrentMenu {
	case ActionMenu:
		if !strings.Contains("1234567", string(signal)) {
			return false, nil
		}
		// Check if Coup (3) or Assassinate (4) can be afforded
		if signal == '3' && game.Players[pIdx].Coins < 7 {
			return false, nil
		}
		if signal == '4' && game.Players[pIdx].Coins < 3 {
			return false, nil
		}
		// Check that Coup (3) has been selected if >10 coins
		if signal != '3' && game.Players[pIdx].Coins > 10 {
			return false, nil
		}
		return true, nil
	case TargetMenu:
		if signal == '0' || int(signal-'0') > len(game.LivingPlayers)-1 {
			return false, nil
		}
		return true, nil
	case ChallengeMenu:
		if !strings.Contains("01", string(signal)) {
			return false, nil
		}
		return true, nil
	case BlockMenu:
		if game.CurrentAction == Steal && !strings.Contains("012", string(signal)) {
			return false, nil
		}
		if !strings.Contains("01", string(signal)) {
			return false, nil
		}
		return true, nil
	case RevealMenu, LossMenu:
		if !strings.Contains("12", string(signal)) {
			return false, nil
		}
		return true, nil
	case ReturnTwo:
		if int(signal-'0') > len(game.CurrentPlayer.CardsHeld) {
			return false, nil
		}
		if int(signal-'0') < 1 {
			return false, nil
		}
		return true, nil
	case ReturnOne:
		if int(signal-'0') > len(game.CurrentPlayer.CardsHeld) {
			return false, nil
		}
		if int(signal-'0') < 0 {
			return false, nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("Unreachable code! (verifyInput)")
	}
}

// Creates a GameUpdate object based on the provided signal.
//
// IMPORTANT: This function does not perform any verification and must
// be used in conjunction with verifyInput or similar.
func getUpdate(game *GameState, signal rune, pIdx int) GameUpdate {
	switch game.CurrentMenu {
	case ActionMenu:
		choice := int(signal - '0')
		return ActionSelection{ActionNumber: choice}
	case TargetMenu:
		choice := int(signal - '0')
		var n int
		for i, p := range game.LivingPlayers {
			if p.IsCurrent {
				continue
			}
			n++
			if n != choice {
				continue
			}
			choice = i
			break
		}
		return TargetSelection{PlayerIndex: choice}
	case ChallengeMenu:
		var isChallenging bool
		choice := int(signal - '0')
		if choice == 1 {
			isChallenging = true
		}
		return ChallengeSelection{PlayerIndex: pIdx, Decision: isChallenging}
	case BlockMenu:
		choice := int(signal - '0')
		return BlockSelection{PlayerIndex: pIdx, Decision: choice}
	case LossMenu:
		choice := int(signal-'0') - 1
		return LossSelection{CardIndex: choice}
	case RevealMenu:
		choice := int(signal-'0') - 1
		return RevealSelection{CardIndex: choice}
	case ReturnTwo:
		choice := int(signal-'0') - 1
		return ReturnTwoSelection{CardIndex: choice}
	case ReturnOne:
		choice := int(signal-'0') - 1
		return ReturnOneSelection{CardIndex: choice}
	default:
		panic("Unreachable code! (getUpdate)")
	}
}

type Menu int

const (
	ActionMenu Menu = iota
	TargetMenu
	ChallengeMenu
	BlockMenu
	RevealMenu
	LossMenu
	ReturnTwo
	ReturnOne
)

type Card int

const (
	Ambassador Card = iota
	Assassin
	Captain
	Contessa
	Duke
)

var CardName = map[Card]string{
	Ambassador: "Ambassador",
	Assassin:   "Assassin",
	Captain:    "Captain",
	Contessa:   "Contessa",
	Duke:       "Duke",
}

var CardColor = map[Card]string{
	Ambassador: "\033[32m",
	Assassin:   "\033[37m",
	Captain:    "\033[36m",
	Contessa:   "\033[31m",
	Duke:       "\033[35m",
}

func (c Card) Full() string {
	return fmt.Sprintf("%s%s\033[0m", CardColor[c], CardName[c])
}

func (c Card) Short() string {
	return fmt.Sprintf("%s%s\033[0m", CardColor[c], strings.ToUpper(CardName[c][:3]))
}

type Action int

const (
	NoAction Action = iota
	Income
	ForeignAid
	Coup
	Assassinate
	Exchange
	Steal
	Tax
)

var actionName = map[Action]string{
	NoAction:    "Nothing",
	Income:      "Income",
	ForeignAid:  "Foreign Aid",
	Coup:        "Coup",
	Assassinate: "\033[37mAssassinate",
	Exchange:    "\033[32mExchange",
	Steal:       "\033[36mSteal",
	Tax:         "\033[35mTax",
}

func (a Action) String() string {
	return fmt.Sprintf("%s\033[0m", actionName[a])
}

type PlayerCategory int

const (
	IsBot PlayerCategory = iota
	IsLocal
	IsRemote
)

type Player struct {
	Name         string
	Coins        int
	CardsHeld    []Card
	CardsLost    []Card
	IsCurrent    bool
	IsTarget     bool
	IsChallenger bool
	IsBlocker    bool
	IsFocused    bool
	ChoiceMade   bool
	Category     PlayerCategory
}

func (p *Player) IsAlive() bool {
	if len(p.CardsHeld) == 0 && len(p.CardsLost) == 2 {
		return false
	}
	return true
}

func (p *Player) AddCoins(n int)  { p.Coins += n }
func (p *Player) LoseCoins(n int) { p.Coins -= n }
func (p *Player) RevealCard(i int) Card {
	return p.CardsHeld[i]
}
func (p *Player) LoseCard(i int) {
	lost := p.CardsHeld[i]
	p.CardsHeld = slices.Delete(p.CardsHeld, i, i+1)
	p.CardsLost = append(p.CardsLost, lost)
}

func NewPlayer(name string, cat PlayerCategory) *Player {
	return &Player{Name: name, Coins: 2, Category: cat}
}

// GameState Type and Functions

type GameState struct {
	Players           []*Player
	FocusedPlayers    []*Player
	LivingPlayers     []*Player
	CurrentMenu       Menu
	CurrentAction     Action
	CurrentPlayer     *Player
	TargetPlayer      *Player
	Challenger        *Player
	Blocker           *Player
	BlockType         Card
	Passes            int
	AssassinPostBlock bool
	DrawnCards        []Card
	ReturnedCards     []Card
	Deck              []Card
	RNG               *rand.Rand
	ActionLog         *ActionLog
}

func (gs *GameState) AdvanceTurn() {
	log.Print(gs.CurrentAction)
	gs.CurrentMenu = ActionMenu
	gs.CurrentAction = NoAction
	gs.FocusedPlayers = nil
	gs.TargetPlayer = nil
	gs.Challenger = nil
	gs.Blocker = nil
	gs.BlockType = Card(0)
	gs.Passes = 0
	gs.DrawnCards = nil
	gs.ReturnedCards = nil
	gs.AssassinPostBlock = false
	gs.UpdateLivingPlayers()
	for i, p := range gs.LivingPlayers {
		if !p.IsCurrent {
			continue
		}
		idx := (i + 1) % len(gs.LivingPlayers)
		p.IsCurrent = false
		gs.CurrentPlayer = gs.LivingPlayers[idx]
		gs.LivingPlayers[idx].IsCurrent = true
		return
	}
}

func NewGameState(
	players []*Player,
	deck []Card,
	rng *rand.Rand,
	actionLog *ActionLog,
) *GameState {
	newState := GameState{
		Players:       players,
		LivingPlayers: players,
		CurrentMenu:   ActionMenu,
		CurrentAction: NoAction,
		Deck:          deck,
		RNG:           rng,
		ActionLog:     actionLog,
	}
	for i, p := range players {
		if !p.IsCurrent {
			continue
		}
		newState.CurrentPlayer = players[i]
	}
	return &newState
}

func (gs *GameState) UpdateLivingPlayers() {
	var update []*Player
	for _, p := range gs.LivingPlayers {
		if !p.IsAlive() {
			continue
		}
		update = append(update, p)
	}
	gs.LivingPlayers = update
}

func (gs *GameState) UpdateFocusedPlayers() {
	gs.FocusedPlayers = nil
	for _, p := range gs.Players {
		p.IsFocused = false
	}
	switch gs.CurrentMenu {
	case ChallengeMenu:
		// Skip blocker if there is one, otherwise skip current
		if gs.Blocker != nil {
			for _, p := range gs.LivingPlayers {
				if p.IsBlocker {
					continue
				}
				p.IsFocused = true
				gs.FocusedPlayers = append(gs.FocusedPlayers, p)
			}
		} else {
			for _, p := range gs.LivingPlayers {
				if p.IsCurrent {
					continue
				}
				p.IsFocused = true
				gs.FocusedPlayers = append(gs.FocusedPlayers, p)
			}
		}
	case BlockMenu:
		// Foreign Aid needs different blocking behaviour
		if gs.CurrentAction != ForeignAid {
			gs.TargetPlayer.IsFocused = true
			gs.FocusedPlayers = append(gs.FocusedPlayers, gs.TargetPlayer)
			return
		}
		for _, p := range gs.LivingPlayers {
			if p.IsCurrent {
				continue
			}
			p.IsFocused = true
			gs.FocusedPlayers = append(gs.FocusedPlayers, p)
		}
	case RevealMenu:
		if gs.Blocker != nil {
			gs.Blocker.IsFocused = true
			gs.FocusedPlayers = append(gs.FocusedPlayers, gs.Blocker)
			return
		}
		gs.CurrentPlayer.IsFocused = true
		gs.FocusedPlayers = append(gs.FocusedPlayers, gs.CurrentPlayer)
	case LossMenu:
		if gs.TargetPlayer != nil {
			gs.TargetPlayer.IsFocused = true
		}
		gs.FocusedPlayers = append(gs.FocusedPlayers, gs.TargetPlayer)
	default:
		gs.CurrentPlayer.IsFocused = true
		gs.FocusedPlayers = append(gs.FocusedPlayers, gs.CurrentPlayer)
	}
}

func AdvanceState(gs *GameState, challengeSuccessful bool, cancel bool) {
	for _, p := range gs.Players {
		p.ChoiceMade = false
	}
	gs.UpdateLivingPlayers()
	switch gs.CurrentAction {
	case Income:
		gs.AdvanceTurn()
	case ForeignAid:
		advanceForeignAid(gs, challengeSuccessful)
	case Coup:
		advanceCoup(gs)
	case Assassinate:
		advanceAssassinate(gs, challengeSuccessful)
	case Exchange:
		advanceExchange(gs, challengeSuccessful, cancel)
	case Steal:
		advanceSteal(gs, challengeSuccessful)
	case Tax:
		advanceTax(gs, challengeSuccessful)
	default:
		panic("Unreachable code! (AdvanceState)")
	}
}

func advanceForeignAid(gs *GameState, challengeSuccessful bool) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = BlockMenu
	case BlockMenu:
		if gs.Blocker != nil {
			gs.CurrentMenu = ChallengeMenu
			return
		}
		gs.AdvanceTurn()
	case ChallengeMenu:
		if gs.Challenger != nil {
			gs.CurrentMenu = RevealMenu
			return
		}
		gs.AdvanceTurn()
	case RevealMenu:
		if challengeSuccessful {
			gs.CurrentPlayer.AddCoins(2)
			gs.AdvanceTurn()
			return
		}
		gs.AdvanceTurn()
	default:
		panic("Unreachable code! (advanceForeignAid)")
	}
}

func advanceCoup(gs *GameState) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = TargetMenu
	case TargetMenu:
		gs.CurrentMenu = LossMenu
	case LossMenu:
		gs.AdvanceTurn()
	default:
		panic("Unreachable code! (advanceCoup)")
	}
}

func advanceAssassinate(gs *GameState, challengeSuccessful bool) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = TargetMenu
	case TargetMenu:
		gs.CurrentMenu = ChallengeMenu
	case ChallengeMenu:
		if gs.Challenger != nil {
			gs.CurrentMenu = RevealMenu
			return
		}
		if gs.Blocker != nil {
			gs.AdvanceTurn()
			return
		}
		gs.CurrentMenu = BlockMenu
	case BlockMenu:
		gs.AssassinPostBlock = true
		if gs.Blocker != nil {
			gs.CurrentMenu = ChallengeMenu
			return
		}
		gs.CurrentMenu = LossMenu
	case RevealMenu:
		if challengeSuccessful && gs.Blocker == nil {
			gs.AdvanceTurn()
			return
		}
		if !challengeSuccessful && gs.Blocker == nil {
			gs.CurrentMenu = LossMenu
			return
		}
		if challengeSuccessful && gs.Blocker != nil {
			gs.CurrentMenu = LossMenu
			return
		}
		gs.CurrentMenu = LossMenu
	case LossMenu:
		if gs.Blocker != nil || gs.AssassinPostBlock {
			gs.AdvanceTurn()
			return
		}
		gs.CurrentMenu = BlockMenu
	default:
		panic("Unreachable code! (advanceAssassinate)")
	}
}

func advanceExchange(gs *GameState, challengeSuccessful bool, cancel bool) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = ChallengeMenu
	case ChallengeMenu:
		if gs.Challenger != nil {
			gs.CurrentMenu = RevealMenu
			return
		}
		gs.CurrentMenu = ReturnTwo
	case RevealMenu:
		if challengeSuccessful {
			gs.AdvanceTurn()
			return
		}
		gs.CurrentMenu = ReturnTwo
	case ReturnTwo:
		gs.CurrentMenu = ReturnOne
	case ReturnOne:
		if cancel {
			gs.CurrentMenu = ReturnTwo
			return
		}
		gs.AdvanceTurn()
	default:
		panic("Unreachable code! (advanceExchange)")
	}
}

func advanceSteal(gs *GameState, challengeSuccessful bool) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = TargetMenu
	case TargetMenu:
		gs.CurrentMenu = ChallengeMenu
	case ChallengeMenu:
		if gs.Challenger != nil {
			gs.CurrentMenu = RevealMenu
			return
		}
		if gs.Blocker != nil {
			gs.AdvanceTurn()
			return
		}
		gs.CurrentMenu = BlockMenu
	case BlockMenu:
		if gs.Blocker != nil {
			gs.CurrentMenu = ChallengeMenu
			return
		}
		stolen := min(2, gs.TargetPlayer.Coins)
		gs.TargetPlayer.LoseCoins(stolen)
		gs.CurrentPlayer.AddCoins(stolen)
		gs.AdvanceTurn()
	case RevealMenu:
		if challengeSuccessful && gs.Blocker == nil {
			gs.AdvanceTurn()
			return
		}
		if challengeSuccessful && gs.Blocker != nil {
			stolen := min(2, gs.TargetPlayer.Coins)
			gs.TargetPlayer.LoseCoins(stolen)
			gs.CurrentPlayer.AddCoins(stolen)
			gs.AdvanceTurn()
			return
		}
		if !challengeSuccessful && gs.Blocker == nil {
			gs.CurrentMenu = LossMenu
			return
		}
		gs.CurrentMenu = LossMenu
	case LossMenu:
		if gs.Blocker == nil {
			gs.CurrentMenu = BlockMenu
			return
		}
		gs.AdvanceTurn()
	default:
		panic("Unreachable code! (advanceSteal)")
	}
}

func advanceTax(gs *GameState, challengeSuccessful bool) {
	switch gs.CurrentMenu {
	case ActionMenu:
		gs.CurrentMenu = ChallengeMenu
	case ChallengeMenu:
		if gs.Challenger != nil {
			gs.CurrentMenu = RevealMenu
			return
		}
		gs.CurrentPlayer.AddCoins(3)
		gs.ActionLog.Enqueue(fmt.Sprintf("    ... %s succeeds! +3 coins", gs.CurrentPlayer.Name))
		gs.AdvanceTurn()
	case RevealMenu:
		if challengeSuccessful {
			gs.AdvanceTurn()
			return
		}
	case LossMenu:
		gs.ActionLog.Enqueue(fmt.Sprintf("    %s performs %s (+3 coins)", gs.CurrentPlayer.Name, Tax))
		gs.CurrentPlayer.AddCoins(3)
		gs.AdvanceTurn()
	default:
		panic("Unreachable code! (advanceTax)")
	}
}

func InitialiseGame() (*GameState, *time.Ticker, []chan rune, chan GameUpdate, chan error) {
	var rng = rand.New(rand.NewPCG(782, 782)) // REPLACE: random seed before release
	var Clock = time.NewTicker(250 * time.Millisecond)
	var Players []*Player
	var names = []string{"You", "Alice", "Bob"}
	var deck []Card
	var AllCardTypes = [5]Card{Ambassador, Assassin, Captain, Contessa, Duke}
	var inputChannels = make([]chan rune, 5)
	var actionLog = NewActionLog(10)

	// Create Players
	for i, name := range names {
		p := NewPlayer(name, IsBot)
		if i == 0 {
			p.IsCurrent = true
			p.Category = IsLocal
		}
		Players = append(Players, p)
	}

	// Create Deck
	for _, c := range AllCardTypes {
		for range 3 {
			deck = append(deck, c)
		}
	}

	// Deal Cards
	for _, p := range Players {
		for range 2 {
			n := rng.IntN(len(deck))
			dealt := deck[n]
			deck = slices.Delete(deck, n, n+1)
			p.CardsHeld = append(p.CardsHeld, dealt)
		}
	}
	Game := NewGameState(Players, deck, rng, actionLog)

	// Setup i/o channels
	outputChannel := make(chan GameUpdate)
	errorChannel := make(chan error)
	for i := range len(Game.Players) {
		inputChannels[i] = make(chan rune)
	}
	return Game, Clock, inputChannels, outputChannel, errorChannel
}

func EnterGame() error {
	var logFile, _ = os.Create("debug.log")
	log.SetOutput(logFile)
	// Initialise
	Game, Clock, inputChannels, outputChannel, errorChannel := InitialiseGame()
	// Spawn a single thread to handle input
	go AsyncInputHandler(Game, inputChannels, outputChannel, errorChannel)

	// Main Loop
	for {
		// Draw to terminal
		Game.UpdateFocusedPlayers()
		DrawDisplay(Game)
		// Spawn one thread for each focused player and listen for input
		for i, p := range Game.Players {
			if !p.IsFocused || p.ChoiceMade {
				log.Printf("Not Focused: %s", p.Name)
				continue
			}
			log.Printf("Focused: %s", p.Name)
			if p.Category != IsBot {
				go GetClientInputs(Game, i, inputChannels[i], errorChannel)
				continue
			}
			go GetBotInputs(Game, i, inputChannels[i], errorChannel)
		}

		// Update game state based on that input
		// Wait for update ticker or error out
		select {
		case err := <-errorChannel:
			if errors.Is(err, UserQuitError) {
				return nil
			}
			return err
		case UpdateInfo := <-outputChannel:
			log.Printf("menu: %v; %v; action: %v", Game.CurrentMenu, UpdateInfo, Game.CurrentAction)
			UpdateInfo.Update(Game)
		}
		<-Clock.C
	}
}

func main() {
	err := WrapDisplay(EnterGame)
	if err != nil {
		panic(err)
	}
}

// Scratch

type Queue[E any] interface {
	Enqueue(E)
	Dequeue() E
}

type ActionLog struct {
	Items  []string
	Length int
}

func (al *ActionLog) Enqueue(item string) {
	if len(al.Items) >= al.Length {
		al.Items = al.Items[1:]
	}
	al.Items = append(al.Items, item)
}

func (al *ActionLog) Dequeue() string {
	popped := al.Items[0]
	al.Items = al.Items[1:]
	return popped
}

func NewActionLog(length int) *ActionLog {
	aLog := ActionLog{Length: length}
	return &aLog
}
