package fantrax

import "fmt"

// PlayerInfo represents a player with their Average Draft Position (ADP)
type PlayerInfo struct {
	Pos  string  `json:"pos"`
	Name string  `json:"name"`
	ID   string  `json:"id"`
	ADP  float64 `json:"ADP"`
}

// PlayerInfoResponse represents the response from the getAdp endpoint
// It's a slice of PlayerInfo objects
type PlayerInfoResponse []PlayerInfo

type PlayerInfoOptions struct {
	start            int
	limit            int
	order            string
	position         string
	showAllPositions bool
}

type PlayerInfoOption func(*PlayerInfoOptions)

func WithStart(start int) PlayerInfoOption {
	return func(o *PlayerInfoOptions) {
		o.start = start
	}
}

func WithLimit(limit int) PlayerInfoOption {
	return func(o *PlayerInfoOptions) {
		o.limit = limit
	}
}

func WithOrder(order string) PlayerInfoOption {
	return func(o *PlayerInfoOptions) {
		o.order = order
	}
}

func WithPosition(position string) PlayerInfoOption {
	return func(o *PlayerInfoOptions) {
		o.position = position
	}
}

func WithShowAllPositions(showAllPositions bool) PlayerInfoOption {
	return func(o *PlayerInfoOptions) {
		o.showAllPositions = showAllPositions
	}
}

// GetPlayerInfo gets player info including ADP, and allows for sorting, filtering, and limiting results
func (c *Client) GetPlayerInfo(sport Sport, opts ...PlayerInfoOption) (*PlayerInfoResponse, error) {
	endpoint := "/general/getAdp"
	params := map[string]string{"sport": string(sport)}

	playerInfoOptions := &PlayerInfoOptions{}
	for _, o := range opts {
		o(playerInfoOptions)
	}

	if playerInfoOptions.start != 0 {
		params["start"] = fmt.Sprintf("%d", playerInfoOptions.start)
	}
	if playerInfoOptions.limit > 0 {
		params["limit"] = fmt.Sprintf("%d", playerInfoOptions.limit)
	}
	if playerInfoOptions.order != "" {
		params["order"] = playerInfoOptions.order
	}
	if playerInfoOptions.position != "" {
		params["position"] = playerInfoOptions.position
	}
	if playerInfoOptions.showAllPositions {
		params["showAllPositions"] = "true"
	}

	var results PlayerInfoResponse
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get player info: %w", err)
	}

	return &results, nil
}
