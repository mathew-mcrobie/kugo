package action

import "fmt"

type ActionState int

const(
	play 			ActionState = iota
	iSelect
	challenge
	block
	blockChallenge
	resolve
)

// ActionInfo is a struct that will be embedded into each action to hold data.
// implements setter and getter for state field.
type ActionInfo struct {
	actingPlayer, targetPlayer Player
	state                      ActionState
}

func (a *ActionInfo) State() { return a.state }
func (a *ActionInfo) SetState(newState ActionState) {a.state = newState }

// Action is an interface for setting state and executing a specific action.
type Action interface {
	Execute() string
}

// Income will represent the income action.
// Implements the Action interface and embeds ActionInfo.
type Income struct {
	ActionInfo
}

// Execute satisfies the Action interface for Income.
// Executing the Income action simply adds one coin to the player's
// coin total and returns a log entry. Because it cannot be interupted,
// there is no need for state changes or conditional code.
func (a *Income) Execute() string {
	a.targetPlayer.AddCoins(1)
	return fmt.Sprintf("{a.actingPlayer} used Income (+1 coin)")
}


// ForeignAid represents the foreign aid action.
// Implements the Action interface and embeds ActionInfo.
type ForeignAid struct {
	ActionInfo
}

func(a *ForeignAid) Execute() string {
	switch a.state {
	case resolve:
		a.targetPlayer.AddCoins(2)
		return fmt.Sprintf("{a.actingPlayer} used Foreign Aid (+2 coins)")
	case blockChallenge:
		return fmt.Sprintf("Declare challenges?")
	case block:
		return fmt.Sprintf("Declare blocks?")
	case play:
		return fmt.Sprintf("{a.actingPlayer} declares Foreign Aid")
	}
}


type Coup struct {
	ActionInfo
}

func (a *Coup) Execute() string {
	switch a.state {
	case resolve:
		// TODO: implement iLoss for coup
	case play:
		a.actingPlayer.LoseCoins(7)
		return fmt.Sprintf("{a.actingPlayer} spends 7 coins to launch a Coup on {a.targetPlayer}!")
	}
}


type Assassinate struct {
	ActionInfo
}


// Exchange represents the exchange action.
// Implements the Action interface and embeds ActionInfo.
// NOTE: Unlike all other Action types, the Execute method on Exchange takes
// an argument!
type Exchange struct {
	ActionInfo
}


type Steal struct {
	ActionInfo
}


type Tax struct {
	ActionInfo
}

func (a *Tax) Execute() string {
	switch a.state {
	case resolve:
		a.targetPlayer.AddCoins(3)
		return fmt.Sprintf("{a.actingPlayer} used Tax (+3 coins)")
	case challenge:
		return fmt.Sprintf("Declare Challenges?")
	case play:
		return fmt.Sprintf("{a.actingPlayer} declares Tax")
	}
}


