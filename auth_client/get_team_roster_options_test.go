package auth_client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTeamRosterInfoOptions_SetFields verifies that WithScoringCategoryType
// and WithStatsType populate the unexported options struct.
func TestTeamRosterInfoOptions_SetFields(t *testing.T) {
	options := &teamRosterInfoOptions{}
	WithScoringCategoryType("1")(options)
	WithStatsType("2")(options)

	if options.scoringCategoryType != "1" {
		t.Errorf("scoringCategoryType = %q, want %q", options.scoringCategoryType, "1")
	}
	if options.statsType != "2" {
		t.Errorf("statsType = %q, want %q", options.statsType, "2")
	}
}

// TestGetTeamRosterInfoRequest_MarshalsNewFields verifies that when the new
// optional fields are set, the marshaled JSON includes them under the
// expected keys (the wire-format names the Fantrax API requires).
func TestGetTeamRosterInfoRequest_MarshalsNewFields(t *testing.T) {
	req := GetTeamRosterInfoRequest{
		LeagueID:            "league123",
		Reload:              "1",
		Period:              "5",
		TeamID:              "team1",
		ScoringCategoryType: "1",
		StatsType:           "2",
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(raw)
	for _, want := range []string{`"scoringCategoryType":"1"`, `"statsType":"2"`} {
		if !strings.Contains(got, want) {
			t.Errorf("marshaled JSON %s missing %s", got, want)
		}
	}
}

// TestGetTeamRosterInfoRequest_OmitsEmptyOptionalFields confirms that the
// new fields use json:"...,omitempty" so callers who don't set the options
// get the same wire payload as before this change.
func TestGetTeamRosterInfoRequest_OmitsEmptyOptionalFields(t *testing.T) {
	req := GetTeamRosterInfoRequest{
		LeagueID: "league123",
		Reload:   "1",
		Period:   "5",
		TeamID:   "team1",
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(raw)
	for _, unwanted := range []string{"scoringCategoryType", "statsType"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("marshaled JSON %s should not contain %s when unset", got, unwanted)
		}
	}
}
