package game

import (
	"fmt"
)

type Player struct {
	Name      string
	Index     int
	Coins     int
	CardsHeld []Card
	CardsLost []Card
	IsHuman   bool
	IsLocal   bool
	Responded bool
}

func NewPlayer(name string, index int, isHuman, isLocal bool) (*Player, error) {
	if !isHuman && isLocal {
		return nil, fmt.Errorf("AI controlled player cannot be local")
	}
	playerOut := Player{
		Name:    name,
		Index:   index,
		Coins:   2,
		IsHuman: isHuman,
		IsLocal: isLocal,
	}
	return &playerOut, nil
}

func (p *Player) IsAlive() bool {
	return len(p.CardsLost) != 2
}

func (p *Player) String() string {
	return p.Name
}
