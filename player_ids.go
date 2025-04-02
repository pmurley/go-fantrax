package fantrax

import "fmt"

type Sport string

const (
	MLB    Sport = "MLB"
	NFL    Sport = "NFL"
	NHL    Sport = "NHL"
	NBA    Sport = "NBA"
	NCAAF  Sport = "NCAAF"
	NCAAB  Sport = "NCAAB"
	PGA    Sport = "PGA"
	NASCAR Sport = "NASCAR"
	EPL    Sport = "EPL"
)

// Player represents a player in the system with all optional fields
type Player struct {
	StatsIncId   *int    `json:"statsIncId,omitempty"`
	RotowireId   *int    `json:"rotowireId,omitempty"`
	SportRadarId *string `json:"sportRadarId,omitempty"`
	Name         string  `json:"name"`
	FantraxId    string  `json:"fantraxId"`
	Team         string  `json:"team"`
	Position     string  `json:"position"`
}

// PlayersResponse represents the response from the getPlayerIds endpoint
// It's a map of fantraxId to PlayerStatus details
type PlayersResponse map[string]Player

// GetPlayerIds gets the list of all players in the database for a particular sport
func (c *Client) GetPlayerIds(sport Sport) (*PlayersResponse, error) {
	endpoint := "/general/getPlayerIds"
	params := map[string]string{"sport": string(sport)}

	var results PlayersResponse
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get player IDs: %w", err)
	}

	return &results, nil
}
