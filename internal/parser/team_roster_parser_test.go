package parser

import (
	"os"
	"testing"
)

func TestParseTeamRosterResponse(t *testing.T) {
	// Read the example response file
	data, err := os.ReadFile("../../example_responses/get_team_roster_info_response.json")
	if err != nil {
		t.Fatalf("Failed to read example response: %v", err)
	}

	// Parse the response
	roster, err := ParseTeamRosterResponse(data)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify team info
	if roster.TeamInfo.OwnerName != "Cyclone852" {
		t.Errorf("Expected owner name 'Cyclone852', got '%s'", roster.TeamInfo.OwnerName)
	}
	if roster.TeamInfo.Record != "30-30-0" {
		t.Errorf("Expected record '30-30-0', got '%s'", roster.TeamInfo.Record)
	}
	if roster.TeamInfo.Rank != "15th" {
		t.Errorf("Expected rank '15th', got '%s'", roster.TeamInfo.Rank)
	}

	// Verify claim budget
	if roster.ClaimBudget != 99999998.49 {
		t.Errorf("Expected claim budget 99999998.49, got %f", roster.ClaimBudget)
	}

	// Verify roster sizes (excluding empty slots)
	if len(roster.ActiveRoster) != 23 {
		t.Errorf("Expected 23 active players (excluding empty slots), got %d", len(roster.ActiveRoster))
	}
	if len(roster.ReserveRoster) != 18 {
		t.Errorf("Expected 18 reserve players (excluding empty slots), got %d", len(roster.ReserveRoster))
	}

	// Verify first player details
	if len(roster.ActiveRoster) > 0 {
		firstPlayer := roster.ActiveRoster[0]
		if firstPlayer.Name != "Yainer Diaz" {
			t.Errorf("Expected first player 'Yainer Diaz', got '%s'", firstPlayer.Name)
		}
		if firstPlayer.TeamName != "Houston Astros" {
			t.Errorf("Expected team 'Houston Astros', got '%s'", firstPlayer.TeamName)
		}
		if firstPlayer.PlayerID != "0524l" {
			t.Errorf("Expected player ID '0524l', got '%s'", firstPlayer.PlayerID)
		}

		// Check positions
		expectedPositions := 2 // C and UTIL
		if len(firstPlayer.Positions) != expectedPositions {
			t.Errorf("Expected %d positions, got %d", expectedPositions, len(firstPlayer.Positions))
		}

		// Check next game info
		if firstPlayer.NextGame == nil {
			t.Error("Expected next game info for first player")
		} else {
			if firstPlayer.NextGame.Opponent != "@PIT" {
				t.Errorf("Expected opponent '@PIT', got '%s'", firstPlayer.NextGame.Opponent)
			}
		}
	}

	// Verify league teams
	if len(roster.LeagueTeams) != 30 {
		t.Errorf("Expected 30 league teams, got %d", len(roster.LeagueTeams))
	}

	// Check a few teams
	if len(roster.LeagueTeams) > 0 {
		firstTeam := roster.LeagueTeams[0]
		if firstTeam.Name != "51st State Freedom Flotilla" {
			t.Errorf("Expected first team '51st State Freedom Flotilla', got '%s'", firstTeam.Name)
		}
	}

	t.Logf("Successfully parsed roster with %d active and %d reserve players",
		len(roster.ActiveRoster), len(roster.ReserveRoster))
}
