package main

import (
	"github.com/imega/snake-game/ai"
	"github.com/imega/snake-game/snake"
	"github.com/imega/snake-game/state"
)

func main() {
	ch := make(chan state.SnakeGame)
	go ai.New(ch, snake.KeyboardEventsChan)
	snake.NewGame().Start(ch)
}
