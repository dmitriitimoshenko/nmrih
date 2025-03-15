package dto

import "time"

type OnlineStatistics []OnlineStatisticsHourUnit

type OnlineStatisticsHourUnit struct {
	Hour                  int `json:"hour"`
	ConcurrentPlayersCount int `json:"concurrent_players_count"`
}

type Session struct {
	NickName string
	Start    time.Time
	End      time.Time
}
