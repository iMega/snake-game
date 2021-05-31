package snake

import "github.com/nsf/termbox-go"

type keyboardEventType int

// Keyboard events
const (
	MOVE keyboardEventType = 1 + iota
	RETRY
	END
)

type KeyboardEvent struct {
	EventType keyboardEventType
	Key       termbox.Key
}

func keyToDirection(k termbox.Key) direction {
	switch k {
	case termbox.KeyArrowLeft:
		return LEFT
	case termbox.KeyArrowDown:
		return DOWN
	case termbox.KeyArrowRight:
		return RIGHT
	case termbox.KeyArrowUp:
		return UP
	default:
		return 0
	}
}

func listenToKeyboard(evChan chan KeyboardEvent) {
	termbox.SetInputMode(termbox.InputEsc)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowLeft:
				evChan <- KeyboardEvent{EventType: MOVE, Key: ev.Key}
			case termbox.KeyArrowDown:
				evChan <- KeyboardEvent{EventType: MOVE, Key: ev.Key}
			case termbox.KeyArrowRight:
				evChan <- KeyboardEvent{EventType: MOVE, Key: ev.Key}
			case termbox.KeyArrowUp:
				evChan <- KeyboardEvent{EventType: MOVE, Key: ev.Key}
			case termbox.KeyEsc:
				evChan <- KeyboardEvent{EventType: END, Key: ev.Key}
			default:
				if ev.Ch == 'r' {
					evChan <- KeyboardEvent{EventType: RETRY, Key: ev.Key}
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
