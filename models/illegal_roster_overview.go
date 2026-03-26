package models

import (
	"time"
)

// IllegalRosterOverview contains the league-wide illegal roster status
// for all teams across all dates, as returned by the commissioner's
// illegal roster override admin page.
//
// Note: the Fantrax page shows one column per date, not per scoring period.
// Multiple dates may fall within the same scoring period.
type IllegalRosterOverview struct {
	// Dates lists all dates shown on the page, in column order.
	Dates []time.Time

	// Teams contains one entry per team in the league
	Teams []IllegalRosterTeam
}

// IllegalRosterTeam represents a single team's illegal roster status
type IllegalRosterTeam struct {
	TeamID   string
	TeamName string

	// IllegalDates lists the dates where this team has an illegal roster
	IllegalDates []time.Time
}

// HasIllegalRosters returns true if any team in the league has an illegal roster
// on any date.
func (o *IllegalRosterOverview) HasIllegalRosters() bool {
	for _, team := range o.Teams {
		if len(team.IllegalDates) > 0 {
			return true
		}
	}
	return false
}

// TeamsWithIllegalRosters returns only the teams that have at least one illegal date.
func (o *IllegalRosterOverview) TeamsWithIllegalRosters() []IllegalRosterTeam {
	var result []IllegalRosterTeam
	for _, team := range o.Teams {
		if len(team.IllegalDates) > 0 {
			result = append(result, team)
		}
	}
	return result
}

// TeamsWithIllegalRostersForDate returns teams that have an illegal roster
// on the specified date. Only the date portion (year/month/day) is compared.
func (o *IllegalRosterOverview) TeamsWithIllegalRostersForDate(date time.Time) []IllegalRosterTeam {
	var result []IllegalRosterTeam
	for _, team := range o.Teams {
		if team.IsIllegalOnDate(date) {
			result = append(result, team)
		}
	}
	return result
}

// IsIllegalOnDate returns true if the team has an illegal roster on the given date.
// Only the date portion (year/month/day) is compared.
func (t *IllegalRosterTeam) IsIllegalOnDate(date time.Time) bool {
	dy, dm, dd := date.Date()
	for _, d := range t.IllegalDates {
		y, m, day := d.Date()
		if y == dy && m == dm && day == dd {
			return true
		}
	}
	return false
}
