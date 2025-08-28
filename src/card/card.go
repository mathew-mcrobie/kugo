package card

// Card enum with underlying int
type Card int

const (
	Ambassador Card = iota
	Assassin
	Captain
	Contessa
	Duke
)

// CardName is a map to associate Card names to string values.
var CardName = map[Card]string{
	Ambassador: "Ambassador",
	Assassin:   "Assassin",
	Captain:    "Captain",
	Contessa:   "Contessa",
	Duke:       "Duke",
}

// Implement the String function to satisfy the Stringer interface.
func (c Card) String() string {
	return CardName[c]
}
