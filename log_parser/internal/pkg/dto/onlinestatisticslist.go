package dto

import "time"

type OnlineStatisticsList []OnlineStatistics

type OnlineStatistics []OnlineStatisticsHourUnit

type OnlineStatisticsHourUnit struct {
	Hour                  int `json:"hour"`
	ConcurentPlayersCount int `json:"concurent_players_count"`
}

type Session struct {
	NickName string
	Start    time.Time
	End      time.Time
}
