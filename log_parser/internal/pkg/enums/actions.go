package enums

const (
	enteredAction          = "entered"
	connectedAction        = "connected"
	disconnectedAction     = "disconnected"
	committedSuicideAction = "committed suicide"
)

//nolint:gochecknoglobals // needs var to be public to use the enum
var Actions actions

type Action string

func (a Action) IsValid() bool {
	switch a {
	case enteredAction, connectedAction, disconnectedAction, committedSuicideAction:
		return true
	default:
		return false
	}
}

func (a Action) String() string {
	return string(a)
}

type actions struct{}

func (actions) Entered() Action                { return enteredAction }
func (actions) Connected() Action              { return connectedAction }
func (actions) Disconnected() Action           { return disconnectedAction }
func (actions) CommittedSuicideAction() Action { return committedSuicideAction }
