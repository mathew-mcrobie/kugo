package input

import (
	"context"
	"fmt"
	"kugo/game"
	"math/rand/v2"
	"os"
	"slices"
	"time"
)

var inputSubHandlers = map[game.Phase]func(*InputHandler) *game.InputData{
	game.SelectAction:    (*InputHandler).selectAction,
	game.SelectTarget:    (*InputHandler).selectTarget,
	game.MakeChallenge:   (*InputHandler).makeChallenge,
	game.ChallengeReveal: (*InputHandler).selectCard,
	game.ChallengeLoss:   (*InputHandler).selectCard,
	game.MakeBlock:       (*InputHandler).makeBlock,
	game.ChallengeBlock:  (*InputHandler).makeChallenge,
	game.BlockReveal:     (*InputHandler).selectCard,
	game.BlockLoss:       (*InputHandler).selectCard,
	game.ResolveAction:   (*InputHandler).resolveAction,
	game.ExchangeMiddle:  (*InputHandler).exchangeMiddle,
	game.ExchangeFinal:   (*InputHandler).exchangeFinal,
	game.EndGame:         (*InputHandler).endGame,
}

type InputHandler struct {
	phase         game.Phase
	action        game.Action
	rng           *rand.Rand
	allPlayers	  []*game.Player
	activePlayers []*game.Player
	validTargets  []*game.Player
	blockType     game.Card
	PlayerChans   []chan rune
	chanErr       chan error
	inputData     *game.InputData
}

// NewInputHandler is called during initialization to set up the InputHandler.
func NewInputHandler(players []*game.Player, chanErr chan error) *InputHandler {
	var PlayerChans []chan rune
	for range len(players) {
		PlayerChans = append(PlayerChans, make(chan rune))
	}
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	ih := InputHandler{
		rng:         rng,
		allPlayers:  players,
		PlayerChans: PlayerChans,
		chanErr:     chanErr,
	}
	return &ih
}

func (ih *InputHandler) UpdateStateData(data *game.StateData) {
	ih.activePlayers = data.ActivePlayers
	ih.validTargets = data.ValidTargets
	ih.blockType = data.BlockType
	ih.phase = data.State.Phase
	ih.action = data.State.Action
}

// GetInputData is the core method of InputHandler. When called it deplyoys an
// an input channel for each active player. It then calls a phase dependent
// subhandler which handles all the state specific logic needed to get valid
// input. This subhandler concludes by clearing the player data before
// returning an InputData struct to pass to the controller to update the state
// machine.
func (ih *InputHandler) GetInputData() *game.InputData {
	defer ih.RecoverPanic("Panic handled by GetInputData")
	// select appropriate handler for phase
	handler := inputSubHandlers[ih.phase]

	// run handler to create input data to pass to controller
	inputData := handler(ih)

	// clear player data and return the input data
	ih.clearData()
	return inputData
}

// clearData wipes the requested player data, but does not change state data.
// As a result, phase and action held by the input handler cannot be used
// reliably unless it was updated by the controller before use.
func (ih *InputHandler) clearData() {
	ih.activePlayers = nil
	ih.validTargets = nil
	ih.blockType = game.NoCard
}

// getSignal is a useful helper function that makes up the core functionality
// of all the handler functions. It continuously checks a player's input channel
// (given by playerIndex), converts the sent rune to an int, and compares it to
// minVal and maxVal. If it is outside the inclusive range given by those two
// values it continues to loop for input. Once it gets a reply in range it
// returns the value to either be processed or sent directly to the controller.
func (ih *InputHandler) getSignal(minVal, maxVal int) (int, int) {
	for {
		select {
		case runeIn := <-ih.PlayerChans[0]:
			if ih.checkIfActive(0) && ih.checkSignal(runeIn, minVal, maxVal) {
				return int(runeIn - '0'), 0
			}
		case runeIn := <-ih.PlayerChans[1]:
			if ih.checkIfActive(1) && ih.checkSignal(runeIn, minVal, maxVal) {
				return int(runeIn - '0'), 1
			}
		case runeIn := <-ih.PlayerChans[2]:
			if ih.checkIfActive(2) && ih.checkSignal(runeIn, minVal, maxVal) {
				return int(runeIn - '0'), 2
			}
			/*
				case runeIn := <-ih.PlayerChans[3]:
					if ih.checkIfActive(3) && ih.checkSignal(runeIn, minVal, maxVal) {
						return int(runeIn - '0'), 3
					}
				case runeIn := <-ih.PlayerChans[4]:
					if ih.checkIfActive(4) && ih.checkSignal(runeIn, minVal, maxVal) {
						return int(runeIn - '0'), 4
					}
				case runeIn := <-ih.PlayerChans[5]:
					if ih.checkIfActive(4) && ih.checkSignal(runeIn, minVal, maxVal) {
						return int(runeIn - '0'), 4
					}
			*/
		}
	}
}

func (ih *InputHandler) checkSignal(rawSig rune, minVal, maxVal int) bool {
	sig := int(rawSig - '0')
	if sig < minVal || sig > maxVal {
		return false
	}
	return true
}

func (ih *InputHandler) checkIfActive(pIdx int) bool {
	for _, p := range ih.activePlayers {
		if pIdx == p.Index {
			return true
		}
	}
	return false
}

func (ih *InputHandler) selectAction() *game.InputData {
	defer ih.resetResponseFlags()
	var current = ih.activePlayers[0]
	var pIdx = current.Index
	var minVal, maxVal = 1, 7
	if current.Coins >= 10 {
		minVal, maxVal = 3, 3
	}
	for {
		sig, _ := ih.getSignal(minVal, maxVal)
		if current.Coins < 7 && sig == 3 {
			continue
		}
		if current.Coins < 3 && sig == 4 {
			continue
		}
		ih.flagAsResponded(pIdx)
		return game.NewInputData(sig, pIdx)
	}
}

func (ih *InputHandler) selectTarget() *game.InputData {
	defer ih.resetResponseFlags()
	var pIdx = ih.activePlayers[0].Index
	sig, _ := ih.getSignal(0, len(ih.validTargets))
	// Because 0 means cancel, controller will need to subtract 1 from sig to
	// get correct player index.
	ih.flagAsResponded(pIdx)
	return game.NewInputData(sig, pIdx)
}

func (ih *InputHandler) makeChallenge() *game.InputData {
	// From an input perspective this is identical before and after blocks.
	defer ih.resetResponseFlags()
	var maxResponses = len(ih.activePlayers)
	for range maxResponses {
		sig, pIdx := ih.getSignal(0, 1)
		if sig == 1 {
			return game.NewInputData(sig, pIdx)
		}
		ih.flagAsResponded(pIdx)
	}
	return game.NewInputData(0, 0)
}

func (ih *InputHandler) flagAsResponded(pIdx int) {
	for _, p := range ih.activePlayers {
		if p.Index != pIdx {
			continue
		}
		p.Responded = true
		return
	}
}

func (ih *InputHandler) resetResponseFlags() {
	for _, p := range ih.activePlayers {
		p.Responded = false
	}
}

func (ih *InputHandler) selectCard() *game.InputData {
	defer ih.resetResponseFlags()
	// This can be reused for all the Reveal/Loss phases
	maxVal := len(ih.activePlayers[0].CardsHeld)
	sig, pIdx := ih.getSignal(1, maxVal)
	ih.flagAsResponded(pIdx)
	return game.NewInputData(sig-1, pIdx)
}

func (ih *InputHandler) makeBlock() *game.InputData {
	defer ih.resetResponseFlags()
	// Foreign Aid is the only action that can be blocked by multiple players.
	// Fortunately, this makes it identical to a challenge, so just reuse
	// the makeChallenge code.
	if ih.action == game.ForeignAid {
		return ih.makeChallenge()
	}
	// No need to spawn a bunch of threads with only one active player.
	// But do still need to check if the action is steal to account for the
	// additional option.
	var maxVal = 1
	if ih.action == game.Steal {
		maxVal = 2
	}
	sig, pIdx := ih.getSignal(0, maxVal)
	ih.flagAsResponded(pIdx)
	return game.NewInputData(sig, pIdx)
}

func (ih *InputHandler) resolveAction() *game.InputData {
	defer ih.resetResponseFlags()
	// Unfortunately this differs depending on action. Luckily we only have
	// to handle Coup/Assassinate (just selectCard) and Exchange, which just
	// needs punting to ExchangeMiddle, which is the same as just passing.
	switch ih.action {
	case game.Assassinate, game.Coup:
		return ih.selectCard()
	default:
		ih.flagAsResponded(0)
		return game.NewInputData(0, 0)
	}
}

func (ih *InputHandler) exchangeMiddle() *game.InputData {
	defer ih.resetResponseFlags()
	var handLength = len(ih.activePlayers[0].CardsHeld)
	sig, pIdx := ih.getSignal(1, handLength)
	ih.flagAsResponded(pIdx)
	return game.NewInputData(sig-1, pIdx)
}

func (ih *InputHandler) exchangeFinal() *game.InputData {
	defer ih.resetResponseFlags()
	var pIdx = ih.activePlayers[0].Index
	var handLength = len(ih.activePlayers[0].CardsHeld)
	sig, _ := ih.getSignal(0, handLength) // can cancel, so min is 0
	// Controller knows to subtract 1 from sig if sig != 0.
	ih.flagAsResponded(pIdx)
	return game.NewInputData(sig, pIdx)
}

func (ih *InputHandler) endGame() *game.InputData {
	ih.flagAsResponded(0)
	return game.NewInputData(0, 0)
}

// CreateBotInputStream uses the input handler's random number generator to
// produce semi-random behaviour from bots when they have no obvious choice
// to make. If the bot has the required card for a reveal then they will
// reveal it, and will always block if they have the required card. Otherwise
// they choose actions randomly from the available options. They challenge at
// a fixed percentage rate, given by challengeRate.
func (ih *InputHandler) CreateBotInputStream(ctx context.Context, outChan chan<- rune, pIdx int) {
	defer ih.RecoverPanic("Panic captured by CreateBotInputStream")
	var retryCounter int
	for {
		var n int
		var challengeRate = 20
		var waitTime = 3000 + ih.rng.IntN(2000)
		// wait for 3.0 - 5.0 seconds to allow humans to read the logs, but still
		// keep the game flowing quickly. Randomness makes it feel more like the
		// bot is thinking organically. Without this wait the response is immediate
		// which is very disorienting for human players.
		select {
		case <-ctx.Done():
			return
		default:
			// move to switch statement if player is alive, otherwise close the stream.
			for _, p := range ih.allPlayers {
				if p.Index != pIdx || p.IsAlive() {
					continue
				}
				return
			}
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
		switch ih.phase {
		case game.SelectAction:
			time.Sleep(time.Duration(2000) * time.Millisecond)
			n = ih.rng.IntN(7) + 1
			outChan <- rune(n + '0')
		case game.SelectTarget:
			// No need for the bots to cancel their target selections, so need
			// to add one to the final result.
			if len(ih.validTargets) == 0 {
				retryCounter ++
				time.Sleep(time.Duration(100) * time.Millisecond)
				continue
			}
			if retryCounter >= 10 {
				panic("Retried valid target check 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(len(ih.validTargets)) + 1
			outChan <- rune(n + '0')
		case game.MakeChallenge, game.ChallengeBlock:
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
			roll := ih.rng.IntN(100)
			if roll < challengeRate {
				outChan <- rune(1 + '0')
				continue
			}
			outChan <- rune(0 + '0')
		case game.MakeBlock:
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
			hand := ih.activePlayers[0].CardsHeld
			if ih.action == game.Assassinate && slices.Contains(hand, game.Contessa) {
				outChan <- rune(1 + '0')
				continue
			}
			if ih.action == game.ForeignAid && slices.Contains(hand, game.Duke) {
				outChan <- rune(1 + '0')
				continue
			}
			if ih.action == game.Steal && slices.Contains(hand, game.Ambassador) {
				outChan <- rune(1 + '0')
				continue
			}
			if ih.action == game.Steal && slices.Contains(hand, game.Captain) {
				outChan <- rune(2 + '0')
				continue
			}
			roll := ih.rng.IntN(100) // Roll a d100 if no card to block
			if roll >= challengeRate {
				outChan <- rune(0 + '0')
				continue
			}
			if ih.action == game.Steal {
				n = ih.rng.IntN(2) + 1
				outChan <- rune(n + '0')
				continue
			}
			outChan <- rune(1 + '0')
		case game.ChallengeReveal:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			hand := ih.activePlayers[0].CardsHeld
			if len(hand) == 0 {
				retryCounter++
				continue
			}
			if retryCounter >= 10 {
				panic("retried ChallengeReveal 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(len(hand)) + 1
			// Replace n if they have the correct card
			if ih.action == game.Assassinate && slices.Contains(hand, game.Assassin) {
				n = slices.Index(hand, game.Assassin) + 1
			}
			if ih.action == game.Exchange && slices.Contains(hand, game.Ambassador) {
				n = slices.Index(hand, game.Ambassador) + 1
			}
			if ih.action == game.Steal && slices.Contains(hand, game.Captain) {
				n = slices.Index(hand, game.Captain) + 1
			}
			if ih.action == game.Tax && slices.Contains(hand, game.Duke) {
				n = slices.Index(hand, game.Duke) + 1
			}
			outChan <- rune(n + '0')
		case game.BlockReveal:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			hand := ih.activePlayers[0].CardsHeld
			if len(hand) == 0 {
				retryCounter++
				continue
			}
			if retryCounter >= 10 {
				panic("retried BlockReveal 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(len(hand)) + 1
			// Replace n if they have the correct card
			if ih.action == game.Assassinate && slices.Contains(hand, game.Contessa) {
				n = slices.Index(hand, game.Contessa)
			}
			if ih.action == game.ForeignAid && slices.Contains(hand, game.Duke) {
				n = slices.Index(hand, game.Duke)
			}
			if ih.blockType == game.Ambassador && slices.Contains(hand, game.Ambassador) {
				n = slices.Index(hand, game.Ambassador)
			}
			if ih.blockType == game.Captain && slices.Contains(hand, game.Captain) {
				n = slices.Index(hand, game.Captain)
			}
			outChan <- rune(n + '0')
		case game.ChallengeLoss, game.BlockLoss:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			hand := ih.activePlayers[0].CardsHeld
			if len(hand) == 0 {
				retryCounter++
				continue
			}
			if retryCounter >= 10 {
				panic("retried Challenge/BlockLoss 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(len(hand)) + 1
			outChan <- rune(n + '0')
		case game.ResolveAction:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			if ih.action == game.Assassinate || ih.action == game.Coup {
				if len(ih.activePlayers) == 0 {
					retryCounter++
					continue
				}
				hand := ih.activePlayers[0].CardsHeld
				if len(hand) == 0 {
					retryCounter++
					continue
				}
				if retryCounter >= 10 {
					panic("retried ResolveAction Assassinate 10+ times")
				}
				retryCounter = 0
				n = ih.rng.IntN(len(hand)) + 1
				outChan <- rune(n + '0')
			}
			// If not Assassinate, Coup or Exchange, no input is required, so
			// just continue to get to ExchangeMiddle
		case game.ExchangeMiddle:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			handLength := len(ih.activePlayers[0].CardsHeld)
			if handLength == 0 {
				retryCounter++
				continue
			}
			if retryCounter >= 10 {
				panic("retried ExchangeMiddle 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(handLength) + 1
			outChan <- rune(n + '0')
		case game.ExchangeFinal:
			time.Sleep(time.Duration(1500) * time.Millisecond)
			// Bots don't need to cancel, so just return a number randomly.
			handLength := len(ih.activePlayers[0].CardsHeld)
			if handLength == 0 {
				retryCounter++
				continue
			}
			if retryCounter >= 10 {
				panic("retried ExchangeFinal 10+ times")
			}
			retryCounter = 0
			n = ih.rng.IntN(handLength) + 1
			outChan <- rune(n + '0')
		case game.EndGame:
			return
		default:
			panic("Unreachable code! (ih.generateBotInput)")
		}
	}
}

// receiveLocalInputs is the method used to get input from a local player.
// This is where the code to quit the game on 'q' exists and handles errors
// that may arise from reading stdin.
func (ih *InputHandler) CreateHumanInputStream(ctx context.Context, outChan chan<- rune) {
	defer ih.RecoverPanic("Panic captured by CreateHumanInputStream")
	var buf = make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			// ih.chanErr <- fmt.Errorf("Error reading from stdin: %w", err)
			panic(err)
		}
		if rune(buf[0]) == 'q' {
			ih.chanErr <- fmt.Errorf("User Quit")
		}
		select {
		case <-ctx.Done():
			return
		case outChan <- rune(buf[0]):
			continue
		}
	}
}

// Scratch

func (ih *InputHandler) RecoverPanic(msg string) {
	if captured := recover(); captured != nil {
		err := fmt.Errorf("%s: %v", msg, captured)
		ih.chanErr <- err
	}
}
