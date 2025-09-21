package main

import (
	"context"
	"fmt"
	"os"
	"time"

	dis "kugo/display"
	"kugo/game"
	inp "kugo/input"
)

func RunMainMenu(chanErr chan error) (int, error) {
	menuChan := make(chan rune)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				chanErr <- fmt.Errorf("menu input panic: %v", r)
			}
		}()
		var buf = make([]byte, 1)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				chanErr <- err
				return
			}
			if rune(buf[0]) == 'q' {
				chanErr <- fmt.Errorf("User Quit")
				return
			}
			select {
			case <-ctx.Done():
				return
			case menuChan <- rune(buf[0]):
			}
		}
	}()

	display := dis.NewDisplay(chanErr)

	var selection int
	var confirmed bool

	go func() {
		for !confirmed {
			display.DrawMenuScreen(selection)
			time.Sleep(time.Millisecond * 41)
		}
	}()

	for !confirmed {
		select {
		case r := <- menuChan:
			switch r {
			case '3', '4', '5', '6':
				selection = int(r - '3') // so it fits 0-3
			case '\r', '\n':
				confirmed = true
			}
		case err := <- chanErr:
			return 0, err
		}
	}
	return selection + 3, nil
}

func GameLoop() error {
	// Run the main menu to get number of players
	var chanErr = make(chan error)
	numPlayers, err := RunMainMenu(chanErr)
	if err != nil {
		return err
	}

	var players []*game.Player
	var playerNames = []string{"Human"}

	// Get player names
	for _, name := range game.BOT_NAMES {
		if len(playerNames) >= numPlayers {
			break
		}
		playerNames = append(playerNames, name)
	}

	// Create players
	for i, name := range playerNames {
		var isHuman, isLocal bool
		if i == 0 {
			isHuman, isLocal = true, true
		}
		p, err := game.NewPlayer(name, i, isHuman, isLocal)
		if err != nil {
			return err
		}
		players = append(players, p)
	}

	// Initialize game loop
	controller := game.NewController(players)
	controller.ShuffleAndDeal()
	inputHandler := inp.NewInputHandler(players, chanErr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize input streams
	for i, _ := range controller.AllPlayers {
		if i == 0 {
			go inputHandler.CreateHumanInputStream(ctx, inputHandler.PlayerChans[i])
			continue
		}
		go inputHandler.CreateBotInputStream(ctx, inputHandler.PlayerChans[i], i)
	}

	// Initialize displays
	display := dis.NewDisplay(chanErr)
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
