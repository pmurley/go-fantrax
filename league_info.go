package fantrax

import "fmt"

// LeagueInfo represents the response from the getLeagueInfo endpoint
type LeagueInfo struct {
	DraftSettings  DraftSettings           `json:"draftSettings"`
	Matchups       []MatchupPeriod         `json:"matchups"`
	RosterInfo     RosterInfo              `json:"rosterInfo"`
	PlayerStatuses map[string]PlayerStatus `json:"playerInfo"`
	PoolSettings   PoolSettings            `json:"poolSettings"`
	ScoringSystem  ScoringSystem           `json:"scoringSystem"`
	TeamInfo       map[string]TeamInfo     `json:"teamInfo"`
	DraftType      string                  `json:"draftType"`
}

// DraftSettings contains information about the draft configuration
type DraftSettings struct {
	DraftType string `json:"draftType"`
}

// MatchupPeriod represents a period of matchups in the schedule
type MatchupPeriod struct {
	Period      int       `json:"period"`
	MatchupList []Matchup `json:"matchupList"`
}

// Matchup represents a single matchup between two teams
type Matchup struct {
	Away Team `json:"away"`
	Home Team `json:"home"`
}

// Team represents a team in a matchup
type Team struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	ShortName string `json:"shortName"`
}

// RosterInfo contains roster configuration details
type RosterInfo struct {
	PositionConstraints    map[string]PositionConstraint `json:"positionConstraints"`
	MaxTotalPlayers        int                           `json:"maxTotalPlayers"`
	MaxTotalActivePlayers  int                           `json:"maxTotalActivePlayers"`
	MaxTotalReservePlayers int                           `json:"maxTotalReservePlayers"`
}

// PositionConstraint defines limits for a specific position
type PositionConstraint struct {
	MaxActive int `json:"maxActive"`
}

// PlayerStatus represents a player's eligibility and status
type PlayerStatus struct {
	EligiblePos string `json:"eligiblePos"`
	Status      string `json:"status"`
}

// PoolSettings contains player pool configuration
type PoolSettings struct {
	DuplicatePlayerType string `json:"duplicatePlayerType"`
	PlayerSourceType    string `json:"playerSourceType"`
}

// ScoringSystem defines the scoring rules for the league
type ScoringSystem struct {
	ScoringCategories       ScoringCategories        `json:"scoringCategories"`
	ScoringCategorySettings []ScoringCategorySetting `json:"scoringCategorySettings"`
	Type                    string                   `json:"type"`
}

// ScoringCategories contains scoring categories organized by group
type ScoringCategories struct {
	HITTING  map[string]map[string]string `json:"HITTING"`
	PITCHING map[string]map[string]string `json:"PITCHING"`
}

// ScoringCategorySetting represents a group of scoring configurations
type ScoringCategorySetting struct {
	Configs []ScoringConfig `json:"configs"`
	Group   Group           `json:"group"`
}

// ScoringConfig defines scoring for a specific category and position
type ScoringConfig struct {
	Position        Position        `json:"position"`
	Cumulative      bool            `json:"cumulative"`
	ScoringCategory ScoringCategory `json:"scoringCategory"`
	Points          float64         `json:"points"`
}

// Position represents a player position
type Position struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	ShortName string `json:"shortName"`
}

// ScoringCategory represents a specific scoring stat category
type ScoringCategory struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	ShortName string `json:"shortName"`
}

// Group represents a group of scoring categories
type Group struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	ShortName string `json:"shortName"`
}

// TeamInfo contains information about a team
type TeamInfo struct {
	Division string `json:"division"`
	Name     string `json:"name"`
	ID       string `json:"id"`
}

// GetLeagueInfo fetches draft results for a specific league
func (c *Client) GetLeagueInfo(leagueID string) (*LeagueInfo, error) {
	endpoint := "/general/getLeagueInfo"
	params := map[string]string{"leagueId": leagueID}

	var results LeagueInfo
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get league info: %w", err)
	}

	return &results, nil
}
