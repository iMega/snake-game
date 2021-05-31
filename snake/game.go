package snake

import (
	"time"

	"github.com/imega/snake-game/state"
	"github.com/nsf/termbox-go"
)

var (
	pointsChan         = make(chan int)
	KeyboardEventsChan = make(chan KeyboardEvent)
)

// Game type
type Game struct {
	arena  *arena
	score  int
	isOver bool
}

func initialSnake() *snake {
	return newSnake(RIGHT, []coord{
		coord{x: 1, y: 1},
		coord{x: 1, y: 2},
		coord{x: 1, y: 3},
		coord{x: 1, y: 4},
	})
}

func initialScore() int {
	return 0
}

func initialArena() *arena {
	return newArena(initialSnake(), pointsChan, 20, 50)
}

func (g *Game) end() {
	g.isOver = true
}

func (g *Game) moveInterval() time.Duration {
	ms := 1000 - (g.score / 10)
	return time.Duration(ms) * time.Millisecond
}

func (g *Game) retry() {
	g.arena = initialArena()
	g.score = initialScore()
	g.isOver = false
}

func (g *Game) addPoints(p int) {
	g.score += p
}

// NewGame creates new Game object
func NewGame() *Game {
	return &Game{arena: initialArena(), score: initialScore()}
}

// Start starts the game
func (g *Game) Start(ch chan state.SnakeGame) {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	go listenToKeyboard(KeyboardEventsChan)

	if err := g.render(); err != nil {
		panic(err)
	}

mainloop:
	for {
		select {
		case p := <-pointsChan:
			g.addPoints(p)
		case e := <-KeyboardEventsChan:
			switch e.EventType {
			case MOVE:
				d := keyToDirection(e.Key)
				g.arena.snake.changeDirection(d)
			case RETRY:
				g.retry()
			case END:
				break mainloop
			}
		default:
			if !g.isOver {
				if err := g.arena.moveSnake(); err != nil {
					g.end()
				}
			}

			var body []state.Coord
			for _, v := range g.arena.snake.body {
				body = append(body, state.Coord{
					X: v.x,
					Y: v.y,
				})
			}

			ch <- state.SnakeGame{
				Score:  g.score,
				IsOver: g.isOver,
				Arena: state.Arena{
					Width:  g.arena.width,
					Height: g.arena.height,
				},
				Food: state.Coord{
					X: g.arena.food.x,
					Y: g.arena.food.y,
				},
				Snake: state.Snake{
					Head: state.Coord{
						X: g.arena.snake.head().x,
						Y: g.arena.snake.head().y,
					},
					Body: body,
				},
			}

			if err := g.render(); err != nil {
				panic(err)
			}

			time.Sleep(g.moveInterval())
		}
	}
}
