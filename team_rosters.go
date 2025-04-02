package fantrax

import "fmt"

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
func (c *Client) GetTeamRosters(leagueID string, opts ...TeamRosterOption) (*LeagueInfo, error) {
	endpoint := "/general/getTeamRosters"
	params := map[string]string{"leagueId": leagueID}

	teamRosterOptions := &TeamRosterOptions{}

	for _, o := range opts {
		o(teamRosterOptions)
	}

	if teamRosterOptions.period > 0 {
		params["period"] = fmt.Sprintf("%d", teamRosterOptions.period)
	}

	var results LeagueInfo
	err := c.fetchWithCache(endpoint, params, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get team rosters: %w", err)
	}

	return &results, nil
}
