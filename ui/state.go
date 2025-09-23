package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

//go:generate go tool stringer -type=StateID

// StateID identifies the different states the app can be in
type StateID int

const (
	StateRomList StateID = iota
	StateRomRunning
	StateRomPaused
	StateConfig
	StateDebug
	// Add more states as needed
)

// State represents a UI state with its own update and draw logic
type State interface {
	// Enter is called when transitioning to this state
	Enter(prevState State)

	// Exit is called when transitioning from this state
	Exit(nextState State)

	// Update handles state-specific logic
	Update()

	// Draw renders the state's UI
	Draw(screen *ebiten.Image)

	// ID returns the unique identifier of this state
	ID() StateID
}
