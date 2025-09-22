package display

import(
	"fmt"
)

const HelpText = `
Overview
========
Coup is a game of intrigue and espionage played by up to six people. On your
turn you will select one of seven actions to improve your position, intefere
with your enemies, and emerge victorious as the only remaining actor.

But be careful; your enemies are ever watchful, and will wield their own
influence to challenge you at every turn.

Setup
=====
The deck is prepared with 15 cards, 3 each of the 5 influential characters
that the game is centered around. Each player will receive 2 coins and 2 of
these cards face down. These cards are known as "influence" and represent
a player's connections and allies. The remaining cards form a communal deck.

Gameplay
========
On each turn a player may select any one of the 7 following game actions.
If a player has 10 or more coins they must select Coup.

The first three are universal and do not require any influence claims.
1. Income - gain 1 coin
2. Foreign Aid - gain 2 coins
3. Coup - costs 7 coins - target player loses influence

The remaining four all require the player to claim a particular influence.
4. Assassinate (Assassin) - costs 3 coins - target player loses influence
5. Exchange (Ambassador) - draw 2 cards, then return 2 cards to the deck
6. Steal (Captain) - steal up to 2 coins from target player
7. Tax (Duke) - gain 3 coins

Three of the above actions may be blocked by claiming an influence as below.
* Foreign Aid - blocked by Duke
* Assassinate - blocked by Contessa
* Steal - blocked by Ambassador or Captain

Challenges
==========
After any influence claim any other player may challenge that claim. The
claimant must then reveal a card. If it does not match the claimed influence it
is lost. If it does match, it is shuffled back into the deck and a new card is
drawn in its place. The challenger must then lose one influence of their choice.

Victory and Elimination
=======================
When a player loses all of their influence, they are eliminated from the game.
When only one player remains with any influence, they are crowned the victor.
`

func highlightSelected(str string) string {
	return fmt.Sprintf("\033[7m%s\033[0m", str)
}

func highlight[T any](value T) string {
	return fmt.Sprintf("\033[7m%v\033[0m", value)
}

func (d *Display) DrawMainMenu() {
	d.buildString(d.row, 3, "select number of players by number keys")
	d.row += 2
	d.buildString(d.row, 12, "# Players: ")
	for i := range 4 {
		if i == d.Selection {
			d.buildString(d.row, 23 + i*2, fmt.Sprintf("%s", highlight(i+3)))
			continue
		}
		d.buildString(d.row, 23 + i*2, fmt.Sprintf("%d", i+3))
	}
	d.row += 4
	d.buildString(d.row, 12, "press Enter to begin")
	d.row += 2
	d.buildString(d.row, 8, "press 'q' at any time to quit")
}

