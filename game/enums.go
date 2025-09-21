package game

import "strings"

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
	Income:      "Income",
	ForeignAid:  "Foreign Aid",
	Coup:        "Coup",
	Assassinate: "Assassinate",
	Exchange:    "Exchange",
	Steal:       "Steal",
	Tax:         "Tax",
}

var actionColor = map[Action]string{
	Income:      "",
	ForeignAid:  "",
	Coup:        "",
	Assassinate: "\033[37m",
	Exchange:    "\033[32m",
	Steal:       "\033[36m",
	Tax:         "\033[35m",
}

var actionCard = map[Action]Card{
	Assassinate: Assassin,
	Exchange:    Ambassador,
	Steal:       Captain,
	Tax:         Duke,
}

func (a Action) String() string {
	return actionColor[a] + actionName[a] + "\033[0m"
}

func (a Action) Card() Card {
	return actionCard[a]
}

type Phase int

const (
	SelectAction Phase = iota
	SelectTarget
	MakeChallenge
	ChallengeReveal
	ChallengeLoss
	MakeBlock
	ChallengeBlock
	BlockReveal
	BlockLoss
	ResolveAction
	ExchangeMiddle
	ExchangeFinal
	EndGame
)

var phaseName = map[Phase]string{
	SelectAction: 		"SelectAction",
	SelectTarget: 		"SelectTarget",
	MakeChallenge: 		"MakeChallenge",
	ChallengeReveal: 	"ChallengeReveal",
	ChallengeLoss: 		"ChallengeLoss",
	MakeBlock:			"MakeBlock",
	ChallengeBlock: 	"ChallengeBlock",
	BlockReveal:		"BlockReveal",
	BlockLoss:			"BlockLoss",
	ResolveAction:		"ResolveAction",
	ExchangeMiddle:		"ExchangeMiddle",
	ExchangeFinal:		"ExchangeFinal",
	EndGame:			"EndGame",
}

func (p Phase) String() string {
	return phaseName[p]
}

type Card int

const (
	NoCard Card = iota
	Ambassador
	Assassin
	Captain
	Contessa
	Duke
)

var cardName = map[Card]string{
	Ambassador: "Ambassador",
	Assassin:   "Assassin",
	Captain:    "Captain",
	Contessa:   "Contessa",
	Duke:       "Duke",
}

var cardColor = map[Card]string{
	Ambassador: "\033[32m",
	Assassin:   "\033[37m",
	Captain:    "\033[36m",
	Contessa:   "\033[31m",
	Duke:       "\033[35m",
}

func (c Card) String() string {
	return cardColor[c] + cardName[c] + "\033[0m"
}

func (c Card) Short() string {
	return cardColor[c] + strings.ToUpper(cardName[c][:3]) + "\033[0m"
}
