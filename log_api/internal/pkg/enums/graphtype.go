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

func (gt GraphType) CanCache() bool {
	return gt != playersInfoGraphType
}

// NeedsLogs reports whether the graph is built from the parsed logs. The live
// player list is not: it comes straight from the game server over A2S, so it
// must keep working before anything has been parsed into CSV.
func (gt GraphType) NeedsLogs() bool {
	return gt != playersInfoGraphType
}

type graphTypes struct{}

func (graphTypes) TopTimeSpentGraphType() GraphType     { return topTimeSpentGraphType }
func (graphTypes) TopCountriesGraphType() GraphType     { return topCountriesGraphType }
func (graphTypes) PlayersInfoGraphType() GraphType      { return playersInfoGraphType }
func (graphTypes) OnlineStatisticsGraphType() GraphType { return onlineStatisticsGraphType }
