package fantrax

import "fmt"

// DraftResults represents the response from the getDraftResults endpoint
type DraftResults struct {
	DraftDate  string      `json:"draftDate"`
	DraftPicks []DraftPick `json:"draftPicks"`
	DraftState string      `json:"draftState"`
	EndDate    string      `json:"endDate"`
	DraftOrder []string    `json:"draftOrder"`
	DraftType  string      `json:"draftType"`
	StartDate  string      `json:"startDate"`
}

// DraftPick represents a single draft pick in the results
type DraftPick struct {
	Round       int    `json:"round"`
	Pick        int    `json:"pick"`
	TeamID      string `json:"teamId"`
	Time        int64  `json:"time"`
	PickInRound int    `json:"pickInRound"`
	PlayerID    string `json:"playerId"`
}

// GetDraftResults fetches draft results for a specific league
func (c *Client) GetDraftResults(leagueID string) (*DraftResults, error) {
	endpoint := "/general/getDraftResults"
	params := map[string]string{"leagueId": leagueID}

	var results DraftResults
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get draft results: %w", err)
	}

	return &results, nil
}
