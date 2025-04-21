package fantrax

import "fmt"

// LeagueRosters represents the top-level response from the team rosters endpoint
type LeagueRosters struct {
	Period  int                       `json:"period"`
	Rosters map[string]TeamRosterInfo `json:"rosters"`
}

// TeamRosterInfo represents information about a team's roster
type TeamRosterInfo struct {
	TeamName    string       `json:"teamName"`
	RosterItems []RosterItem `json:"rosterItems"`
}

// RosterItem represents a player on a team's roster
type RosterItem struct {
	ID       string `json:"id"`
	Position string `json:"position"`
	Status   string `json:"status"`
}

// RosterStatus represents the possible statuses for a player
type RosterStatus string

// Player status constants
const (
	StatusActive         RosterStatus = "ACTIVE"
	StatusReserve        RosterStatus = "RESERVE"
	StatusMinors         RosterStatus = "MINORS"
	StatusInjuredReserve RosterStatus = "INJURED_RESERVE"
)

type TeamRosterOptions struct {
	period int
}

type TeamRosterOption func(*TeamRosterOptions)

func WithPeriod(period int) TeamRosterOption {
	return func(o *TeamRosterOptions) {
		o.period = period
	}
}

// GetTeamRosters gets all team rosters for a specific league and period
func (c *Client) GetTeamRosters(opts ...TeamRosterOption) (*LeagueRosters, error) {
	endpoint := "/general/getTeamRosters"
	params := map[string]string{"leagueId": c.LeagueId}

	teamRosterOptions := &TeamRosterOptions{}

	for _, o := range opts {
		o(teamRosterOptions)
	}

	if teamRosterOptions.period > 0 {
		params["period"] = fmt.Sprintf("%d", teamRosterOptions.period)
	}

	var results LeagueRosters
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get team rosters: %w", err)
	}

	return &results, nil
}
