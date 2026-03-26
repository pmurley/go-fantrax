package auth_client

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pmurley/go-fantrax/models"
)

// GetIllegalRosterOverview fetches the commissioner's illegal roster override page
// and parses it to determine which teams have illegal rosters and on which dates.
// This is a single call that covers all teams in the league.
func (c *Client) GetIllegalRosterOverview() (*models.IllegalRosterOverview, error) {
	html, err := c.fetchIllegalRosterHTML()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch illegal roster page: %w", err)
	}

	return parseIllegalRosterOverview(html)
}

// fetchIllegalRosterHTML makes a GET request to the illegal roster override admin page.
func (c *Client) fetchIllegalRosterHTML() (string, error) {
	url := fmt.Sprintf("https://www.fantrax.com/newui/fantasy/illegalRosterOverrideAdmin.go?leagueId=%s", c.LeagueID)
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

// parseIllegalRosterOverview parses the HTML from the illegal roster override admin page.
//
// The page contains a table (id="tblOv") with:
//   - Header row: <th> cells where each column has title="(Mon DD, YYYY)" and text content = column number
//     Note: these column numbers are sequential day indices, NOT scoring period numbers.
//   - Data rows: one per team, with <td class="name"> containing team name/link,
//     followed by <td id="{teamId}_{colNum}" illegal="T"> for illegal dates
func parseIllegalRosterOverview(html string) (*models.IllegalRosterOverview, error) {
	overview := &models.IllegalRosterOverview{}

	// Extract date headers: <th class="center" title="(Mar 25, 2026)">1</th>
	// Map column number -> parsed date
	columnDates := make(map[int]time.Time)
	headerRe := regexp.MustCompile(`<th[^>]*title="\(([^)]+)\)"[^>]*>(\d+)</th>`)
	headerMatches := headerRe.FindAllStringSubmatch(html, -1)
	for _, m := range headerMatches {
		colNum, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		t, err := time.Parse("Jan 2, 2006", m[1])
		if err != nil {
			continue
		}
		columnDates[colNum] = t
		overview.Dates = append(overview.Dates, t)
	}

	// Extract team rows. Each team row has:
	//   <td class="name"><a href="...;teamId={teamId}">{TeamName}</a></td>
	//   followed by <td id="{teamId}_{colNum}" ...> cells, some with illegal="T"
	teamNameRe := regexp.MustCompile(`<td class="name"><a href="[^"]*;teamId=([^"]+)">([^<]+)</a></td>`)
	teamMatches := teamNameRe.FindAllStringSubmatch(html, -1)

	// Extract all illegal cells: <td id="{teamId}_{colNum}" ... illegal="T">
	illegalRe := regexp.MustCompile(`<td id="([^"]+)_(\d+)"[^>]*illegal="T"`)
	illegalMatches := illegalRe.FindAllStringSubmatch(html, -1)

	// Build a set of teamId -> illegal dates
	illegalMap := make(map[string][]time.Time)
	for _, m := range illegalMatches {
		teamID := m[1]
		colNum, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if date, ok := columnDates[colNum]; ok {
			illegalMap[teamID] = append(illegalMap[teamID], date)
		}
	}

	// Build team entries
	for _, m := range teamMatches {
		teamID := m[1]
		teamName := strings.TrimSpace(m[2])

		team := models.IllegalRosterTeam{
			TeamID:       teamID,
			TeamName:     teamName,
			IllegalDates: illegalMap[teamID],
		}
		overview.Teams = append(overview.Teams, team)
	}

	if len(overview.Teams) == 0 {
		return nil, fmt.Errorf("no teams found in illegal roster overview page")
	}

	return overview, nil
}
