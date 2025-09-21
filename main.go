package main

import (
	"context"
	"os"

	dis "kugo/display"
	"kugo/game"
	inp "kugo/input"
)

func GameLoop() error {
	// Initialize Game
	var players []*game.Player
	var chanErr = make(chan error)
	playerNames := []string{"Matthew", "Alice", "Bob"}
	for i, name := range playerNames {
		var isHuman, isLocal bool
		if i == 0 {
			isHuman, isLocal = true, true
		}
		p, err := game.NewPlayer(name, i, isHuman, isLocal)
		if err != nil {
			panic(err)
		}
		players = append(players, p)
	}
	controller := game.NewController(players)
	controller.ShuffleAndDeal()
	inputHandler := inp.NewInputHandler(players, chanErr)
	display := dis.NewDisplay(chanErr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize input streams
	for i, p := range controller.AllPlayers {
		if p.IsHuman {
			go inputHandler.CreateHumanInputStream(ctx, inputHandler.PlayerChans[i])
			continue
		}
		go inputHandler.CreateBotInputStream(ctx, inputHandler.PlayerChans[i], i)
	}

	// Initialize displays
	dispInit := controller.GetDisplayData()
	display.UpdateDisplay(dispInit)
	go display.DrawDisplay(ctx)

	// Start main game loop
	for {
		// Get input
		stateData := controller.GetStateData()
		inputHandler.UpdateStateData(stateData)
		inputChan := make(chan *game.InputData)
		go func(inputChan chan *game.InputData) {
			playerInput := inputHandler.GetInputData()
			inputChan <- playerInput
		}(inputChan)

		var gotInput bool
		for !gotInput {
			// Update Game
			select {
			case inputData := <-inputChan:
				controller.UpdateGame(inputData)
				gotInput = true
			case err := <-chanErr:
				return err
			default:
				// just update display
			}
			toDisplays := controller.GetDisplayData()
			display.UpdateDisplay(toDisplays)
		}
>>>>>>> dev
	}
}

func main() {
	err := dis.WrapDisplay(GameLoop)
	if err != nil && err.Error() == "User Quit" {
		os.Exit(0)
	}
	if err != nil {
		panic(err)
	}
}

// Scratch
