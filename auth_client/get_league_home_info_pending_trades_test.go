package auth_client

import (
	"encoding/json"
	"testing"
)

// pendingTradesFixture is a trimmed sample of the getLeagueHomeInfo response
// with one proposed trade between two teams: Team Alpha sends Pitcher A to
// Team Beta in exchange for Hitter B.
const pendingTradesFixture = `{
  "responses": [
    {
      "data": {
        "settings": {"leagueName": "Test League", "sportId": "1", "leagueDisplayYear": "2026"},
        "fantasyTeams": [
          {"id": "team1", "name": "Team Alpha", "shortName": "ALPHA"},
          {"id": "team2", "name": "Team Beta", "shortName": "BETA"}
        ],
        "standings": {"header": [], "statsTable": []},
        "matchups": {"titlePeriodInfo": "", "games": [], "noMatchupsMsg": "", "live": false},
        "pendingTransactions": {
          "pendingTransactionSets": [
            {
              "id": "trade123",
              "transactions": [
                {"scorerId": "p1", "sourceTeamId": "team1", "destinationTeamId": "team2"},
                {"scorerId": "h1", "sourceTeamId": "team2", "destinationTeamId": "team1"}
              ]
            }
          ],
          "scorerMap": {
            "p1": {"name": "Pitcher A", "posShortNames": "SP"},
            "h1": {"name": "Hitter B",  "posShortNames": "3B,INF"}
          }
        }
      }
    }
  ]
}`

func TestProcessLeagueHomeInfo_PendingTrades(t *testing.T) {
	var raw LeagueHomeInfoRawResponse
	if err := json.Unmarshal([]byte(pendingTradesFixture), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	info, err := processLeagueHomeInfo(&raw)
	if err != nil {
		t.Fatalf("processLeagueHomeInfo: %v", err)
	}

	if len(info.PendingTrades) != 2 {
		t.Fatalf("expected 2 pending trades, got %d", len(info.PendingTrades))
	}

	// First leg: Pitcher A from Team Alpha → Team Beta
	first := info.PendingTrades[0]
	if first.TradeID != "trade123" {
		t.Errorf("trade[0].TradeID = %q, want %q", first.TradeID, "trade123")
	}
	if first.PlayerName != "Pitcher A" {
		t.Errorf("trade[0].PlayerName = %q, want %q", first.PlayerName, "Pitcher A")
	}
	if first.Position != "SP" {
		t.Errorf("trade[0].Position = %q, want %q", first.Position, "SP")
	}
	if first.FromTeam != "Team Alpha" {
		t.Errorf("trade[0].FromTeam = %q, want %q", first.FromTeam, "Team Alpha")
	}
	if first.ToTeam != "Team Beta" {
		t.Errorf("trade[0].ToTeam = %q, want %q", first.ToTeam, "Team Beta")
	}

	// Second leg: Hitter B from Team Beta → Team Alpha
	second := info.PendingTrades[1]
	if second.TradeID != "trade123" {
		t.Errorf("trade[1].TradeID = %q, want %q (same trade)", second.TradeID, "trade123")
	}
	if second.PlayerName != "Hitter B" {
		t.Errorf("trade[1].PlayerName = %q, want %q", second.PlayerName, "Hitter B")
	}
	if second.Position != "3B,INF" {
		t.Errorf("trade[1].Position = %q, want %q", second.Position, "3B,INF")
	}
	if second.FromTeam != "Team Beta" {
		t.Errorf("trade[1].FromTeam = %q, want %q", second.FromTeam, "Team Beta")
	}
	if second.ToTeam != "Team Alpha" {
		t.Errorf("trade[1].ToTeam = %q, want %q", second.ToTeam, "Team Alpha")
	}
}

// TestProcessLeagueHomeInfo_PendingTradesEmpty verifies that a response with
// no pendingTransactions section returns an empty PendingTrades slice rather
// than failing or producing phantom entries.
func TestProcessLeagueHomeInfo_PendingTradesEmpty(t *testing.T) {
	noPending := `{
		"responses": [{
			"data": {
				"settings": {"leagueName": "X", "sportId": "1", "leagueDisplayYear": "2026"},
				"fantasyTeams": [],
				"standings": {"header": [], "statsTable": []},
				"matchups": {"titlePeriodInfo": "", "games": [], "noMatchupsMsg": "", "live": false}
			}
		}]
	}`
	var raw LeagueHomeInfoRawResponse
	if err := json.Unmarshal([]byte(noPending), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	info, err := processLeagueHomeInfo(&raw)
	if err != nil {
		t.Fatalf("processLeagueHomeInfo: %v", err)
	}
	if len(info.PendingTrades) != 0 {
		t.Errorf("expected 0 pending trades, got %d", len(info.PendingTrades))
	}
}

// TestProcessLeagueHomeInfo_PendingTradesUnknownTeamFallsBack confirms that
// when a transaction references a team ID not in the fantasyTeams list, the
// raw team ID is used as a fallback rather than producing an empty string or
// failing.
func TestProcessLeagueHomeInfo_PendingTradesUnknownTeamFallsBack(t *testing.T) {
	mismatched := `{
		"responses": [{
			"data": {
				"settings": {"leagueName": "X", "sportId": "1", "leagueDisplayYear": "2026"},
				"fantasyTeams": [
					{"id": "known", "name": "Known Team", "shortName": "K"}
				],
				"standings": {"header": [], "statsTable": []},
				"matchups": {"titlePeriodInfo": "", "games": [], "noMatchupsMsg": "", "live": false},
				"pendingTransactions": {
					"pendingTransactionSets": [{
						"id": "t1",
						"transactions": [
							{"scorerId": "s1", "sourceTeamId": "known", "destinationTeamId": "missing"}
						]
					}],
					"scorerMap": {"s1": {"name": "Test Player", "posShortNames": "OF"}}
				}
			}
		}]
	}`
	var raw LeagueHomeInfoRawResponse
	if err := json.Unmarshal([]byte(mismatched), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	info, err := processLeagueHomeInfo(&raw)
	if err != nil {
		t.Fatalf("processLeagueHomeInfo: %v", err)
	}
	if len(info.PendingTrades) != 1 {
		t.Fatalf("expected 1 pending trade, got %d", len(info.PendingTrades))
	}
	tr := info.PendingTrades[0]
	if tr.FromTeam != "Known Team" {
		t.Errorf("FromTeam = %q, want %q", tr.FromTeam, "Known Team")
	}
	if tr.ToTeam != "missing" {
		t.Errorf("ToTeam fallback = %q, want raw ID %q", tr.ToTeam, "missing")
	}
}
