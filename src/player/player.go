package player

// Player holds all information that defines a player's state.
type Player struct {
	name          string
	coins         int
	influence     []Card
	lostInfluence []Card
	isHuman       bool
}
