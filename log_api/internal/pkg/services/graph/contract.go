package graph

import "github.com/rumblefrog/go-a2s"

type a2sClient interface {
	QueryPlayer() (*a2s.PlayerInfo, error)
}
