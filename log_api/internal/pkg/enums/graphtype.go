package enums

const (
	topTimeSpentGraphType     = "top-time-spent"
	topCountriesGraphType     = "top-country"
	playersInfoGraphType      = "players-info"
	onlineStatisticsGraphType = "online-statistics"
)

//nolint:gochecknoglobals // enum can ignore it
var GraphTypes graphTypes

type GraphType string

func (gt GraphType) IsValid() bool {
	switch gt {
	case topTimeSpentGraphType, topCountriesGraphType, playersInfoGraphType, onlineStatisticsGraphType:
		return true
	default:
		return false
	}
}

func (gt GraphType) String() string {
	return string(gt)
}

type graphTypes struct{}

func (graphTypes) TopTimeSpentGraphType() GraphType     { return topTimeSpentGraphType }
func (graphTypes) TopCountriesGraphType() GraphType     { return topCountriesGraphType }
func (graphTypes) PlayersInfoGraphType() GraphType      { return playersInfoGraphType }
func (graphTypes) OnlineStatisticsGraphType() GraphType { return onlineStatisticsGraphType }
