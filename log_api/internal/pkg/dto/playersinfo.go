package dto

type PlayersInfo struct {
	Count      int           `json:"count"` // Count of player online
	PlayerInfo []*PlayerInfo `json:"player"`
}

type PlayerInfo struct {
	Name     string  `json:"Name"`     // Name of the player
	Score    uint32  `json:"Score"`    // Player's score (usually "frags" or "kills")
	Duration float32 `json:"Duration"` // Time (in seconds) player has been connected to the server
}
