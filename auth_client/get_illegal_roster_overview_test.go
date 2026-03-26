package auth_client

import (
	"testing"
	"time"
)

func TestParseIllegalRosterOverview(t *testing.T) {
	html := `
		<table id="tblOv" class="fantTable illegalRosterOverrideTable">
			<tr>
				<th>Team</th>
				<th class="center" title="(Mar 25, 2026)">1</th>
				<th class="center" title="(Mar 26, 2026)">2</th>
				<th class="center" title="(Mar 27, 2026)">3</th>
			</tr>
			<tr>
				<td class="name"><a href="/fantasy/league/abc123/team/roster;teamId=team1">Team Alpha</a></td><td id="team1_1" ovType="1" illegal="T">
				</td><td id="team1_2" ovType="1" illegal="T">
				</td><td id="team1_3" ovType="1">
				</td>
			</tr>
			<tr>
				<td class="name"><a href="/fantasy/league/abc123/team/roster;teamId=team2">Team Beta</a></td><td id="team2_1" ovType="1">
				</td><td id="team2_2" ovType="1">
				</td><td id="team2_3" ovType="1">
				</td>
			</tr>
			<tr>
				<td class="name"><a href="/fantasy/league/abc123/team/roster;teamId=team3">Team Gamma</a></td><td id="team3_1" ovType="1">
				</td><td id="team3_2" ovType="1">
				</td><td id="team3_3" ovType="1" illegal="T">
				</td>
			</tr>
		</table>
	`

	overview, err := parseIllegalRosterOverview(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check dates
	if len(overview.Dates) != 3 {
		t.Errorf("expected 3 dates, got %d", len(overview.Dates))
	}
	expectedFirst := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !overview.Dates[0].Equal(expectedFirst) {
		t.Errorf("expected first date = %v, got %v", expectedFirst, overview.Dates[0])
	}

	// Check teams
	if len(overview.Teams) != 3 {
		t.Errorf("expected 3 teams, got %d", len(overview.Teams))
	}

	// Team Alpha: illegal on Mar 25 and Mar 26
	alpha := overview.Teams[0]
	if alpha.TeamID != "team1" || alpha.TeamName != "Team Alpha" {
		t.Errorf("unexpected team 0: %+v", alpha)
	}
	if len(alpha.IllegalDates) != 2 {
		t.Errorf("expected 2 illegal dates for Alpha, got %d", len(alpha.IllegalDates))
	}

	// Team Beta: no illegal dates
	beta := overview.Teams[1]
	if len(beta.IllegalDates) != 0 {
		t.Errorf("expected 0 illegal dates for Beta, got %d", len(beta.IllegalDates))
	}

	// Team Gamma: illegal on Mar 27
	gamma := overview.Teams[2]
	mar27 := time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)
	if len(gamma.IllegalDates) != 1 || !gamma.IllegalDates[0].Equal(mar27) {
		t.Errorf("expected illegal date [Mar 27] for Gamma, got %v", gamma.IllegalDates)
	}

	// Test helper methods
	if !overview.HasIllegalRosters() {
		t.Error("expected HasIllegalRosters to be true")
	}

	illegal := overview.TeamsWithIllegalRosters()
	if len(illegal) != 2 {
		t.Errorf("expected 2 teams with illegal rosters, got %d", len(illegal))
	}

	mar25 := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	if !alpha.IsIllegalOnDate(mar25) {
		t.Error("expected Alpha to be illegal on Mar 25")
	}
	if alpha.IsIllegalOnDate(mar27) {
		t.Error("expected Alpha to NOT be illegal on Mar 27")
	}

	// Test TeamsWithIllegalRostersForDate
	mar25Illegal := overview.TeamsWithIllegalRostersForDate(mar25)
	if len(mar25Illegal) != 1 || mar25Illegal[0].TeamID != "team1" {
		t.Errorf("expected only Team Alpha illegal on Mar 25, got %v", mar25Illegal)
	}
	mar27Illegal := overview.TeamsWithIllegalRostersForDate(mar27)
	if len(mar27Illegal) != 1 || mar27Illegal[0].TeamID != "team3" {
		t.Errorf("expected only Team Gamma illegal on Mar 27, got %v", mar27Illegal)
	}
}

func TestParseIllegalRosterOverviewNoTeams(t *testing.T) {
	html := `<html><body>no table here</body></html>`
	_, err := parseIllegalRosterOverview(html)
	if err == nil {
		t.Error("expected error for empty HTML")
	}
}
