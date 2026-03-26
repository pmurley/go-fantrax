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

	// Check periods
	if len(overview.Periods) != 3 {
		t.Errorf("expected 3 periods, got %d", len(overview.Periods))
	}
	if overview.Periods[1] != "Mar 25, 2026" {
		t.Errorf("expected period 1 = 'Mar 25, 2026', got %q", overview.Periods[1])
	}

	// Check teams
	if len(overview.Teams) != 3 {
		t.Errorf("expected 3 teams, got %d", len(overview.Teams))
	}

	// Team Alpha: illegal in periods 1 and 2
	alpha := overview.Teams[0]
	if alpha.TeamID != "team1" || alpha.TeamName != "Team Alpha" {
		t.Errorf("unexpected team 0: %+v", alpha)
	}
	if len(alpha.IllegalPeriods) != 2 {
		t.Errorf("expected 2 illegal periods for Alpha, got %d", len(alpha.IllegalPeriods))
	}

	// Team Beta: no illegal periods
	beta := overview.Teams[1]
	if len(beta.IllegalPeriods) != 0 {
		t.Errorf("expected 0 illegal periods for Beta, got %d", len(beta.IllegalPeriods))
	}

	// Team Gamma: illegal in period 3
	gamma := overview.Teams[2]
	if len(gamma.IllegalPeriods) != 1 || gamma.IllegalPeriods[0] != 3 {
		t.Errorf("expected illegal period [3] for Gamma, got %v", gamma.IllegalPeriods)
	}

	// Test helper methods
	if !overview.HasIllegalRosters() {
		t.Error("expected HasIllegalRosters to be true")
	}

	illegal := overview.TeamsWithIllegalRosters()
	if len(illegal) != 2 {
		t.Errorf("expected 2 teams with illegal rosters, got %d", len(illegal))
	}

	if !alpha.HasIllegalPeriod(1) {
		t.Error("expected Alpha to have illegal period 1")
	}
	if alpha.HasIllegalPeriod(3) {
		t.Error("expected Alpha to NOT have illegal period 3")
	}

	// Test TeamsWithIllegalRostersForPeriod
	period1Illegal := overview.TeamsWithIllegalRostersForPeriod(1)
	if len(period1Illegal) != 1 || period1Illegal[0].TeamID != "team1" {
		t.Errorf("expected only Team Alpha illegal in period 1, got %v", period1Illegal)
	}
	period3Illegal := overview.TeamsWithIllegalRostersForPeriod(3)
	if len(period3Illegal) != 1 || period3Illegal[0].TeamID != "team3" {
		t.Errorf("expected only Team Gamma illegal in period 3, got %v", period3Illegal)
	}

	// Test CurrentPeriod
	mar25 := time.Date(2026, 3, 25, 14, 0, 0, 0, time.UTC)
	if p := overview.CurrentPeriod(mar25); p != 1 {
		t.Errorf("expected current period 1 on Mar 25, got %d", p)
	}
	mar26 := time.Date(2026, 3, 26, 8, 0, 0, 0, time.UTC)
	if p := overview.CurrentPeriod(mar26); p != 2 {
		t.Errorf("expected current period 2 on Mar 26, got %d", p)
	}
	// Date before any period should return 0
	mar20 := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	if p := overview.CurrentPeriod(mar20); p != 0 {
		t.Errorf("expected current period 0 before season, got %d", p)
	}
}

func TestParseIllegalRosterOverviewNoTeams(t *testing.T) {
	html := `<html><body>no table here</body></html>`
	_, err := parseIllegalRosterOverview(html)
	if err == nil {
		t.Error("expected error for empty HTML")
	}
}
