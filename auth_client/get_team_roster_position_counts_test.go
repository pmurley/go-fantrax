package auth_client

import (
	"encoding/json"
	"testing"
)

const testGamesPerPosJSON = `{
  "responses": [
    {
      "data": {
        "gamePlayedPerPosData": {
          "tableData": [
            {"pos": "Catcher (C)", "posShort": "C", "gp": 12, "min": "No min", "max": "No max"},
            {"pos": "Pitcher (P)", "posShort": "P", "gp": 15, "min": "5", "max": "20"}
          ]
        },
        "scMinMaxData": {
          "tableData": [
            {"scoringCategory": "Games Started - Pitching (GS)", "total": "15", "min": "15", "max": "19"},
            {"scoringCategory": "Innings Pitched (IP)", "total": "42", "min": "No min", "max": "No max"}
          ]
        }
      }
    }
  ]
}`

func TestProcessGamesPerPosition(t *testing.T) {
	var raw GamesPerPositionResponse
	if err := json.Unmarshal([]byte(testGamesPerPosJSON), &raw); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	result, err := ProcessGamesPerPosition(&raw)
	if err != nil {
		t.Fatalf("ProcessGamesPerPosition: %v", err)
	}

	if len(result.Positions) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(result.Positions))
	}
	catcher := result.Positions[0]
	if catcher.Min != nil || catcher.Max != nil {
		t.Errorf("expected Catcher min/max nil (No min/No max), got min=%v max=%v", catcher.Min, catcher.Max)
	}
	pitcher := result.Positions[1]
	if pitcher.Min == nil || *pitcher.Min != 5 {
		t.Errorf("expected Pitcher min=5, got %v", pitcher.Min)
	}
	if pitcher.Max == nil || *pitcher.Max != 20 {
		t.Errorf("expected Pitcher max=20, got %v", pitcher.Max)
	}

	if len(result.CategoryLimits) != 2 {
		t.Fatalf("expected 2 category limits, got %d", len(result.CategoryLimits))
	}
	gs := result.CategoryLimits[0]
	if gs.Category != "Games Started - Pitching (GS)" {
		t.Errorf("expected GS category, got %q", gs.Category)
	}
	if gs.Total != 15 {
		t.Errorf("expected total 15, got %d", gs.Total)
	}
	if gs.Min == nil || *gs.Min != 15 {
		t.Errorf("expected min=15, got %v", gs.Min)
	}
	if gs.Max == nil || *gs.Max != 19 {
		t.Errorf("expected max=19, got %v", gs.Max)
	}

	ip := result.CategoryLimits[1]
	if ip.Min != nil || ip.Max != nil {
		t.Errorf("expected IP min/max nil, got min=%v max=%v", ip.Min, ip.Max)
	}
}

func TestProcessGamesPerPosition_NoResponses(t *testing.T) {
	raw := GamesPerPositionResponse{}
	if _, err := ProcessGamesPerPosition(&raw); err == nil {
		t.Fatal("expected error for empty responses")
	}
}
