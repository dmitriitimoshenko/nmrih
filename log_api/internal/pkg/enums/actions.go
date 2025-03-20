package enums

const (
	// enteredAction          = "entered"
	connectedAction    = "connected"
	disconnectedAction = "disconnected"
	// committedSuicideAction = "committed suicide"
)

//nolint:gochecknoglobals // enum can ignore it
var Actions actions

type Action string

func (a Action) IsValid() bool {
	switch a {
	case connectedAction, disconnectedAction: //, enteredAction, committedSuicideAction:
		return true
	default:
		return false
	}
}

func (a Action) String() string {
	return string(a)
}

type actions struct{}

// func (actions) Entered() Action          { return enteredAction }
func (actions) Connected() Action    { return connectedAction }
func (actions) Disconnected() Action { return disconnectedAction }

// func (actions) CommittedSuicide() Action { return committedSuicideAction }
