package dto

type OnlineStatisticsList []OnlineStatistics

type OnlineStatistics []OnlineStatisticsHourUnit

type OnlineStatisticsHourUnit struct {
	Hour               int `json:"hour"`
	UniquePlayersCount int `json:"unique_players_count"`
}
