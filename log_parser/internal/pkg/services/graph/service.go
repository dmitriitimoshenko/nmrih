package graph

import (
	"log"
	"sort"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/enums"
	"github.com/rumblefrog/go-a2s"
)

const (
	topPlayersCount = 32
	topCountries    = 9
)

type Service struct {
	a2sClient A2SClient
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) TopTimeSpent(logs []*dto.LogData) dto.TopTimeSpentList {
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

	log.Printf("[TopTimeSpent] TotalSessionsDurations: %+v\n", totalSessionsDurations)

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

	log.Printf("[TopTimeSpent] TopTimeSpent: %+v\n", topTimeSpentList)

	return topTimeSpentList
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
