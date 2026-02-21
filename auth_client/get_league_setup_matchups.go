package auth_client

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pmurley/go-fantrax/models"
)

// GetLeagueSetupMatchups fetches the league setup page and parses it to extract
// all matchup data, team metadata, division structure, and form configuration.
// This uses a direct HTML GET (not the standard JSON POST to /fxpa/req).
func (c *Client) GetLeagueSetupMatchups() (*models.LeagueSetupMatchups, error) {
	html, err := c.fetchLeagueSetupHTML()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league setup page: %w", err)
	}

	matchups, err := parseMatchupMap(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse matchup map: %w", err)
	}

	teams, err := parseTeams(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse teams: %w", err)
	}

	divisions, err := parseDivisions(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse divisions: %w", err)
	}

	formConfig, err := parseFormConfig(html, teams, divisions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse form config: %w", err)
	}

	return &models.LeagueSetupMatchups{
		Teams:      teams,
		Divisions:  divisions,
		Matchups:   matchups,
		FormConfig: *formConfig,
	}, nil
}

// fetchLeagueSetupHTML makes a GET request to the league setup page and returns
// the raw HTML. This bypasses the standard Do() method which sets JSON headers.
func (c *Client) fetchLeagueSetupHTML() (string, error) {
	url := fmt.Sprintf("https://www.fantrax.com/newui/fantasy/createLeague.go?goto=1&leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	cookiesString, err := GetCookies()
	if err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}
	req.Header.Set("Cookie", cookiesString)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko)")

	// Use the embedded http.Client directly to avoid JSON headers from Do()
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// parseMatchupMap extracts the matchupMap JS variable from the HTML and parses
// it into a map of period number -> matchup pairs.
//
// Source format:
//
//	var matchupMap = {
//	  '1':['awayId_homeId','awayId_homeId',...],
//	  '2':['awayId_homeId',...],
//	  ...
//	};
func parseMatchupMap(html string) (map[int][]models.MatchupPair, error) {
	// Extract the matchupMap block
	outerRe := regexp.MustCompile(`var\s+matchupMap\s*=\s*\{([\s\S]*?)\};`)
	outerMatch := outerRe.FindStringSubmatch(html)
	if outerMatch == nil {
		return nil, fmt.Errorf("matchupMap not found in HTML")
	}
	mapContent := outerMatch[1]

	// Extract each period's matchup array
	periodRe := regexp.MustCompile(`'(\d+)'\s*:\s*\[(.*?)\]`)
	periodMatches := periodRe.FindAllStringSubmatch(mapContent, -1)
	if len(periodMatches) == 0 {
		return nil, fmt.Errorf("no periods found in matchupMap")
	}

	pairRe := regexp.MustCompile(`'([^']+)'`)
	result := make(map[int][]models.MatchupPair, len(periodMatches))

	for _, pm := range periodMatches {
		period, err := strconv.Atoi(pm[1])
		if err != nil {
			return nil, fmt.Errorf("invalid period number %q: %w", pm[1], err)
		}

		arrayContent := pm[2]
		pairMatches := pairRe.FindAllStringSubmatch(arrayContent, -1)

		var pairs []models.MatchupPair
		for _, pairMatch := range pairMatches {
			parts := strings.SplitN(pairMatch[1], "_", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid matchup pair format: %q", pairMatch[1])
			}
			pairs = append(pairs, models.MatchupPair{
				AwayTeamID: parts[0],
				HomeTeamID: parts[1],
			})
		}

		result[period] = pairs
	}

	return result, nil
}

// parseTeams extracts team data and owner info from addTeam() JS calls in the HTML.
// Teams with multiple owners appear multiple times; owners are collected per team.
//
// Source format:
//
//	addTeam('Name', 'SHORT', 'email', 'teamId', 'userId', isCommissioner, joinedLeague, ...);
//
// The JS function transforms userId='NULL' into 'NULL_N' with an incrementing
// counter. We replicate that logic here so owner email form field keys match.
func parseTeams(html string) ([]models.LeagueSetupTeam, error) {
	re := regexp.MustCompile(`addTeam\('([^']*)',\s*'([^']*)',\s*'([^']*)',\s*'([^']*)',\s*'([^']*)',\s*(true|false),\s*(true|false)`)
	matches := re.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no addTeam() calls found in HTML")
	}

	// Track teams by ID to preserve order and collect owners
	teamIndex := make(map[string]int) // teamID -> index in teams slice
	var teams []models.LeagueSetupTeam
	uniqueTempUserID := 0 // Mirrors JS var uniqueTempUserId

	for _, m := range matches {
		name := m[1]
		shortName := m[2]
		email := m[3]
		teamID := m[4]
		userID := m[5]
		isCommissioner := m[6] == "true"
		joinedLeague := m[7] == "true"

		// Replicate JS logic: if userId is 'NULL', assign 'NULL_N'
		if userID == "NULL" {
			userID = fmt.Sprintf("NULL_%d", uniqueTempUserID)
			uniqueTempUserID++
		}

		owner := models.TeamOwner{
			Email:          email,
			UserID:         userID,
			IsCommissioner: isCommissioner,
			JoinedLeague:   joinedLeague,
		}

		if idx, exists := teamIndex[teamID]; exists {
			// Add additional owner to existing team
			teams[idx].Owners = append(teams[idx].Owners, owner)
		} else {
			// New team
			teamIndex[teamID] = len(teams)
			teams = append(teams, models.LeagueSetupTeam{
				TeamID:    teamID,
				Name:      name,
				ShortName: shortName,
				Owners:    []models.TeamOwner{owner},
			})
		}
	}

	return teams, nil
}

// parseDivisions extracts division structure from divisionName_ inputs and
// __removeTeamFromDivision() calls in the HTML.
func parseDivisions(html string) ([]models.LeagueSetupDivision, error) {
	// Extract division names from input elements.
	// Use an alphanumeric-only ID pattern to skip JS template strings like
	// divisionName_' + tempId + ' that appear in script blocks.
	nameRe := regexp.MustCompile(`<input[^>]*name="divisionName_([a-zA-Z0-9]+)"[^>]*value="([^"]*)"`)
	nameMatches := nameRe.FindAllStringSubmatch(html, -1)
	if len(nameMatches) == 0 {
		return nil, fmt.Errorf("no division names found in HTML")
	}

	divMap := make(map[string]*models.LeagueSetupDivision)
	var divOrder []string
	for _, m := range nameMatches {
		divID := m[1]
		if _, exists := divMap[divID]; !exists {
			divMap[divID] = &models.LeagueSetupDivision{
				DivisionID: divID,
				Name:       m[2],
			}
			divOrder = append(divOrder, divID)
		}
	}

	// Extract team assignments from __removeTeamFromDivision() calls
	// Pattern: __removeTeamFromDivision('tbl_{divId}', '{teamId}', false)
	teamRe := regexp.MustCompile(`__removeTeamFromDivision\('tbl_(\w+)',\s*'(\w+)'`)
	teamMatches := teamRe.FindAllStringSubmatch(html, -1)
	for _, m := range teamMatches {
		divID := m[1]
		teamID := m[2]
		if div, ok := divMap[divID]; ok {
			// Avoid duplicates
			found := false
			for _, existingID := range div.TeamIDs {
				if existingID == teamID {
					found = true
					break
				}
			}
			if !found {
				div.TeamIDs = append(div.TeamIDs, teamID)
			}
		}
	}

	// Return divisions in the order they appeared
	divisions := make([]models.LeagueSetupDivision, 0, len(divOrder))
	for _, divID := range divOrder {
		divisions = append(divisions, *divMap[divID])
	}

	return divisions, nil
}

// parseFormConfig extracts all form field values needed to echo back when
// POSTing matchup changes.
func parseFormConfig(html string, teams []models.LeagueSetupTeam, divisions []models.LeagueSetupDivision) (*models.LeagueSetupFormConfig, error) {
	config := &models.LeagueSetupFormConfig{
		HiddenFields:     make(map[string]string),
		SelectFields:     make(map[string]string),
		CheckboxFields:   make(map[string]string),
		TeamNames:        make(map[string]string),
		TeamShortNames:   make(map[string]string),
		OwnerEmailFields: make(map[string]string),
		DivisionNames:    make(map[string]string),
	}

	// Parse hidden input fields
	// Handles both name="x" value="y" and value="y" name="x" orderings
	hiddenRe := regexp.MustCompile(`<input[^>]*type="hidden"[^>]*>`)
	hiddenMatches := hiddenRe.FindAllString(html, -1)
	nameRe := regexp.MustCompile(`name="([^"]+)"`)
	valueRe := regexp.MustCompile(`value="([^"]*)"`)

	for _, tag := range hiddenMatches {
		nameMatch := nameRe.FindStringSubmatch(tag)
		valueMatch := valueRe.FindStringSubmatch(tag)
		if nameMatch == nil {
			continue
		}
		name := nameMatch[1]
		value := ""
		if valueMatch != nil {
			value = valueMatch[1]
		}

		// Skip JS template strings from <script> blocks (contain ' + or ')
		if strings.Contains(name, "'") || strings.Contains(value, "'") {
			continue
		}

		// Categorize by field name prefix
		if strings.HasPrefix(name, "_") {
			config.CheckboxFields[name] = value
		} else {
			config.HiddenFields[name] = value
		}
	}

	// Parse select fields with selected options
	parseSelectFields(html, config)

	// Parse text input fields for dates and other values
	textInputRe := regexp.MustCompile(`<input[^>]*type="text"[^>]*>`)
	textMatches := textInputRe.FindAllString(html, -1)
	for _, tag := range textMatches {
		nameMatch := nameRe.FindStringSubmatch(tag)
		valueMatch := valueRe.FindStringSubmatch(tag)
		if nameMatch == nil || valueMatch == nil {
			continue
		}
		name := nameMatch[1]
		// Only include form-relevant fields (startDate, endDate), not division names
		if name == "startDate" || name == "endDate" {
			config.HiddenFields[name] = valueMatch[1]
		}
	}

	// Parse checked checkboxes
	checkboxRe := regexp.MustCompile(`<input[^>]*type="checkbox"[^>]*checked[^>]*>`)
	checkboxMatches := checkboxRe.FindAllString(html, -1)
	for _, tag := range checkboxMatches {
		nameMatch := nameRe.FindStringSubmatch(tag)
		if nameMatch == nil {
			continue
		}
		vMatch := valueRe.FindStringSubmatch(tag)
		if vMatch != nil {
			config.HiddenFields[nameMatch[1]] = vMatch[1]
		}
	}

	// Build team name/short name maps from parsed teams
	for _, team := range teams {
		config.TeamNames[team.TeamID] = team.Name
		config.TeamShortNames[team.TeamID] = team.ShortName
	}

	// Build owner email form fields from parsed team owner data.
	// Only owners where !IsCommissioner && !JoinedLeague get an <input> rendered
	// by the JS addTeam() function. These are the ones included in the POST.
	for _, team := range teams {
		for _, owner := range team.Owners {
			if !owner.IsCommissioner && !owner.JoinedLeague {
				key := fmt.Sprintf("teamOwnerEmail,%s,%s,%s", owner.Email, team.TeamID, owner.UserID)
				config.OwnerEmailFields[key] = owner.Email
			}
		}
	}

	// Build division name fields for POST reconstruction
	for _, div := range divisions {
		config.DivisionNames[div.DivisionID] = div.Name
	}

	// Build ~~divisions entries: one per division, each formatted as
	// "{divId}={teamId1}|{teamId2}|...". The POST sends each as a separate
	// ~~divisions form field (repeated key), not a single comma-joined string.
	for _, div := range divisions {
		if len(div.TeamIDs) > 0 {
			config.Divisions = append(config.Divisions, div.DivisionID+"="+strings.Join(div.TeamIDs, "|"))
		}
	}

	return config, nil
}

// parseSelectFields extracts select element names and their selected option values.
func parseSelectFields(html string, config *models.LeagueSetupFormConfig) {
	// Find all select elements with name attributes
	// We use a non-greedy match to find each select...option...selected...value block
	selectRe := regexp.MustCompile(`(?s)<select[^>]*name="([^"]+)"[^>]*>(.*?)</select>`)
	selectMatches := selectRe.FindAllStringSubmatch(html, -1)

	selectedValueRe := regexp.MustCompile(`<option[^>]*value="([^"]*)"[^>]*selected[^>]*>`)
	selectedValueAltRe := regexp.MustCompile(`<option[^>]*selected[^>]*value="([^"]*)"[^>]*>`)

	for _, sm := range selectMatches {
		name := sm[1]
		optionsHTML := sm[2]

		// Try to find the selected option value
		match := selectedValueRe.FindStringSubmatch(optionsHTML)
		if match == nil {
			match = selectedValueAltRe.FindStringSubmatch(optionsHTML)
		}
		// Also handle bare "selected" without ="selected" (e.g., <option value="DAILY" selected>)
		if match == nil {
			bareSelectedRe := regexp.MustCompile(`<option[^>]*value="([^"]*)"[^>]*\bselected\b`)
			match = bareSelectedRe.FindStringSubmatch(optionsHTML)
		}

		if match != nil {
			config.SelectFields[name] = match[1]
		}
	}
}

// GetMatchupsByPeriod returns the matchup pairs for a specific period from the
// parsed league setup data.
func GetMatchupsByPeriod(setup *models.LeagueSetupMatchups, period int) []models.MatchupPair {
	return setup.Matchups[period]
}

// GetTeamByID looks up a team by its ID in the parsed league setup data.
func GetTeamByID(setup *models.LeagueSetupMatchups, teamID string) *models.LeagueSetupTeam {
	for i := range setup.Teams {
		if setup.Teams[i].TeamID == teamID {
			return &setup.Teams[i]
		}
	}
	return nil
}

// GetSortedPeriods returns all period numbers from the matchup map in sorted order.
func GetSortedPeriods(setup *models.LeagueSetupMatchups) []int {
	periods := make([]int, 0, len(setup.Matchups))
	for p := range setup.Matchups {
		periods = append(periods, p)
	}
	sort.Ints(periods)
	return periods
}
