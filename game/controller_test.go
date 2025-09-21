package game

import (
	"fmt"
	"slices"
	"testing"
)

func assertEqual[T comparable](t *testing.T, got, want T, desc string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", desc, got, want)
	}
}

func assertNotNil(t *testing.T, got interface{}, desc string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: %v is nil", desc, got)
	}
}

func assertError(t *testing.T, desc string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error in %s: %v", desc, err)
	}
}

func setupTestController() *Controller {
	var IsHuman, IsLocal bool
	var players []*Player
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Elsie"}
	for i := range 5 {
		if i == 0 {
			IsLocal = true
		}
		if i < 2 {
			IsHuman = true
		}
		p, _ := NewPlayer(names[i], i, IsHuman, IsLocal)
		players = append(players, p)
	}
	newCon := NewController(players)
	for j := 3; j > 1; j-- {
		for i, p := range newCon.allPlayers {
			p.CardsHeld = append(p.CardsHeld, newCon.deck[i*j])
		}
		for _, k := range [5]int{4, 3, 2, 1, 0} {
			newCon.deck = slices.Delete(newCon.deck, k*j, k*j+1)
		}
	}
	return newCon
}

func runControllerUpdateTest(t *testing.T, initial State, sel, pIdx int, final State) {
	testCon := setupTestController()
	testCon.State = initial
	inputData := NewInputData(sel, pIdx)
	testCon.UpdateGame(inputData)
	if testCon.State != final {
		t.Errorf("got %v, want %v", testCon.State, final)
	}
}

func TestGetStateData(t *testing.T) {
	testCon := setupTestController()
	testCon.State = State{SelectTarget, Steal}
	testCon.setActivePlayers()
	data := testCon.GetStateData()
	if data.ActivePlayers[0].Name != "Alice" {
		t.Errorf("Got %s; want %s", data.ActivePlayers[0], "Alice")
	}
	if data.ActivePlayers[0].Coins != 2 {
		t.Errorf("Got %d; want %d", data.ActivePlayers[0].Coins, 2)
	}
	if !data.ActivePlayers[0].IsHuman {
		t.Errorf("Got %t; want %t", data.ActivePlayers[0].IsHuman, true)
	}
	if !data.ActivePlayers[0].IsLocal {
		t.Errorf("Got %t; want %t", data.ActivePlayers[0].IsLocal, true)
	}
	assertEqual[int](t, len(data.ValidTargets), 4, "test valid targets")
}

func TestPlayerInputData(t *testing.T) {
	var testInputData = []struct {
		sIn, pIn, sOut, pOut int
	}{
		{1, 0, 1, 0},
		{2, 1, 2, 1},
		{3, 2, 3, 2},
		{4, 3, 4, 3},
		{5, 4, 5, 4},
		{6, 0, 6, 0},
		{7, 1, 7, 1},
	}
	for _, tt := range testInputData {
		testName := fmt.Sprintf("sig: %d, pIdx: %d", tt.sIn, tt.pIn)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			playerInput := NewInputData(tt.sIn, tt.pIn)
			testCon.UpdateGame(playerInput)
			assertEqual[int](t, testCon.selection, tt.sOut, testName)
			assertEqual[int](t, testCon.playerIndex, tt.pOut, testName)
		})
	}
}

func TestShuffleAndDeal(t *testing.T) {
	var IsHuman, IsLocal bool
	var players []*Player
	var firstDeck []Card
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Elsie"}
	for i := range 5 {
		if i == 0 {
			IsLocal = true
		}
		if i < 2 {
			IsHuman = true
		}
		p, _ := NewPlayer(names[i], i, IsHuman, IsLocal)
		players = append(players, p)
	}
	newCon := NewController(players)
	for _, c := range newCon.deck {
		firstDeck = append(firstDeck, c)
	}
	newCon.ShuffleAndDeal()
	if slices.IsSorted(newCon.deck) {
		t.Errorf("%v should be shuffled.", newCon.deck)
	}
	assertEqual[int](t, len(newCon.deck), 5, "test deck length")
	for i, p := range newCon.allPlayers {
		assertEqual[int](t, len(p.CardsHeld), 2, fmt.Sprintf("test hands %d", i))
	}
}

func TestUpdateGameSelectAction(t *testing.T) {
	var testData = []struct {
		sel, pIdx int
		stateOut  State
	}{
		{1, 0, State{ResolveAction, Income}},
		{2, 1, State{MakeBlock, ForeignAid}},
		{3, 2, State{SelectTarget, Coup}},
		{4, 3, State{SelectTarget, Assassinate}},
		{5, 4, State{MakeChallenge, Exchange}},
		{6, 0, State{SelectTarget, Steal}},
		{7, 1, State{MakeChallenge, Tax}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s", tt.stateOut.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.selection = tt.sel
			testCon.playerIndex = tt.pIdx
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			if testCon.State != tt.stateOut {
				t.Errorf("%s: got %v; want %v", testName, testCon.State, tt.stateOut)
			}
		})
	}
}

func TestUpdateSelectTarget(t *testing.T) {
	var testData = []struct {
		action    Action
		sel, pIdx int
		nameOut   string
		stateOut  State
	}{
		{Coup, 2, 0, "Bob", State{ResolveAction, Coup}},
		{Assassinate, 3, 0, "Charlie", State{MakeChallenge, Assassinate}},
		{Steal, 4, 0, "Diana", State{MakeChallenge, Steal}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %v target", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = State{SelectTarget, tt.action}
			testCon.allPlayers[0].Coins += 6
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertNotNil(t, testCon.target, testName)
			assertEqual[string](t, testCon.target.Name, tt.nameOut, testName)
			assertEqual[State](t, testCon.State, tt.stateOut, testName)
			assertEqual[string](t, testCon.current.Name, "Alice", testName)
			if tt.stateOut.Action == Coup {
				assertEqual[int](t, testCon.current.Coins, 1, testName)
			}
			if tt.stateOut.Action == Assassinate {
				assertEqual[int](t, testCon.current.Coins, 5, testName)
			}
		})
	}
}

func TestSelectTargetCancel(t *testing.T) {
	testCon := setupTestController()
	testCon.State = State{SelectTarget, Assassinate}
	inputData := NewInputData(0, 0)
	testCon.allPlayers[0].Coins += 6
	testCon.UpdateGame(inputData)
	if testCon.target != nil {
		t.Errorf("got %s; want nil", testCon.target)
	}
	assertEqual[State](t, testCon.State, State{SelectAction, NoAction}, "TargetCancel:")
	assertEqual[int](t, testCon.current.Coins, 8, "TargetCancel coins:")
}

func TestUpdateChallengeReveal(t *testing.T) {
	var testData = []struct {
		state                 State
		sel, pIdx, challenger int
		want                  State
	}{
		{State{ChallengeReveal, Assassinate}, 0, 1, 0, State{ChallengeLoss, Assassinate}},
		{State{ChallengeReveal, Exchange}, 0, 0, 1, State{ChallengeLoss, Exchange}},
		{State{ChallengeReveal, Steal}, 1, 2, 4, State{ChallengeLoss, Steal}},
		{State{ChallengeReveal, Tax}, 1, 4, 3, State{ChallengeLoss, Tax}},
		{State{ChallengeReveal, Assassinate}, 1, 3, 4, State{SelectAction, NoAction}},
		{State{ChallengeReveal, Exchange}, 1, 2, 3, State{SelectAction, NoAction}},
		{State{ChallengeReveal, Steal}, 0, 4, 0, State{SelectAction, NoAction}},
		{State{ChallengeReveal, Tax}, 0, 0, 1, State{SelectAction, NoAction}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test challenge %s", tt.state.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = tt.state
			testCon.current = testCon.allPlayers[tt.pIdx]
			testCon.challenger = testCon.allPlayers[tt.challenger]
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.want, testName)
			initState := State{SelectAction, NoAction}
			if testCon.State == initState {
				assertEqual[int](t, testCon.current.Index, (tt.pIdx+1)%5, testName)
			} else {
				assertEqual[int](t, testCon.current.Index, tt.pIdx, testName)
			}
		})
	}
}

func TestUpdateChallengeLoss(t *testing.T) {
	var testData = []struct {
		state     State
		sel, pIdx int
		want      State
	}{
		{State{ChallengeLoss, Assassinate}, 0, 0, State{MakeBlock, Assassinate}},
		{State{ChallengeLoss, Exchange}, 0, 1, State{ResolveAction, Exchange}},
		{State{ChallengeLoss, Steal}, 1, 4, State{MakeBlock, Steal}},
		{State{ChallengeLoss, Tax}, 1, 3, State{ResolveAction, Tax}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("Test challenge loss (%s):", tt.state.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = tt.state
			testCon.challenger = testCon.allPlayers[tt.pIdx]
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.want, testName)
			assertEqual[int](t, len(testCon.challenger.CardsHeld), 1, testName)
		})
	}
}

func TestUpdateBlockReveal(t *testing.T) {
	var testData = []struct {
		state                 State
		sel, pIdx, challenger int
		blockType             Card
		want                  State
	}{
		{State{BlockReveal, Assassinate}, 0, 1, 0, Assassin, State{BlockLoss, Assassinate}},
		{State{BlockReveal, ForeignAid}, 0, 4, 1, Duke, State{BlockLoss, ForeignAid}},
		{State{BlockReveal, Steal}, 1, 0, 4, Ambassador, State{BlockLoss, Steal}},
		{State{BlockReveal, Steal}, 1, 2, 4, Captain, State{BlockLoss, Steal}},
		{State{BlockReveal, Assassinate}, 0, 3, 4, Assassin, State{ResolveAction, Assassinate}},
		{State{BlockReveal, ForeignAid}, 1, 2, 3, Duke, State{ResolveAction, ForeignAid}},
		{State{BlockReveal, Steal}, 0, 4, 0, Ambassador, State{ResolveAction, Steal}},
		{State{BlockReveal, Steal}, 0, 3, 0, Captain, State{ResolveAction, Steal}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test block reveal %s", tt.state.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = tt.state
			testCon.blocker = testCon.allPlayers[tt.pIdx]
			testCon.blockType = tt.blockType
			testCon.challenger = testCon.allPlayers[tt.challenger]
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.want, testName)
			if testCon.State.Phase == ResolveAction {
				assertEqual[int](t, len(testCon.blocker.CardsHeld), 1, testName)
			} else {
				assertEqual[int](t, len(testCon.blocker.CardsHeld), 2, testName)
			}
		})
	}
}

func TestUpdateBlockLoss(t *testing.T) {
	var testData = []struct {
		state     State
		sel, pIdx int
		want      State
	}{
		{State{BlockLoss, Assassinate}, 0, 0, State{SelectAction, NoAction}},
		{State{BlockLoss, ForeignAid}, 0, 1, State{SelectAction, NoAction}},
		{State{BlockLoss, Steal}, 1, 4, State{SelectAction, NoAction}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("Test challenge block loss (%s):", tt.state.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = tt.state
			testCon.challenger = testCon.allPlayers[tt.pIdx]
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.want, testName)
			assertEqual[int](t, testCon.current.Index, 1, testName)
		})
	}
}

func TestUpdateResolveAssassinate(t *testing.T) {
	testCon := setupTestController()
	testCon.State = State{ResolveAction, Assassinate}
	testCon.target = testCon.allPlayers[2]
	inputData := NewInputData(1, 2)
	testCon.UpdateGame(inputData)
	prevTarget := testCon.allPlayers[2]
	testState := State{SelectAction, NoAction}
	assertEqual[State](t, testCon.State, testState, "Resolve Assassinate")
	assertEqual[int](t, len(prevTarget.CardsHeld), 1, "Resolve Assassinate")
}

func TestUpdateResolveCoup(t *testing.T) {
	testCon := setupTestController()
	testCon.State = State{ResolveAction, Coup}
	testCon.target = testCon.allPlayers[2]
	inputData := NewInputData(1, 2)
	prevTarget := testCon.allPlayers[2]
	testCon.UpdateGame(inputData)
	testState := State{SelectAction, NoAction}
	assertEqual[State](t, testCon.State, testState, "Resolve Coup")
	assertEqual[int](t, len(prevTarget.CardsHeld), 1, "Resolve Coup")
}

func TestUpdateResolveAction(t *testing.T) {
	var testData = []struct {
		state                                 State
		current, target, wantCoins, wantSteal int
		wantState                             State
	}{
		{State{ResolveAction, Income}, 1, 0, 3, 2, State{SelectAction, NoAction}},
		{State{ResolveAction, ForeignAid}, 2, 0, 4, 2, State{SelectAction, NoAction}},
		{State{ResolveAction, Steal}, 3, 1, 4, 0, State{SelectAction, NoAction}},
		{State{ResolveAction, Tax}, 4, 0, 5, 2, State{SelectAction, NoAction}},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("Test action resolution (%s):", tt.state.Action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = tt.state
			testCon.current = testCon.allPlayers[tt.current]
			testCon.target = testCon.allPlayers[tt.target]
			inputData := NewInputData(0, 0)
			testCon.UpdateGame(inputData)
			prevPlayer := testCon.allPlayers[tt.current]
			prevTarget := testCon.allPlayers[tt.target]
			initState := State{SelectAction, NoAction}
			assertEqual[State](t, testCon.State, initState, testName)
			assertEqual[int](t, prevPlayer.Coins, tt.wantCoins, testName)
			assertEqual[int](t, prevTarget.Coins, tt.wantSteal, testName)
			assertEqual[int](t, testCon.current.Index, (tt.current+1)%5, testName)
		})
	}
}

func TestGameUpdateMakeChallenge(t *testing.T) {
	var testData = []struct {
		action     Action
		sel, pIdx  int
		challenger string
		want       Phase
	}{
		{Assassinate, 1, 3, "Diana", ChallengeReveal},
		{Exchange, 1, 1, "Bob", ChallengeReveal},
		{Steal, 1, 4, "Elsie", ChallengeReveal},
		{Tax, 1, 2, "Charlie", ChallengeReveal},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s happy challenge", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.allPlayers[0].Coins += 6
			testCon.State = State{MakeChallenge, tt.action}
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertNotNil(t, testCon.challenger, testName)
			assertEqual[string](t, testCon.challenger.Name, tt.challenger, testName)
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			if testCon.Action == Assassinate && testCon.current.Coins != 8 {
				t.Errorf("%s: got %d; want %d", testName, testCon.current.Coins, 8)
			}
		})
	}
}

func TestUpdateGameMakeChallengePass(t *testing.T) {
	var testData = []struct {
		action    Action
		sel, pIdx int
		want      Phase
	}{
		{Assassinate, 0, 3, MakeBlock},
		{Exchange, 0, 1, ResolveAction},
		{Steal, 0, 4, MakeBlock},
		{Tax, 0, 2, ResolveAction},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s sad challenge", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.current.Coins += 6
			testCon.State = State{MakeChallenge, tt.action}
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			if testCon.challenger != nil {
				t.Errorf("%s: erroneous challenger found (%s)", testName, testCon.challenger)
			}
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			if testCon.Action == Assassinate && testCon.allPlayers[0].Coins != 8 {
				t.Errorf("%s: got %d; want %d", testName, testCon.current.Coins, 8)
			}
		})
	}
}

func TestGameUpdateMakeBlock(t *testing.T) {
	var testData = []struct {
		action    Action
		sel, pIdx int
		blocker   string
		want      Phase
		wantBlock Card
	}{
		{Assassinate, 1, 3, "Diana", ChallengeBlock, Contessa},
		{ForeignAid, 1, 1, "Bob", ChallengeBlock, Duke},
		{Steal, 1, 4, "Elsie", ChallengeBlock, Ambassador},
		{Steal, 2, 2, "Charlie", ChallengeBlock, Captain},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s happy block", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = State{MakeBlock, tt.action}
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertNotNil(t, testCon.blocker, testName)
			assertEqual[string](t, testCon.blocker.Name, tt.blocker, testName)
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			assertEqual[Card](t, testCon.blockType, tt.wantBlock, testName)
		})
	}
}

func TestUpdateGameMakeBlockPass(t *testing.T) {
	var testData = []struct {
		action    Action
		sel, pIdx int
		want      Phase
		wantBlock Card
	}{
		{Assassinate, 0, 3, ResolveAction, NoCard},
		{ForeignAid, 0, 1, ResolveAction, NoCard},
		{Steal, 0, 4, ResolveAction, NoCard},
		{Steal, 0, 2, ResolveAction, NoCard},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s pass block", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = State{MakeBlock, tt.action}
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			if testCon.blocker != nil {
				t.Errorf("%s: erroneous blocker found (%s)", testName, testCon.blocker)
			}
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			assertEqual[Card](t, testCon.blockType, tt.wantBlock, testName)
		})
	}
}

func TestGameUpdateChallengeBlock(t *testing.T) {
	var testData = []struct {
		action     Action
		sel, pIdx  int
		challenger string
		want       Phase
		wantBlock  Card
	}{
		{Assassinate, 1, 3, "Diana", BlockReveal, NoCard},
		{ForeignAid, 1, 1, "Bob", BlockReveal, NoCard},
		{Steal, 1, 4, "Elsie", BlockReveal, Ambassador},
		{Steal, 1, 2, "Charlie", BlockReveal, Captain},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s block challenge", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = State{ChallengeBlock, tt.action}
			testCon.blockType = tt.wantBlock
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertNotNil(t, testCon.challenger, testName)
			assertEqual[string](t, testCon.challenger.Name, tt.challenger, testName)
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			assertEqual[Card](t, testCon.blockType, tt.wantBlock, testName)
		})
	}
}

func TestUpdateGameChallengeBlockPass(t *testing.T) {
	var testData = []struct {
		action    Action
		sel, pIdx int
		want      Phase
		blockType Card
	}{
		{Assassinate, 0, 3, SelectAction, NoCard},
		{ForeignAid, 0, 1, SelectAction, NoCard},
		{Steal, 0, 4, SelectAction, Ambassador},
		{Steal, 0, 2, SelectAction, Captain},
	}
	for _, tt := range testData {
		testName := fmt.Sprintf("test %s pass block challenge", tt.action)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.State = State{ChallengeBlock, tt.action}
			testCon.blockType = tt.blockType
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			if testCon.challenger != nil {
				t.Errorf("%s: erroneous challenger found (%s)", testName, testCon.challenger)
			}
			assertEqual[Phase](t, testCon.Phase, tt.want, testName)
			assertEqual[Card](t, testCon.blockType, NoCard, testName)
		})
	}
}

func TestUpdateGameExchangeResolve(t *testing.T) {
	var testData = []struct {
		state                State
		sel, pIdx, wantCards int
		wantState            State
	}{
		{State{ResolveAction, Exchange}, 3, 0, 3, State{ExchangeFinal, Exchange}},
		{State{ResolveAction, Exchange}, 2, 1, 2, State{ExchangeFinal, Exchange}},
		{State{ResolveAction, Exchange}, 1, 2, 3, State{ExchangeFinal, Exchange}},
		{State{ResolveAction, Exchange}, 0, 3, 2, State{ExchangeFinal, Exchange}},
	}
	for i, tt := range testData {
		testName := fmt.Sprintf("Test Exchange resolution %d", i)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.current = testCon.allPlayers[tt.pIdx]
			testCon.allPlayers[0].CardsHeld = append(testCon.allPlayers[0].CardsHeld, testCon.deck[:2]...)
			testCon.deck = testCon.deck[2:]
			testCon.allPlayers[1].CardsHeld = append(testCon.allPlayers[1].CardsHeld, testCon.deck[0])
			testCon.deck = testCon.deck[1:]
			testCon.loseCard(3, 1)
			testCon.State = tt.state
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.wantState, testName)
			assertEqual[int](t, len(testCon.current.CardsHeld), tt.wantCards, testName)
		})
	}
}

func TestUpdateGameExchangeFinalCommit(t *testing.T) {
	var testData = []struct {
		state                State
		sel, pIdx, wantCards int
		wantState            State
	}{
		{State{ExchangeFinal, Exchange}, 2, 0, 2, State{SelectAction, NoAction}},
		{State{ExchangeFinal, Exchange}, 1, 1, 1, State{SelectAction, NoAction}},
	}
	for i, tt := range testData {
		testName := fmt.Sprintf("Test Exchange final commit %d", i)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.current = testCon.allPlayers[tt.pIdx]
			testCon.allPlayers[0].CardsHeld = append(testCon.allPlayers[0].CardsHeld, testCon.deck[0])
			testCon.returnedCards = append(testCon.returnedCards, testCon.deck[1])
			testCon.deck = testCon.deck[2:]
			testCon.State = tt.state
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.wantState, testName)
			assertEqual[int](t, len(testCon.allPlayers[tt.pIdx].CardsHeld), tt.wantCards, testName)
			assertEqual[int](t, len(testCon.returnedCards), 0, testName)
		})
	}
}

func TestUpdateExchangeFinalCancel(t *testing.T) {
	var testData = []struct {
		state                State
		sel, pIdx, wantCards int
		wantState            State
	}{
		{State{ExchangeFinal, Exchange}, 0, 2, 4, State{ResolveAction, Exchange}},
		{State{ExchangeFinal, Exchange}, 0, 3, 3, State{ResolveAction, Exchange}},
	}
	for i, tt := range testData {
		testName := fmt.Sprintf("Test Exchange final cancel %d", i)
		t.Run(testName, func(t *testing.T) {
			testCon := setupTestController()
			testCon.current = testCon.allPlayers[tt.pIdx]
			testCon.returnedCards = append(testCon.returnedCards, testCon.deck[0])
			testCon.allPlayers[2].CardsHeld = append(testCon.allPlayers[2].CardsHeld, testCon.deck[1])
			testCon.deck = testCon.deck[2:]
			testCon.State = tt.state
			inputData := NewInputData(tt.sel, tt.pIdx)
			testCon.UpdateGame(inputData)
			assertEqual[State](t, testCon.State, tt.wantState, testName)
			assertEqual[int](t, len(testCon.current.CardsHeld), tt.wantCards, testName)
			assertEqual[int](t, len(testCon.returnedCards), 0, testName)
		})
	}
}
