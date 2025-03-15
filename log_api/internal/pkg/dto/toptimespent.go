package dto

import "time"

type TopTimeSpent struct {
	NickName  string        `json:"nick_name"`
	TimeSpent time.Duration `json:"time_spent"`
}

type TopTimeSpentList []*TopTimeSpent
