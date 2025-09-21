package game

type InputData struct {
	Selection   int
	PlayerIndex int
}

func NewInputData(sel, pIdx int) *InputData {
	data := InputData{
		Selection:   sel,
		PlayerIndex: pIdx,
	}
	return &data
}

type StateData struct {
	ActivePlayers, ValidTargets []*Player
	BlockType                   Card
	State                       State
}

func NewStateData(active, valid []*Player, block Card, state State) *StateData {
	data := StateData{
		ActivePlayers: active,
		ValidTargets:  valid,
		BlockType:     block,
		State:         state,
	}
	return &data
}

type DisplayData struct {
	AllPlayers    []*Player
	ActivePlayers []*Player
	ValidTargets  []*Player
	Current       *Player
	ActionLog     *ActionLog
	State         State
}

func (c *Controller) NewDisplayData(validTargets []*Player) *DisplayData {
	data := DisplayData{
		AllPlayers:    c.AllPlayers,
		ActivePlayers: c.activePlayers,
		ValidTargets:  validTargets,
		Current:       c.current,
		ActionLog:     c.actionLog,
		State:         c.State,
	}
	return &data
}
