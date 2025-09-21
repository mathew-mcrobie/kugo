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
	Target						*Player
	BlockType                   Card
	State                       State
}

func NewStateData(c *Controller, valid []*Player) *StateData {
	data := StateData{
		ActivePlayers: c.activePlayers,
		ValidTargets:  valid,
		Target:		   c.target,
		BlockType:     c.blockType,
		State:         c.State,
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
	Selection	  int
}

func (c *Controller) NewDisplayData(validTargets []*Player) *DisplayData {
	data := DisplayData{
		AllPlayers:    c.AllPlayers,
		ActivePlayers: c.activePlayers,
		ValidTargets:  validTargets,
		Current:       c.current,
		ActionLog:     c.actionLog,
		State:         c.State,
		Selection:	   c.selection,
	}
	return &data
}
