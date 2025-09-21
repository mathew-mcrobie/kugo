package game

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"slices"
)

// Treat this value as a constant. Go does not allow arrays to be constant but
// This should be considered one regardless.
var BOT_NAMES = [5]string{"Alice", "Bob", "Charlie", "Diana", "Elsie"}

var(
	logFile, _ = os.Create("debug.log")
	debug = log.New(logFile, "[DEBUG]", log.Lshortfile)
)

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

type State struct {
	Phase  Phase
	Action Action
}

type stateHandler func(*Controller, int, int) State

var handlers = map[State]stateHandler{
	State{Phase: SelectAction, Action: NoAction}:       (*Controller).selectAction,
	State{Phase: SelectTarget, Action: Coup}:           (*Controller).selectTarget,
	State{Phase: SelectTarget, Action: Assassinate}:    (*Controller).selectTarget,
	State{Phase: SelectTarget, Action: Steal}:          (*Controller).selectTarget,
	State{Phase: MakeChallenge, Action: Assassinate}:   (*Controller).makeChallenge,
	State{Phase: MakeChallenge, Action: Exchange}:      (*Controller).makeChallenge,
	State{Phase: MakeChallenge, Action: Steal}:         (*Controller).makeChallenge,
	State{Phase: MakeChallenge, Action: Tax}:           (*Controller).makeChallenge,
	State{Phase: ChallengeReveal, Action: Assassinate}: (*Controller).challengeReveal,
	State{Phase: ChallengeReveal, Action: Exchange}:    (*Controller).challengeReveal,
	State{Phase: ChallengeReveal, Action: Steal}:       (*Controller).challengeReveal,
	State{Phase: ChallengeReveal, Action: Tax}:         (*Controller).challengeReveal,
	State{Phase: ChallengeLoss, Action: Assassinate}:   (*Controller).challengeLoss,
	State{Phase: ChallengeLoss, Action: Exchange}:      (*Controller).challengeLoss,
	State{Phase: ChallengeLoss, Action: Steal}:         (*Controller).challengeLoss,
	State{Phase: ChallengeLoss, Action: Tax}:           (*Controller).challengeLoss,
	State{Phase: MakeBlock, Action: ForeignAid}:        (*Controller).makeBlock,
	State{Phase: MakeBlock, Action: Assassinate}:       (*Controller).makeBlock,
	State{Phase: MakeBlock, Action: Steal}:             (*Controller).makeBlock,
	State{Phase: ChallengeBlock, Action: ForeignAid}:   (*Controller).challengeBlock,
	State{Phase: ChallengeBlock, Action: Assassinate}:  (*Controller).challengeBlock,
	State{Phase: ChallengeBlock, Action: Steal}:        (*Controller).challengeBlock,
	State{Phase: BlockReveal, Action: ForeignAid}:      (*Controller).blockReveal,
	State{Phase: BlockReveal, Action: Assassinate}:     (*Controller).blockReveal,
	State{Phase: BlockReveal, Action: Steal}:           (*Controller).blockReveal,
	State{Phase: BlockLoss, Action: ForeignAid}:        (*Controller).blockLoss,
	State{Phase: BlockLoss, Action: Assassinate}:       (*Controller).blockLoss,
	State{Phase: BlockLoss, Action: Steal}:             (*Controller).blockLoss,
	State{Phase: ResolveAction, Action: Income}:        (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: ForeignAid}:    (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: Coup}:          (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: Assassinate}:   (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: Exchange}:      (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: Steal}:         (*Controller).resolveAction,
	State{Phase: ResolveAction, Action: Tax}:           (*Controller).resolveAction,
	State{Phase: ExchangeMiddle, Action: Exchange}:     (*Controller).exchangeMiddle,
	State{Phase: ExchangeFinal, Action: Exchange}:      (*Controller).exchangeFinal,
	State{Phase: EndGame, Action: NoAction}:			(*Controller).endGame,
	State{Phase: MainMenu, Action: NoAction}:			(*Controller).mainMenu,
}

type Controller struct {
	State
	rng           *rand.Rand
	deck          []Card
	actionLog     *ActionLog
	TotalPlayers  int
	AllPlayers    []*Player
	activePlayers []*Player
	current       *Player
	target        *Player
	blocker       *Player
	challenger    *Player
	passed        int
	blockType     Card
	returnedCards []Card
	selection     int
	playerIndex   int
	exchangeDrawn bool
}

func NewController(players []*Player) *Controller {
	var newDeck []Card
	rngOut := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	actionLog := NewActionLog(10)
	stateIn := State{Phase: SelectAction, Action: NoAction}
	allCards := [5]Card{Ambassador, Assassin, Captain, Contessa, Duke}
	for _, c := range allCards {
		for range 3 {
			newDeck = append(newDeck, c)
		}
	}
	cOut := Controller{
		State:         stateIn,
		actionLog:     actionLog,
		AllPlayers:    players,
		activePlayers: []*Player{players[0]},
		rng:           rngOut,
		deck:          newDeck,
		current:       players[0],
	}
	return &cOut
}

func (c *Controller) ShuffleAndDeal() {
	c.shuffle()
	c.deal()
}

func (c *Controller) shuffle() {
	c.rng.Shuffle(len(c.deck), func(i, j int) {
		c.deck[i], c.deck[j] = c.deck[j], c.deck[i]
	})
}

func (c *Controller) deal() {
	for _, p := range c.AllPlayers {
		n := c.rng.IntN(len(c.deck) - 1)
		p.CardsHeld = append(p.CardsHeld, c.deck[n])
		p.CardsHeld = append(p.CardsHeld, c.deck[n+1])
		c.deck = slices.Delete(c.deck, n, n+2)
	}
}

func (c *Controller) getValidTargets() []*Player {
	var validTargets []*Player
	for _, p := range c.AllPlayers {
		if !p.IsAlive() || p.Index == c.current.Index {
			continue
		}
		validTargets = append(validTargets, p)
	}
	return validTargets
}

func (c *Controller) GetDisplayData() *DisplayData {
	var validTargets []*Player
	if c.Phase == SelectTarget {
		validTargets = c.getValidTargets()
	}
	data := c.NewDisplayData(validTargets)
	return data
}
func (c *Controller) GetStateData() *StateData {
	var validTargets []*Player
	if c.Phase == SelectTarget {
		validTargets = c.getValidTargets()
	}
	data := NewStateData(c, validTargets)
	return data
}

func (c *Controller) excludeOneLivingPlayer(p *Player) []*Player {
	var playersOut []*Player
	for _, player := range c.AllPlayers {
		if player == p || !player.IsAlive() {
			continue
		}
		playersOut = append(playersOut, player)
	}
	return playersOut
}

func (c *Controller) setActivePlayers() {
	c.activePlayers = nil
	switch c.Phase {
	case SelectAction, SelectTarget, ChallengeReveal, ExchangeMiddle, ExchangeFinal:
		c.activePlayers = []*Player{c.current}
	case MakeChallenge:
		c.activePlayers = c.excludeOneLivingPlayer(c.current)
	case ChallengeLoss, BlockLoss:
		c.activePlayers = []*Player{c.challenger}
	case MakeBlock:
		if c.Action == ForeignAid {
			c.activePlayers = c.excludeOneLivingPlayer(c.current)
			return
		}
		c.activePlayers = []*Player{c.target}
	case ChallengeBlock:
		c.activePlayers = c.excludeOneLivingPlayer(c.blocker)
	case BlockReveal:
		c.activePlayers = []*Player{c.blocker}
	case ResolveAction:
		if c.Action == Coup || c.Action == Assassinate {
			c.activePlayers = []*Player{c.target}
			return
		}
		if c.Action == Exchange {
			c.activePlayers = []*Player{c.current}
			return
		}
	case MainMenu, EndGame:
		for _, p := range c.AllPlayers {
			if !p.IsLocal {
				continue
			}
			c.activePlayers = append(c.activePlayers, p)
		}
	}
}

func (c *Controller) swapCard(playerIdx, cardIdx int) {
	player := c.AllPlayers[playerIdx]
	card := player.CardsHeld[cardIdx]
	player.CardsHeld = slices.Delete(player.CardsHeld, cardIdx, cardIdx+1)
	c.deck = append(c.deck, card)
	drawIdx := c.rng.IntN(len(c.deck))
	player.CardsHeld = append(player.CardsHeld, c.deck[drawIdx])
	c.deck = slices.Delete(c.deck, drawIdx, drawIdx+1)
}

func (c *Controller) loseCard(playerIdx, cardIdx int) {
	player := c.AllPlayers[playerIdx]
	card := player.CardsHeld[cardIdx]
	player.CardsHeld = slices.Delete(player.CardsHeld, cardIdx, cardIdx+1)
	player.CardsLost = append(player.CardsLost, card)
}

func (c *Controller) getCurrentCard() Card {
	switch c.Action {
	case Assassinate:
		return Assassin
	case Exchange:
		return Ambassador
	case Steal:
		return Captain
	case Tax:
		return Duke
	default:
		return NoCard
	}
}

func (c *Controller) UpdateGame(data *InputData) {
	if c.target != nil && !c.target.IsAlive() {
		if c.Action == Steal {
			c.current.Coins += 2
		}
		c.State = c.advanceTurn()
		c.setActivePlayers()
		return
	}
	gameContinues, _ := c.checkForGameEnd()
	if !gameContinues {
		c.State = State{EndGame, NoAction}
		c.setActivePlayers()
		return
	}
	c.selection = data.Selection
	c.playerIndex = data.PlayerIndex

	debug.Printf("state - %v; active - %v; input - %v", c.State, c.activePlayers, *data)
	handler := handlers[c.State]
	newState := handler(c, c.selection, c.playerIndex)
	c.State = newState
	c.setActivePlayers()
}

func (c *Controller) checkForGameEnd() (bool, int) {
	l := len(c.AllPlayers)
	// Loops through all players exactly once. Checks for a living players, and
	// on finding one makes them the current player. A full loop finding no living
	// players means the current player is the only one left and has won.
	for i := (c.current.Index + 1) % l; i != c.current.Index; i = (i + 1) % l {
		if !c.AllPlayers[i].IsAlive() {
			continue
		}
		return true, i
	}
	return false, c.current.Index
}

func (c *Controller) advanceTurn() State {
	gameContinues, pIdx := c.checkForGameEnd()
	if !gameContinues {
		return State{Phase: EndGame, Action: NoAction}
	}

	// Now we know that the game will be continuing, we can update for the next turn.
	c.current = c.AllPlayers[pIdx]
	c.activePlayers = nil
	c.target = nil
	c.blocker = nil
	c.challenger = nil
	c.returnedCards = nil
	c.exchangeDrawn = false
	c.blockType = 0
	c.passed = 0
	c.selection = 0
	c.playerIndex = 0
	return State{Phase: SelectAction, Action: NoAction}
}

func (c *Controller) mainMenu(sel, pIdx int) State {
	if sel != 200 {
		c.TotalPlayers = sel
		return c.State
	}
	return State{SelectAction, NoAction}
}

func (c *Controller) selectAction(sel, pIdx int) State {
	nextAction := Action(sel)
	if nextAction != Assassinate && nextAction != Coup && nextAction != Steal {
		toLog := fmt.Sprintf("%s has selected %s", c.current, nextAction)
		c.actionLog.Enqueue(toLog)
	}
	switch nextAction {
	case Assassinate, Coup, Steal:
		return State{Phase: SelectTarget, Action: nextAction}
	case ForeignAid:
		return State{Phase: MakeBlock, Action: nextAction}
	case Exchange, Tax:
		return State{Phase: MakeChallenge, Action: nextAction}
	default:
		return State{Phase: ResolveAction, Action: nextAction}
	}
}

func (c *Controller) selectTarget(sel, pIdx int) State {
	if sel == 0 {
		return State{Phase: SelectAction, Action: NoAction}
	}
	// InputHandler added 1 to target index before sending as sel to
	// disambiguate targeting validTargets[0] and cancelling. So must subtract
	// 1 now to correct.
	validTargets := c.getValidTargets()
	c.target = validTargets[sel-1]
	if c.Action == Steal {
		toLog := fmt.Sprintf("%s is attempting to %s from %s", c.current, c.Action, c.target)
		c.actionLog.Enqueue(toLog)
		return State{Phase: MakeChallenge, Action: c.Action}
	}
	if c.Action == Assassinate {
		c.current.Coins -= 3
		toLog := fmt.Sprintf("%s has spent 3 coins to attempt to %s %s", c.current, c.Action, c.target)
		c.actionLog.Enqueue(toLog)
		return State{Phase: MakeChallenge, Action: c.Action}
	}
	// Coup is the only other targeted action, and it can't be challenged/blocked.
	c.current.Coins -= 7
	toLog := fmt.Sprintf("%s has spent 7 coins to launch a %s against %s!", c.current, c.Action, c.target)
	c.actionLog.Enqueue(toLog)
	return State{Phase: ResolveAction, Action: c.Action}
}

func (c *Controller) makeChallenge(sel, pIdx int) State {
	if sel == 0 {
		currentCard := c.getCurrentCard()
		c.actionLog.Enqueue(fmt.Sprintf("No one dares challenge %s's %s claim", c.current, currentCard))
		switch c.Action {
		case Assassinate, Steal:
			return State{Phase: MakeBlock, Action: c.Action}
		default:
			return State{Phase: ResolveAction, Action: c.Action}
		}
	}
	c.challenger = c.AllPlayers[pIdx]
	toLog := fmt.Sprintf(
		"%s is challenging the %s claim of %s",
		c.challenger,
		c.Action.Card(),
		c.current,
	)
	c.actionLog.Enqueue(toLog)
	return State{Phase: ChallengeReveal, Action: c.Action}
}

func (c *Controller) challengeReveal(sel, pIdx int) State {
	revealedCard := c.current.CardsHeld[sel]
	var toLogFail = fmt.Sprintf(
		"Challenge fails! %s shuffles %s into the deck and draws a new card",
		c.current,
		revealedCard,
	)
	var toLogSucceed = fmt.Sprintf(
		"Challenge succeeds! %s loses %s",
		c.current,
		revealedCard,
	)
	c.actionLog.Enqueue(fmt.Sprintf("%s reveals... %s!", c.current, revealedCard))
	// Use the switch statement to check if the challenge fails.
	// If it does, swap the current player's card and get the
	// challenger to lose a card. Else, lose the revealed card
	// and advance the turn.
	switch c.Action {
	case Assassinate:
		if revealedCard == Assassin {
			c.actionLog.Enqueue(toLogFail)
			c.swapCard(c.current.Index, sel)
			c.actionLog.Enqueue(fmt.Sprintf("%s must lose influence", c.challenger))
			return State{Phase: ChallengeLoss, Action: c.Action}
		}
		c.actionLog.Enqueue(toLogSucceed)
		c.loseCard(c.current.Index, sel)
		return c.advanceTurn()
	case Exchange:
		if revealedCard == Ambassador {
			c.actionLog.Enqueue(toLogFail)
			c.swapCard(c.current.Index, sel)
			c.actionLog.Enqueue(fmt.Sprintf("%s must lose influence", c.challenger))
			return State{Phase: ChallengeLoss, Action: c.Action}
		}
		c.actionLog.Enqueue(toLogSucceed)
		c.loseCard(c.current.Index, sel)
		return c.advanceTurn()
	case Steal:
		if revealedCard == Captain {
			c.actionLog.Enqueue(toLogFail)
			c.swapCard(c.current.Index, sel)
			c.actionLog.Enqueue(fmt.Sprintf("%s must lose influence", c.challenger))
			return State{Phase: ChallengeLoss, Action: c.Action}
		}
		c.actionLog.Enqueue(toLogSucceed)
		c.loseCard(c.current.Index, sel)
		return c.advanceTurn()
	case Tax:
		if revealedCard == Duke {
			c.actionLog.Enqueue(toLogFail)
			c.swapCard(c.current.Index, sel)
			c.actionLog.Enqueue(fmt.Sprintf("%s must lose influence", c.challenger))
			return State{Phase: ChallengeLoss, Action: c.Action}
		}
		c.actionLog.Enqueue(toLogSucceed)
		c.loseCard(c.current.Index, sel)
		return c.advanceTurn()
	default:
		panic("Unreachable code! (*Controller.challengeReveal)")
	}
}

func (c *Controller) challengeLoss(sel, pIdx int) State {
	lostCard := c.challenger.CardsHeld[sel]
	c.loseCard(pIdx, sel)
	c.actionLog.Enqueue(fmt.Sprintf("%s chooses to lose %s", c.challenger, lostCard))
	// Assassinate and Steal can still be blocked after the initial challenge.
	if (c.Action == Assassinate || c.Action == Steal) && c.target.IsAlive() {
		return State{Phase: MakeBlock, Action: c.Action}
	}
	// Exchange and Tax can't be blocked, so head straight to resolution.
	return State{Phase: ResolveAction, Action: c.Action}
}

func (c *Controller) makeBlock(sel, pIdx int) State {
	if sel == 0 {
		return State{Phase: ResolveAction, Action: c.Action}
	}
	if sel == 1 && c.Action == Steal {
		c.blocker = c.AllPlayers[pIdx]
		c.blockType = Ambassador
		toLog := fmt.Sprintf("%s is claiming %s to block %s's %s", c.blocker, c.blockType, c.current, c.Action)
		c.actionLog.Enqueue(toLog)
		return State{Phase: ChallengeBlock, Action: c.Action}
	}
	if sel == 2 && c.Action == Steal {
		c.blocker = c.AllPlayers[pIdx]
		c.blockType = Captain
		toLog := fmt.Sprintf("%s is claiming %s to block %s's %s", c.blocker, c.blockType, c.current, c.Action)
		c.actionLog.Enqueue(toLog)
		return State{Phase: ChallengeBlock, Action: c.Action}
	}
	if sel == 1 {
		c.blocker = c.AllPlayers[pIdx]
		if c.Action == Assassinate {
			c.blockType = Contessa
		}
		if c.Action == ForeignAid {
			c.blockType = Duke
		}
		toLog := fmt.Sprintf("%s is claiming %s to block %s's %s", c.blocker, c.blockType, c.current, c.Action)
		c.actionLog.Enqueue(toLog)
		return State{Phase: ChallengeBlock, Action: c.Action}
	}
	panic("Unreachable code! (makeBlock)")
}

func (c *Controller) challengeBlock(sel, pIdx int) State {
	// An unchallenged block ends the turn.
	if sel == 0 {
		c.actionLog.Enqueue(fmt.Sprintf("%s successfully blocks %s's %s attempt!", c.blocker, c.current, c.Action))
		return c.advanceTurn()
	}

	c.challenger = c.AllPlayers[pIdx]
	toLog := fmt.Sprintf(
		"%s is challenging the %s claim of %s",
		c.challenger,
		c.blockType,
		c.blocker,
	)
	c.actionLog.Enqueue(toLog)
	return State{Phase: BlockReveal, Action: c.Action}
}

func (c *Controller) blockReveal(sel, pIdx int) State {
	revealedCard := c.blocker.CardsHeld[sel]
	c.actionLog.Enqueue(fmt.Sprintf("%s reveals... %s!", c.blocker, revealedCard))
	// Same as challengeReveal, except a failed challenge always leads to
	// action resolution, simplifying significantly.
	if revealedCard == c.blockType {
		c.actionLog.Enqueue(fmt.Sprintf(
			"Challenge fails. %s returns %s to the deck and draws a new card",
			c.blocker,
			revealedCard,
		))
		c.swapCard(c.blocker.Index, sel)
		c.actionLog.Enqueue(fmt.Sprintf("%s must lose influence", c.challenger))
		return State{Phase: BlockLoss, Action: c.Action}
	}
	c.loseCard(pIdx, sel)
	toLog := fmt.Sprintf("Challenge succeeds. %s loses %s and %s's %s continues!",
		c.blocker,
		revealedCard,
		c.current,
		c.Action,
	)
	c.actionLog.Enqueue(toLog)
	return State{Phase: ResolveAction, Action: c.Action}
}

func (c *Controller) blockLoss(sel, pIdx int) State {
	cardLost := c.challenger.CardsHeld[sel]
	// Blocker has succeeded in blocking by surviving the challenge.
	// Turn will end.
	c.actionLog.Enqueue(fmt.Sprintf("%s chooses to lose %s", c.challenger, cardLost))
	c.loseCard(pIdx, sel)
	return c.advanceTurn()
}

func (c *Controller) resolveAction(sel, pIdx int) State {
	switch c.Action {
	case Income:
		c.current.Coins += 1
		c.actionLog.Enqueue(fmt.Sprintf("%s gains 1 coin", c.current))
	case ForeignAid:
		c.current.Coins += 2
		c.actionLog.Enqueue(fmt.Sprintf("%s gains 2 coins", c.current))
	case Coup:
		c.loseCard(pIdx, sel)
		totalLost := len(c.target.CardsLost)
		lost := c.target.CardsLost[totalLost-1]
		c.actionLog.Enqueue(fmt.Sprintf("%s loses %s", c.target, lost))
	case Assassinate:
		if c.target.IsAlive() {
			c.loseCard(pIdx, sel)
			totalLost := len(c.target.CardsLost)
			lost := c.target.CardsLost[totalLost-1]
			c.actionLog.Enqueue(fmt.Sprintf("%s loses %s", c.target, lost))
		}
	case Exchange:
		// Need to draw the cards before getting input, so just draw them and
		// move to the next phase for card selection
		c.exchangeDrawTwo()
		c.actionLog.Enqueue(fmt.Sprintf("%s draws 2 cards", c.current))
		return State{ExchangeMiddle, c.Action}
	case Steal:
		stolen := min(2, c.target.Coins)
		c.target.Coins -= stolen
		c.current.Coins += stolen
		c.actionLog.Enqueue(fmt.Sprintf("%s steals %d coins from %s", c.current, stolen, c.target))
	case Tax:
		c.current.Coins += 3
		c.actionLog.Enqueue(fmt.Sprintf("%s gains 3 coins", c.current))
	}
	return c.advanceTurn()
}

func (c *Controller) exchangeMiddle(sel, pIdx int) State {
	// Hold on to the returned card for now so we can undo if the player
	// cancels their choice in the next phase.
	c.returnedCards = append(c.returnedCards, c.current.CardsHeld[sel])
	c.current.CardsHeld = slices.Delete(c.current.CardsHeld, sel, sel+1)
	return State{Phase: ExchangeFinal, Action: c.Action}
}

func (c *Controller) exchangeFinal(sel, pIdx int) State {
	if sel == 0 {
		// Player changed their mind, so go back to the previous step.
		c.current.CardsHeld = append(c.current.CardsHeld, c.returnedCards...)
		c.returnedCards = nil
		return State{Phase: ExchangeMiddle, Action: c.Action}
	}
	// Must subtract 1 from sel to get correct index as 0 means cancel.
	c.returnedCards = append(c.returnedCards, c.current.CardsHeld[sel-1])
	c.current.CardsHeld = slices.Delete(c.current.CardsHeld, sel-1, sel)
	// Now we can put the returned cards back in the deck.
	c.deck = append(c.deck, c.returnedCards...)
	c.actionLog.Enqueue(fmt.Sprintf("%s returns 2 chosen cards to the deck", c.current))
	return c.advanceTurn()
}

func (c *Controller) endGame(sel, pIdx int) State {
	c.selection, c.playerIndex = sel, pIdx
	return c.State
}

func (c *Controller) exchangeDrawTwo() {
	for range 2 {
		n := c.rng.IntN(len(c.deck))
		c.current.CardsHeld = append(c.current.CardsHeld, c.deck[n])
		c.deck = slices.Delete(c.deck, n, n+1)
	}
}
