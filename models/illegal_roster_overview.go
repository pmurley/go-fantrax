package models

import (
	"time"
)

// IllegalRosterOverview contains the league-wide illegal roster status
// for all teams across all periods, as returned by the commissioner's
// illegal roster override admin page.
type IllegalRosterOverview struct {
	// Periods maps period number to its date string (e.g., "Mar 25, 2026")
	Periods map[int]string

	// Teams contains one entry per team in the league
	Teams []IllegalRosterTeam
}

// IllegalRosterTeam represents a single team's illegal roster status
type IllegalRosterTeam struct {
	TeamID   string
	TeamName string

	// IllegalPeriods lists the period numbers where this team has an illegal roster
	IllegalPeriods []int
}

// HasIllegalRosters returns true if any team in the league has an illegal roster
// in any period.
func (o *IllegalRosterOverview) HasIllegalRosters() bool {
	for _, team := range o.Teams {
		if len(team.IllegalPeriods) > 0 {
			return true
		}
	}
	return false
}

// TeamsWithIllegalRosters returns only the teams that have at least one illegal period.
func (o *IllegalRosterOverview) TeamsWithIllegalRosters() []IllegalRosterTeam {
	var result []IllegalRosterTeam
	for _, team := range o.Teams {
		if len(team.IllegalPeriods) > 0 {
			result = append(result, team)
		}
	}
	return result
}

// CurrentPeriod returns the period number whose date matches the given time,
// or the most recent period with a date <= that time. Returns 0 if no period matches.
// The provided time should already be in the appropriate timezone (e.g., the user's
// Fantrax timezone via client.UserInfo.Timezone).
func (o *IllegalRosterOverview) CurrentPeriod(now time.Time) int {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	bestPeriod := 0
	var bestDate time.Time
	for period, dateStr := range o.Periods {
		// Parse "Mar 25, 2026" format
		t, err := time.Parse("Jan 2, 2006", dateStr)
		if err != nil {
			continue
		}
		if t.Equal(today) {
			return period
		}
		if t.Before(today) && t.After(bestDate) {
			bestDate = t
			bestPeriod = period
		}
	}
	return bestPeriod
}

// TeamsWithIllegalRostersForPeriod returns teams that have an illegal roster
// for the specified period.
func (o *IllegalRosterOverview) TeamsWithIllegalRostersForPeriod(period int) []IllegalRosterTeam {
	var result []IllegalRosterTeam
	for _, team := range o.Teams {
		if team.HasIllegalPeriod(period) {
			result = append(result, team)
		}
	}
	return result
}

// HasIllegalPeriod returns true if the team has an illegal roster for the given period.
func (t *IllegalRosterTeam) HasIllegalPeriod(period int) bool {
	for _, p := range t.IllegalPeriods {
		if p == period {
			return true
		}
	}
	return false
}
