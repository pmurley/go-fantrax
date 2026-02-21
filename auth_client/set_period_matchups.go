package auth_client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pmurley/go-fantrax/models"
)

// SetPeriodMatchups saves matchup changes for a specific period by POSTing the
// full league setup form back to the createLeague.go endpoint.
//
// It updates setup.Matchups[period] with the provided matchup pairs, then builds
// the complete form body (all 179 periods, divisions, hidden fields, etc.) and
// submits it. A successful save returns a 302 redirect; any other status is an error.
//
// The setup struct is modified in-place with the new matchups for the given period.
func (c *Client) SetPeriodMatchups(setup *models.LeagueSetupMatchups, period int, matchups []models.MatchupPair) error {
	// Validate that the period exists in the setup data
	if _, exists := setup.Matchups[period]; !exists {
		return fmt.Errorf("period %d not found in setup matchups", period)
	}
	if len(matchups) == 0 {
		return fmt.Errorf("matchups must not be empty")
	}

	// Update the matchups for the target period
	setup.Matchups[period] = matchups

	// Build the full form body
	formBody := BuildFormBody(setup, period)

	// POST to createLeague.go
	postURL := fmt.Sprintf("https://www.fantrax.com/newui/fantasy/createLeague.go?leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("POST", postURL, strings.NewReader(formBody.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}

	cookiesString, err := GetCookies()
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}
	req.Header.Set("Cookie", cookiesString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko)")

	// Use a client that does NOT follow redirects so we can detect the 302
	noRedirectClient := &http.Client{
		Transport: c.Client.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// A successful save returns 302; anything else is an error.
	// Include response body in error for diagnostics.
	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500] + "..."
		}
		return fmt.Errorf("expected 302 redirect on success, got status %d; body: %s", resp.StatusCode, snippet)
	}

	return nil
}

// BuildFormBody assembles the full url.Values form body for the league setup POST.
// This includes all hidden fields, select fields, checkbox fields, team names,
// owner emails, divisions, hardcoded fields, and all 179 periods of matchup data.
func BuildFormBody(setup *models.LeagueSetupMatchups, period int) url.Values {
	form := url.Values{}
	cfg := &setup.FormConfig

	// Hidden fields, with overrides for matchup-edit signals
	for name, value := range cfg.HiddenFields {
		if name == "h2hConfigChangesMade" {
			form.Set(name, "y")
		} else {
			form.Set(name, value)
		}
	}

	// Select fields
	for name, value := range cfg.SelectFields {
		form.Set(name, value)
	}

	// Checkbox shadow fields
	for name, value := range cfg.CheckboxFields {
		form.Set(name, value)
	}

	// Team names and short names
	for teamID, name := range cfg.TeamNames {
		form.Set("teamName_"+teamID, name)
	}
	for teamID, shortName := range cfg.TeamShortNames {
		form.Set("teamShortName_"+teamID, shortName)
	}

	// Owner email fields
	for key, value := range cfg.OwnerEmailFields {
		form.Set(key, value)
	}

	// Division names
	for divID, name := range cfg.DivisionNames {
		form.Set("divisionName_"+divID, name)
	}

	// ~~divisions: repeated key, one per division
	for _, divEntry := range cfg.Divisions {
		form.Add("~~divisions", divEntry)
	}

	// Hardcoded fields required by the form submission
	form.Set("tabId", "Matchups")
	form.Set("gotoNextPage", "false")
	form.Set("divisionName", "")
	form.Set("inviteMessage", "")
	form.Set("calculatedHeadToHeadOpponentType", "1")
	form.Set("playoffMatchupSetConfigId", "")

	// Matchup edit metadata
	form.Set("matchupScoringPeriodToEdit", strconv.Itoa(period))
	form.Set("matchupsEditedManually", "true")

	// All matchup period data: repeated "matchups" key, one per period
	for _, entry := range serializeMatchups(setup) {
		form.Add("matchups", entry)
	}

	return form
}

// serializeMatchups converts the matchup map into a sorted slice of strings,
// one per period, each formatted as "{period}|{away}_{home}|{away}_{home}|...".
func serializeMatchups(setup *models.LeagueSetupMatchups) []string {
	// Collect and sort period numbers
	periods := make([]int, 0, len(setup.Matchups))
	for p := range setup.Matchups {
		periods = append(periods, p)
	}
	sort.Ints(periods)

	result := make([]string, 0, len(periods))
	for _, p := range periods {
		pairs := setup.Matchups[p]
		parts := make([]string, 0, len(pairs)+1)
		parts = append(parts, strconv.Itoa(p))
		for _, pair := range pairs {
			parts = append(parts, pair.AwayTeamID+"_"+pair.HomeTeamID)
		}
		result = append(result, strings.Join(parts, "|"))
	}

	return result
}
