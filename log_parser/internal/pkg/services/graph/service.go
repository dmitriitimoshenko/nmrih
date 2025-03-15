package graph

import (
	"sort"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/enums"
	"github.com/rumblefrog/go-a2s"
)

const (
	topPlayersCount = 32
	topCountries    = 9
	breakHourNumber = 4
)

type Service struct {
	a2sClient A2SClient
}

func NewService(a2sClient A2SClient) *Service {
	return &Service{a2sClient: a2sClient}
}

func (s *Service) TopTimeSpent(logs []*dto.LogData) dto.TopTimeSpentList {
	totalSessionsDurations := s.getTotalSessionsDuration(logs)

	topTimeSpentList := make(dto.TopTimeSpentList, 0, len(totalSessionsDurations))
	for nickName, totalSessionsDuration := range totalSessionsDurations {
		topTimeSpentList = append(topTimeSpentList, &dto.TopTimeSpent{
			NickName:  nickName,
			TimeSpent: totalSessionsDuration,
		})
	}

	sort.Slice(topTimeSpentList, func(i, j int) bool {
		return topTimeSpentList[i].TimeSpent > topTimeSpentList[j].TimeSpent
	})

	if len(topTimeSpentList) > topPlayersCount {
		topTimeSpentList = topTimeSpentList[:topPlayersCount]
	}

	return topTimeSpentList
}

func (s *Service) getTotalSessionsDuration(logs []*dto.LogData) map[string]time.Duration {
	totalSessionsDurations := make(map[string]time.Duration)
	lastConnected := make(map[string]time.Time)

	for _, logEntry := range logs {
		switch logEntry.Action {
		case enums.Actions.Connected():
			{
				if _, ok := lastConnected[logEntry.NickName]; ok {
					lastActivityTimeStamp := s.findLastUserActivityTimeStampBefore(
						logs,
						logEntry.TimeStamp,
						logEntry.NickName,
					)
					if lastActivityTimeStamp == nil {
						// Impossible sceanrio, but just in case
						// If there is no activity at all before the current (second in line) connection, skip
						continue
					}
					s.addDurationToTotal(
						logEntry.NickName,
						*lastActivityTimeStamp,
						lastConnected,
						totalSessionsDurations,
					)
				}
				lastConnected[logEntry.NickName] = logEntry.TimeStamp
				break
			}
		case enums.Actions.Disconnected():
			{
				if _, ok := lastConnected[logEntry.NickName]; !ok {
					continue
				}
				s.addDurationToTotal(
					logEntry.NickName,
					logEntry.TimeStamp,
					lastConnected,
					totalSessionsDurations,
				)
				break
			}
		}
	}

	if len(lastConnected) > 0 {
		// If there are still connected users at the end of the logs
		// Add their current session duration to the total
		for nickName, _ := range lastConnected {
			lastActivityTimeStamp := s.findLastUserActivityTimeStamp(
				logs,
				nickName,
			)
			if lastActivityTimeStamp == nil {
				// Impossible sceanrio, but just in case
				// If there is no activity at all before the current (second in line) connection, skip
				continue
			}
			s.addDurationToTotal(
				nickName,
				*lastActivityTimeStamp,
				lastConnected,
				totalSessionsDurations,
			)
		}
	}

	return totalSessionsDurations
}

func (s *Service) addDurationToTotal(
	nickName string,
	lastActivityTimeStamp time.Time,
	lastConnected map[string]time.Time,
	totalSessionsDurations map[string]time.Duration,
) {
	lastSessionDuration := lastActivityTimeStamp.Sub(lastConnected[nickName])
	totalSessionsDurations[nickName] += lastSessionDuration
	delete(lastConnected, nickName)
}

func (s *Service) findLastUserActivityTimeStampBefore(
	logs []*dto.LogData,
	before time.Time,
	nickName string,
) *time.Time {
	var lastActivity time.Time
	for _, entry := range logs {
		if entry.NickName == nickName && entry.TimeStamp.Before(before) {
			if entry.TimeStamp.After(lastActivity) {
				lastActivity = entry.TimeStamp
			}
		}
	}

	if lastActivity.IsZero() {
		return nil
	}
	return &lastActivity
}

func (s *Service) findLastUserActivityTimeStamp(
	logs []*dto.LogData,
	nickName string,
) *time.Time {
	var lastActivity time.Time
	for _, entry := range logs {
		if entry.NickName == nickName && entry.TimeStamp.After(lastActivity) {
			lastActivity = entry.TimeStamp
		}
	}

	if lastActivity.IsZero() {
		return nil
	}
	return &lastActivity
}

func (s *Service) TopCountries(logs []*dto.LogData) dto.TopCountriesPercentageList {
	countriesConnectionsList := make(map[string]int)
	var allConnectionsCount int

	for _, logEntry := range logs {
		if logEntry.Action == enums.Actions.Connected() {
			if logEntry.Country == "" {
				countriesConnectionsList["Unknown"]++
			}
			countriesConnectionsList[logEntry.Country]++
			allConnectionsCount++
		}
	}

	topCountriesList := make(dto.TopCountriesList, 0, topCountries)
	for range topCountries {
		var (
			maxConnectionsCount   int
			maxConnectionsCountry string
		)
		for country, connectionsCount := range countriesConnectionsList {
			if connectionsCount > maxConnectionsCount {
				maxConnectionsCount = connectionsCount
				maxConnectionsCountry = country
			}
		}
		topCountriesList = append(topCountriesList, dto.TopCountry{
			Country:          maxConnectionsCountry,
			ConnectionsCount: maxConnectionsCount,
		})
		delete(countriesConnectionsList, maxConnectionsCountry)
	}

	topCountriesPercentageList := make(dto.TopCountriesPercentageList, 0, len(topCountriesList))
	otherPercentage := 100.0
	for _, topCountry := range topCountriesList {
		percentage := float64(topCountry.ConnectionsCount) / float64(allConnectionsCount) * 100
		otherPercentage -= percentage
		topCountriesPercentageList = append(topCountriesPercentageList, dto.TopCountriesPercentage{
			Country:    topCountry.Country,
			Percentage: percentage,
		})
	}

	topCountriesPercentageList = append(topCountriesPercentageList, dto.TopCountriesPercentage{
		Country:    "Other",
		Percentage: otherPercentage,
	})

	return topCountriesPercentageList
}

func (s *Service) PlayersInfo() (*dto.PlayersInfo, error) {
	playersInfo, err := s.a2sClient.QueryPlayer()
	if err != nil {
		return nil, err
	}

	playersInfoDto := s.mapPlayersInfo(playersInfo)

	return playersInfoDto, nil
}

func (s *Service) mapPlayersInfo(playersInfo *a2s.PlayerInfo) *dto.PlayersInfo {
	playersInfoDto := &dto.PlayersInfo{}

	playersInfoDto.Count = int(playersInfo.Count)

	for _, playerInfo := range playersInfo.Players {
		playersInfoDto.PlayerInfo = append(playersInfoDto.PlayerInfo, &dto.PlayerInfo{
			Name:     playerInfo.Name,
			Score:    playerInfo.Score,
			Duration: playerInfo.Duration,
		})
	}

	return playersInfoDto
}

// Session represents a player's connection session.
type Session struct {
	NickName string
	Start    time.Time
	End      time.Time
}

// daysIn returns the number of days in the given month/year.
func daysIn(month time.Month, year int) int {
	// time.Date with day 0 gives the last day of the previous month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// OnlineStatistics computes, for each complete day and hour, the average concurrent players online.
// It uses the connection/disconnection logs to build sessions, then for each hour bucket it simulates
// connection events to calculate the average count over the hour.
func (s *Service) OnlineStatistics(logs []*dto.LogData) dto.OnlineStatisticsList {
	// Sort logs by timestamp.
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].TimeStamp.Before(logs[j].TimeStamp)
	})

	// Determine the request time and the earliest log entry.
	requestTimeStamp := time.Now()
	earliestLogEntry := requestTimeStamp
	for _, logEntry := range logs {
		if logEntry.TimeStamp.Before(earliestLogEntry) {
			earliestLogEntry = logEntry.TimeStamp
		}
	}

	// Build sessions from logs.
	var sessions []Session
	activeConnections := make(map[string]time.Time)
	for _, logEntry := range logs {
		switch logEntry.Action {
		case enums.Actions.Connected():
			// Start a session if not already connected.
			if _, exists := activeConnections[logEntry.NickName]; !exists {
				activeConnections[logEntry.NickName] = logEntry.TimeStamp
			}
		case enums.Actions.Disconnected():
			// End a session if there is an active connection.
			if start, exists := activeConnections[logEntry.NickName]; exists {
				sessions = append(sessions, Session{
					NickName: logEntry.NickName,
					Start:    start,
					End:      logEntry.TimeStamp,
				})
				delete(activeConnections, logEntry.NickName)
			}
		}
	}
	// For any players still connected, assume they disconnected at requestTimeStamp.
	for nick, start := range activeConnections {
		sessions = append(sessions, Session{
			NickName: nick,
			Start:    start,
			End:      requestTimeStamp,
		})
	}

	// Prepare the final result.
	var stats dto.OnlineStatisticsList

	// We’ll compute statistics for complete days only.
	currentDay := time.Date(earliestLogEntry.Year(), earliestLogEntry.Month(), earliestLogEntry.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(requestTimeStamp.Year(), requestTimeStamp.Month(), requestTimeStamp.Day(), 0, 0, 0, 0, time.UTC)

	// Iterate day by day.
	for day := currentDay; day.Before(endDay); day = day.Add(24 * time.Hour) {
		var dailyStats dto.OnlineStatistics
		// For each hour of the day.
		for hour := 0; hour < 24; hour++ {
			hourStart := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, time.UTC)
			hourEnd := hourStart.Add(time.Hour)

			// For simulation we use events: each event has a timestamp and a delta (+1 or -1).
			type event struct {
				time  time.Time
				delta int
			}
			var events []event

			// For each session that overlaps this hour, determine the effective entry and exit times.
			for _, session := range sessions {
				// Skip sessions that don't overlap this hour.
				if session.End.Before(hourStart) || session.Start.After(hourEnd) {
					continue
				}
				// The effective start is the later of session.Start and hourStart.
				effectiveStart := session.Start
				if effectiveStart.Before(hourStart) {
					effectiveStart = hourStart
				}
				// The effective end is the earlier of session.End and hourEnd.
				effectiveEnd := session.End
				if effectiveEnd.After(hourEnd) {
					effectiveEnd = hourEnd
				}
				events = append(events, event{time: effectiveStart, delta: 1})
				events = append(events, event{time: effectiveEnd, delta: -1})
			}

			// If no sessions overlap this hour, concurrent count is zero.
			if len(events) == 0 {
				dailyStats = append(dailyStats, dto.OnlineStatisticsHourUnit{
					Hour:               hour,
					UniquePlayersCount: 0,
				})
				continue
			}

			// Sort events by time.
			sort.Slice(events, func(i, j int) bool {
				if events[i].time.Equal(events[j].time) {
					// For events at the same time, process +1 before -1.
					return events[i].delta > events[j].delta
				}
				return events[i].time.Before(events[j].time)
			})

			// Simulate the hour to compute a time-weighted average of concurrent players.
			currentCount := 0
			lastTime := hourStart
			var total float64
			for _, ev := range events {
				duration := ev.time.Sub(lastTime).Seconds()
				total += float64(currentCount) * duration
				currentCount += ev.delta
				lastTime = ev.time
			}
			// Account for any remaining time in the hour.
			total += float64(currentCount) * hourEnd.Sub(lastTime).Seconds()

			// Average concurrent players over the hour (3600 seconds).
			avgConcurrency := total / 3600.0
			concurrentPlayers := int(avgConcurrency + 0.5) // round to nearest int

			dailyStats = append(dailyStats, dto.OnlineStatisticsHourUnit{
				Hour:               hour,
				UniquePlayersCount: concurrentPlayers, // now represents concurrent players
			})
		}
		stats = append(stats, dailyStats)
	}

	return stats
}

// func daysIn(month time.Month, year int) int {
// 	// time.Date automatically adjusts when the day is 0,
// 	// so this returns the last day of the given month.
// 	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
// }

// func (s *Service) OnlineStatistics(logs []*dto.LogData) dto.OnlineStatisticsList {
// 	requestTimeStamp := time.Now()
// 	earliestLogEntry := requestTimeStamp
// 	for _, logEntry := range logs {
// 		if logEntry.TimeStamp.Before(earliestLogEntry) {
// 			earliestLogEntry = logEntry.TimeStamp
// 		}
// 	}

// 	for year := earliestLogEntry.Year(); year <= requestTimeStamp.Year(); year++ {
// 		monthUntil := time.December
// 		if year == requestTimeStamp.Year() {
// 			monthUntil = requestTimeStamp.Month()
// 		}
// 		for month := earliestLogEntry.Month(); month <= monthUntil; month++ {
// 			dayUntil := daysIn(month, year)
// 			if year == requestTimeStamp.Year() && month == requestTimeStamp.Month() {
// 				dayUntil = requestTimeStamp.Day()
// 			}
// 			for day := earliestLogEntry.Day(); day <= dayUntil; day++ {
// 				if requestTimeStamp.Year() == year && requestTimeStamp.Month() == month && requestTimeStamp.Day() == day {
// 					continue
// 				}

// 				// implement day by day logic here

// 				// day by day logic finished here
// 			}
// 		}
// 	}
// }

// playersOnlineCountPerHour := make(map[int]int, 0)
// uniqueNickNamesPerHour := make(map[int][]string, 0)
// lastConnected := make(map[string]time.Time)

// for _, logEntry := range logs {
// 	switch logEntry.Action {
// 	case enums.Actions.Connected():
// 		{
// 			lastConnected[logEntry.NickName] = logEntry.TimeStamp
// 			break
// 		}
// 	case enums.Actions.Disconnected():
// 		{
// 			if _, ok := lastConnected[logEntry.NickName]; !ok {
// 				continue
// 			}
// 			if logEntry.TimeStamp.Sub(lastConnected[logEntry.NickName]) > time.Minute*10 {
// 				if lastConnected[logEntry.NickName].Hour() == logEntry.TimeStamp.Hour() {
// 					uniqueNickNamesPerHour[logEntry.TimeStamp.Hour()] = append(
// 						uniqueNickNamesPerHour[logEntry.TimeStamp.Hour()],
// 						logEntry.NickName,
// 					)
// 					playersOnlineCountPerHour[logEntry.TimeStamp.Hour()]++
// 					delete(lastConnected, logEntry.NickName)
// 				} else if lastConnected[logEntry.NickName].Hour() < logEntry.TimeStamp.Hour() {
// 					var hoursOfActivity []int
// 					if lastConnected[logEntry.NickName].Minute() > 50 {
// 						hoursOfActivity = append(hoursOfActivity, lastConnected[logEntry.NickName].Hour())
// 					}

// 					for h := lastConnected[logEntry.NickName].Hour(); h <=
// 				}
// 			}
// 			continue
// 		}
// 	}
// }
