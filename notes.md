## Bugs
- Automatic player turns?
- Quit not working.
- Card loss mechanics are bugged.
- Exchange doesn't give all return options.
- Display not updating after blocking Foreign Aid?
- Double printing of Block notifications?

## States
- SelectAction: NoAction    **1**
- SelectTarget: Assassinate, Coup, Steal *3*: reduces to **1**
- MakeChallenge: Assassinate, Exchange, Steal, Tax *4*: reduces to **1**
- ChallengeReveal: Assassinate, Exchange, Steal, Tax    *4* **1**
- ChallengeLoss: Assassinate, Exchange, Steal, Tax  *4* **1**
- MakeBlock: ForeignAid, Assassinate, Steal             *3* **4**
- ChallengeBlock: ForeignAid, Assassinate, Steal(AMB), Steal(CAP) *4* **1**
- ChallengeBlockReveal: ForeignAid, Assassinate, Steal(AMB), Steal(CAP) *4* **1**
- ChallengeBlockLoss: ForeignAid, Assassinate, Steal        *3* **1**
- ResolveAction: Income, ForeignAid, Coup, Assassinate, Exchange, Steal, Tax *7* **4/6**

- Total States = 1 + 3 + 4 + 4 + 4 + 3 + 4 + 4 + 3 + 7
               = 1 + 3x3 + 5x4 + 7
               = **37 Total States** reduces to **16** or **18** update functions

- 10 "true" states, each with average ~4 action subtypes

## Flow
- set raw mode and defer restoration
- initialise screen, state and i/o channels
### Loop
- active players send input
- input verified and sent to controller
- user command is used to update state
- state sends specific serialised data to each display
- displays unmarshall serialised data and render to screen

## Actor Model architecture!
