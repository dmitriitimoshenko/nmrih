package graph

import (
	"math"
	"sort"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools"
	"github.com/rumblefrog/go-a2s"
)

const (
	topPlayersCount             = 32
	topCountriesCount           = 9
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
	totalDurations := s.getTotalSessionsDuration(logs)

	topList := make(dto.TopTimeSpentList, 0, len(totalDurations))
	for nick, duration := range totalDurations {
		topList = append(topList, &dto.TopTimeSpent{
			NickName:  nick,
			TimeSpent: duration,
		})
	}
	sort.Slice(topList, func(i, j int) bool {
		return topList[i].TimeSpent > topList[j].TimeSpent
	})
	if len(topList) > topPlayersCount {
		topList = topList[:topPlayersCount]
	}
	return topList
}

func (s *Service) getTotalSessionsDuration(logs []*dto.LogData) map[string]time.Duration {
	totalDurations := make(map[string]time.Duration)
	lastConnected := make(map[string]time.Time)

	for _, entry := range logs {
		switch entry.Action {
		case enums.Actions.Connected():
			if _, exists := lastConnected[entry.NickName]; exists {
				if lastActivity := s.findLastUserActivityTimeStampBefore(logs, entry.NickName, &entry.TimeStamp); lastActivity != nil {
					s.addDurationToTotal(entry.NickName, *lastActivity, lastConnected, totalDurations)
				}
			}
			lastConnected[entry.NickName] = entry.TimeStamp

		case enums.Actions.Disconnected():
			if _, exists := lastConnected[entry.NickName]; !exists {
				continue
			}
			s.addDurationToTotal(entry.NickName, entry.TimeStamp, lastConnected, totalDurations)
		}
	}

	for nick := range lastConnected {
		if lastActivity := s.findLastUserActivityTimeStamp(logs, nick); lastActivity != nil {
			s.addDurationToTotal(nick, *lastActivity, lastConnected, totalDurations)
		}
	}
	return totalDurations
}

func (s *Service) addDurationToTotal(
	nick string,
	lastActivity time.Time,
	lastConnected map[string]time.Time,
	totalDurations map[string]time.Duration,
) {
	sessionDuration := lastActivity.Sub(lastConnected[nick])
	totalDurations[nick] += sessionDuration
	delete(lastConnected, nick)
}

func (s *Service) findLastUserActivityTimeStampBefore(logs []*dto.LogData, nick string, before *time.Time) *time.Time {
	var lastActivity time.Time
	for _, entry := range logs {
		if entry.NickName == nick &&
			(before == nil || entry.TimeStamp.Before(*before)) &&
			entry.TimeStamp.After(lastActivity) {
			lastActivity = entry.TimeStamp
		}
	}
	if lastActivity.IsZero() {
		return nil
	}
	return &lastActivity
}

func (s *Service) TopCountries(logs []*dto.LogData) dto.TopCountriesPercentageList {
	countryCounts := make(map[string]int)
	var totalConnections int

	for _, entry := range logs {
		if entry.Action != enums.Actions.Connected() {
			continue
		}
		country := entry.Country
		if country == "" {
			country = "Unknown"
		}
		countryCounts[country]++
		totalConnections++
	}

	topCountries := make(dto.TopCountriesList, 0, topCountriesCount)
	for i := 0; i < topCountriesCount && len(countryCounts) > 0; i++ {
		var maxCount int
		var topCountry string
		for country, count := range countryCounts {
			if count > maxCount {
				maxCount = count
				topCountry = country
			}
		}
		topCountries = append(topCountries, dto.TopCountry{
			Country:          topCountry,
			ConnectionsCount: maxCount,
		})
		delete(countryCounts, topCountry)
	}

	var remainingPercentage float64 = 100.0
	topCountriesPercentage := make(dto.TopCountriesPercentageList, 0, len(topCountries)+1)
	for _, tc := range topCountries {
		percentage := float64(tc.ConnectionsCount) / float64(totalConnections) * maxCentsCount
		remainingPercentage -= percentage
		topCountriesPercentage = append(topCountriesPercentage, dto.TopCountriesPercentage{
			Country:    tc.Country,
			Percentage: percentage,
		})
	}
	topCountriesPercentage = append(topCountriesPercentage, dto.TopCountriesPercentage{
		Country:    "Other",
		Percentage: remainingPercentage,
	})

	return topCountriesPercentage
}

func (s *Service) PlayersInfo() (*dto.PlayersInfo, error) {
	playersInfo, err := s.a2sClient.QueryPlayer()
	if err != nil {
		return nil, err
	}
	return s.mapPlayersInfo(playersInfo), nil
}

func (s *Service) mapPlayersInfo(info *a2s.PlayerInfo) *dto.PlayersInfo {
	dtoInfo := &dto.PlayersInfo{Count: int(info.Count)}
	for _, p := range info.Players {
		dtoInfo.PlayerInfo = append(dtoInfo.PlayerInfo, &dto.PlayerInfo{
			Name:     p.Name,
			Score:    p.Score,
			Duration: p.Duration,
		})
	}
	return dtoInfo
}

func (s *Service) OnlineStatistics(logsInput []*dto.LogData) dto.OnlineStatistics {
	var events []*dto.LogData
	sort.Slice(logsInput, func(i, j int) bool {
		return logsInput[i].TimeStamp.Before(logsInput[j].TimeStamp)
	})
	requestTimeStamp := time.Now()
	earliest := requestTimeStamp
	for _, entry := range logsInput {
		if entry.Action != enums.Actions.Connected() && entry.Action != enums.Actions.Disconnected() {
			continue
		}
		events = append(events, entry)
		if entry.TimeStamp.Before(earliest) {
			earliest = entry.TimeStamp
		}
	}

	sessions := s.getSessionsFromLogs(events)
	sessions = s.filterInvalidSessions(sessions)

	timelineStart := time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 0, 0, 0, 0, time.UTC)
	timelineEnd := time.Date(requestTimeStamp.Year(), requestTimeStamp.Month(), requestTimeStamp.Day(), 0, 0, 0, 0, time.UTC)
	dayCount := int(timelineEnd.Sub(timelineStart).Hours() / 24)
	if dayCount == 0 {
		dayCount = 1
	}

	hourlyOverlap := make([]float64, hoursInDay)
	for _, session := range sessions {
		effectiveStart := tools.MaxTime(session.Start, timelineStart)
		effectiveEnd := tools.MinTime(session.End, timelineEnd.Add(24*time.Hour))
		if !effectiveEnd.After(effectiveStart) {
			continue
		}
		startIdx := int(effectiveStart.Sub(timelineStart).Hours())
		endIdx := int(math.Ceil(effectiveEnd.Sub(timelineStart).Hours()))
		for i := startIdx; i < endIdx && i < hoursInDay; i++ {
			blockStart := timelineStart.Add(time.Duration(i) * time.Hour)
			blockEnd := blockStart.Add(time.Hour)
			overlapStart := tools.MaxTime(effectiveStart, blockStart)
			overlapEnd := tools.MinTime(effectiveEnd, blockEnd)
			overlap := overlapEnd.Sub(overlapStart).Seconds()
			if overlap > 0 {
				hourlyOverlap[i] += overlap
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

func (s *Service) filterInvalidSessions(sessions []dto.Session) []dto.Session {
	minDuration := time.Duration(minSessionDurationInMinutes) * time.Minute
	valid := make([]dto.Session, 0, len(sessions))
	for _, sess := range sessions {
		if sess.End.Sub(sess.Start) >= minDuration {
			valid = append(valid, sess)
		}
	}
	return valid
}

func (s *Service) getSessionsFromLogs(logs []*dto.LogData) []dto.Session {
	var sessions []dto.Session
	active := make(map[string]time.Time)
	for _, entry := range logs {
		switch entry.Action {
		case enums.Actions.Connected():
			if _, exists := active[entry.NickName]; !exists {
				active[entry.NickName] = entry.TimeStamp
				continue
			}
			if lastActivity := s.findLastUserActivityTimeStampBefore(logs, entry.NickName, &entry.TimeStamp); lastActivity != nil {
				sessions = append(sessions, dto.Session{
					NickName: entry.NickName,
					Start:    active[entry.NickName],
					End:      *lastActivity,
				})
				delete(active, entry.NickName)
			}
		case enums.Actions.Disconnected():
			if start, exists := active[entry.NickName]; exists {
				sessions = append(sessions, dto.Session{
					NickName: entry.NickName,
					Start:    start,
					End:      entry.TimeStamp,
				})
				delete(active, entry.NickName)
			}
		}
	}
	for nick, start := range active {
		if lastActivity := s.findLastUserActivityTimeStampBefore(logs, nick, nil); lastActivity != nil {
			sessions = append(sessions, dto.Session{
				NickName: nick,
				Start:    start,
				End:      *lastActivity,
			})
		}
	}
	return sessions
}
