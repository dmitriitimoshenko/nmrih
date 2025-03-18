package graph

import (
	"math"
	"sort"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/rumblefrog/go-a2s"
)

const (
	topPlayersCount             = 32
	topCountries                = 9
	minSessionDurationInMinutes = 10

	secondsInHour = 3600.0
	hoursInDay    = 24
	maxCentsCount = 100
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
		for nickName := range lastConnected {
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
		percentage := float64(topCountry.ConnectionsCount) / float64(allConnectionsCount) * maxCentsCount
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

func (s *Service) OnlineStatistics(logsInput []*dto.LogData) dto.OnlineStatistics {
	sort.Slice(logsInput, func(i, j int) bool {
		return logsInput[i].TimeStamp.Before(logsInput[j].TimeStamp)
	})

	requestTimeStamp := time.Now()
	earliestLogEntry := requestTimeStamp
	var logs []*dto.LogData
	for _, logEntry := range logsInput {
		if logEntry.Action != enums.Actions.Connected() && logEntry.Action != enums.Actions.Disconnected() {
			continue
		}
		logs = append(logs, logEntry)
		if logEntry.TimeStamp.Before(earliestLogEntry) {
			earliestLogEntry = logEntry.TimeStamp
		}
	}

	var sessions []dto.Session
	activeConnections := make(map[string]time.Time)
	for _, logEntry := range logs {
		switch logEntry.Action {
		case enums.Actions.Connected():
			if _, exists := activeConnections[logEntry.NickName]; !exists {
				activeConnections[logEntry.NickName] = logEntry.TimeStamp
				continue
			}
			lastActivityTimeStamp := s.findLastUserActivityTimeStampBefore(
				logs,
				logEntry.TimeStamp,
				logEntry.NickName,
			)
			if lastActivityTimeStamp == nil {
				continue
			}
			sessions = append(sessions, dto.Session{
				NickName: logEntry.NickName,
				Start:    activeConnections[logEntry.NickName],
				End:      *lastActivityTimeStamp,
			})
			delete(activeConnections, logEntry.NickName)
		case enums.Actions.Disconnected():
			if start, exists := activeConnections[logEntry.NickName]; exists {
				sessions = append(sessions, dto.Session{
					NickName: logEntry.NickName,
					Start:    start,
					End:      logEntry.TimeStamp,
				})
				delete(activeConnections, logEntry.NickName)
			}
		}
	}
	if len(activeConnections) > 0 {
		for nickName := range activeConnections {
			lastActivityTimeStamp := s.findLastUserActivityTimeStamp(logs, nickName)
			if lastActivityTimeStamp == nil {
				continue
			}
			sessions = append(sessions, dto.Session{
				NickName: nickName,
				Start:    activeConnections[nickName],
				End:      *lastActivityTimeStamp,
			})
			delete(activeConnections, nickName)
		}
	}

	minSessionDuration := minSessionDurationInMinutes * time.Minute
	validSessions := make([]dto.Session, 0, len(sessions))
	for _, sess := range sessions {
		if sess.End.Sub(sess.Start) >= minSessionDuration {
			validSessions = append(validSessions, sess)
		}
	}
	sessions = validSessions

	timelineStart := time.Date(
		earliestLogEntry.Year(),
		earliestLogEntry.Month(),
		earliestLogEntry.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	timelineEnd := time.Date(
		requestTimeStamp.Year(),
		requestTimeStamp.Month(),
		requestTimeStamp.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	dayCount := 0
	for d := timelineStart; d.Before(timelineEnd); d = d.Add(hoursInDay * time.Hour) {
		dayCount++
	}
	if dayCount == 0 {
		dayCount = 1
	}

	hourlyOverlap := make([]float64, hoursInDay)

	for _, session := range sessions {
		effectiveStart := session.Start
		if effectiveStart.Before(timelineStart) {
			effectiveStart = timelineStart
		}
		effectiveEnd := session.End
		if effectiveEnd.After(timelineEnd.Add(hoursInDay * time.Hour)) {
			effectiveEnd = timelineEnd.Add(hoursInDay * time.Hour)
		}
		if !effectiveEnd.After(effectiveStart) {
			continue
		}

		startIndex := int(effectiveStart.Sub(timelineStart).Hours())
		endIndex := int(math.Ceil(effectiveEnd.Sub(timelineStart).Hours()))
		for i := startIndex; i < endIndex; i++ {
			blockStart := timelineStart.Add(time.Duration(i) * time.Hour)
			blockEnd := blockStart.Add(time.Hour)
			overlapStart := effectiveStart
			if blockStart.After(overlapStart) {
				overlapStart = blockStart
			}
			overlapEnd := effectiveEnd
			if blockEnd.Before(overlapEnd) {
				overlapEnd = blockEnd
			}
			overlap := overlapEnd.Sub(overlapStart).Seconds()
			if overlap > 0 {
				bucket := blockStart.Hour()
				hourlyOverlap[bucket] += overlap
			}
		}
	}

	avgHourlyStats := make(dto.OnlineStatistics, 0, hoursInDay)
	for hour, totalOverlap := range hourlyOverlap {
		avg := totalOverlap / (float64(dayCount) * secondsInHour)
		avgHourlyStats = append(avgHourlyStats, dto.OnlineStatisticsHourUnit{
			Hour:                   hour,
			ConcurrentPlayersCount: avg,
		})
	}

	return append(avgHourlyStats[3:], avgHourlyStats[:4]...)
}
