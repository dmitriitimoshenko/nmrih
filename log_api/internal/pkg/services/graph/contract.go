package graph

import "github.com/rumblefrog/go-a2s"

type A2SClient interface {
	QueryPlayer() (*a2s.PlayerInfo, error)
}
