package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// StandingsResponse represents the top-level response from the Fantrax API
type StandingsResponse struct {
	Data      ResponseMetadata `json:"data"`
	Roles     []string         `json:"roles"`
	Responses []Response       `json:"responses"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	SDate int64  `json:"sDate"`
	Adrt  int    `json:"adrt"`
	Up    string `json:"up"`
}

// Response represents a single response from the API
type Response struct {
	Data ResponseData `json:"data"`
}

// ResponseData contains all the standings and team data
type ResponseData struct {
	GoBackDays          []int                  `json:"goBackDays"`
	FantasyTeamInfo     map[string]FantasyTeam `json:"fantasyTeamInfo"`
	DisplayedSelections DisplayedSelections    `json:"displayedSelections"`
	MiscData            MiscData               `json:"miscData"`
	TableList           []Table                `json:"tableList"`
	DisplayedLists      DisplayedLists         `json:"displayedLists"`
}

// FantasyTeam represents information about a single team
type FantasyTeam struct {
	Name       string `json:"name"`
	LogoURL512 string `json:"logoUrl512"`
	ShortName  string `json:"shortName"`
}

// DisplayedSelections represents user selected display options
type DisplayedSelections struct {
	ProjectionsAvailable bool   `json:"projectionsAvailable"`
	Period               int    `json:"period"`
	TimeStartType        string `json:"timeStartType"`
	View                 string `json:"view"`
	ShowTabs             bool   `json:"showTabs"`
	HideGoBackDays       bool   `json:"hideGoBackDays"`
	TimeframeType        string `json:"timeframeType"`
	Proj                 bool   `json:"proj"`
	DisplayedStartDate   int64  `json:"displayedStartDate"`
	DisplayedEndDate     int64  `json:"displayedEndDate"`
}

// MiscData contains additional display metadata
type MiscData struct {
	DisplayedMinDate int64  `json:"displayedMinDate"`
	ShowLogos        bool   `json:"showLogos"`
	Heading          string `json:"heading"`
	DisplayedMaxDate int64  `json:"displayedMaxDate"`
}

// Table represents a single table in the response (standings or matchups)
type Table struct {
	FixedRows   bool        `json:"fixedRows"`
	TableType   string      `json:"tableType"`
	FixedHeader *HeaderData `json:"fixedHeader,omitempty"`
	Caption     string      `json:"caption"`
	SubCaption  string      `json:"subCaption"`
	Header      HeaderData  `json:"header"`
	Rows        []Row       `json:"rows"`
}

// HeaderData represents table header information
type HeaderData struct {
	Cells []Cell `json:"cells"`
}

// Row represents a row in a table
type Row struct {
	Cells      []Cell `json:"cells"`
	FixedCells []Cell `json:"fixedCells,omitempty"`
	Highlight  bool   `json:"highlight,omitempty"`
}

// Cell represents a single cell in a table
type Cell struct {
	Content       string `json:"content,omitempty"`
	ToolTip       string `json:"toolTip,omitempty"`
	Align         string `json:"align,omitempty"`
	SortDirection int    `json:"sortDirection,omitempty"`
	Name          string `json:"name,omitempty"`
	ShortName     string `json:"shortName,omitempty"`
	Key           string `json:"key,omitempty"`
	LeagueID      string `json:"leagueId,omitempty"`
	ID            string `json:"id,omitempty"`
	TeamID        string `json:"teamId,omitempty"`
}

// DisplayedLists contains various display configuration lists
type DisplayedLists struct {
	GoBackDays     []int           `json:"goBackDays"`
	Pagination     Pagination      `json:"pagination"`
	Tabs           []Tab           `json:"tabs"`
	Periods        []Period        `json:"periods"`
	TimeframeTypes []TimeframeType `json:"timeframeTypes"`
	TimeStartTypes []TimeStartType `json:"timeStartTypes"`
}

// Pagination contains pagination information
type Pagination struct {
	StartPageNum    int `json:"startPageNum"`
	NumTeamsPerPage int `json:"numTeamsPerPage"`
	EndPageNum      int `json:"endPageNum"`
	PageNum         int `json:"pageNum"`
}

// Tab represents a navigation tab
type Tab struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Period represents a scoring period
type Period struct {
	Object1 int    `json:"object1"`
	Object2 string `json:"object2"`
}

// TimeframeType represents a time frame display option
type TimeframeType struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// TimeStartType represents a time start display option
type TimeStartType struct {
	Object1 string `json:"object1"`
	Object2 string `json:"object2"`
}

// StandingsTeam represents processed standings data for a single team
type StandingsTeam struct {
	Rank          int
	TeamName      string
	TeamID        string
	Wins          int
	Losses        int
	Ties          int
	WinPct        float64
	DivRecord     string
	GamesBack     float64
	WaiverOrder   int
	PointsFor     float64
	PointsAgainst float64
	Streak        string
}

// LeagueStandings represents the processed standings data in an intuitive format
type LeagueStandings struct {
	LeagueName  string         `json:"leagueName"`
	Teams       []TeamStanding `json:"teams"`
	Divisions   []Division     `json:"divisions"`
	Matchups    []Matchup      `json:"matchups"`
	SeasonDates DateRange      `json:"seasonDates"`
}

// TeamStanding represents a single team's standing information
type TeamStanding struct {
	TeamID        string  `json:"teamId"`
	Name          string  `json:"name"`
	ShortName     string  `json:"shortName"`
	LogoURL       string  `json:"logoUrl"`
	Rank          int     `json:"rank"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Ties          int     `json:"ties"`
	WinPct        float64 `json:"winPct"`
	DivRecord     string  `json:"divRecord"`
	GamesBack     float64 `json:"gamesBack"`
	WaiverOrder   int     `json:"waiverOrder"`
	PointsFor     float64 `json:"pointsFor"`
	PointsAgainst float64 `json:"pointsAgainst"`
	Streak        string  `json:"streak"`
	DivisionID    string  `json:"divisionId,omitempty"`
}

// Division represents a division in the league
type Division struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Teams []string `json:"teamIds"` // TeamIDs in this division
}

// Matchup represents a single matchup between two teams
type Matchup struct {
	ScoringPeriod int       `json:"scoringPeriod"`
	Date          string    `json:"date"`
	AwayTeam      MatchTeam `json:"awayTeam"`
	HomeTeam      MatchTeam `json:"homeTeam"`
}

// MatchTeam represents a team in a matchup with score
type MatchTeam struct {
	TeamID     string  `json:"teamId"`
	Points     float64 `json:"points"`
	Adjustment float64 `json:"adjustment"`
	Total      float64 `json:"total"`
}

// DateRange represents a time period with start and end dates
type DateRange struct {
	StartDate int64 `json:"startDate"`
	EndDate   int64 `json:"endDate"`
}

// ProcessStandings converts the raw API response into a more intuitive structure
func ProcessStandings(response *StandingsResponse) (*LeagueStandings, error) {
	if len(response.Responses) == 0 {
		return nil, fmt.Errorf("no response data found")
	}

	responseData := response.Responses[0].Data

	// Create a new LeagueStandings
	standings := &LeagueStandings{
		LeagueName: responseData.MiscData.Heading,
		Teams:      make([]TeamStanding, 0),
		Divisions:  make([]Division, 0),
		Matchups:   make([]Matchup, 0),
		SeasonDates: DateRange{
			StartDate: responseData.MiscData.DisplayedMinDate,
			EndDate:   responseData.MiscData.DisplayedMaxDate,
		},
	}

	// Process divisions from tabs
	divisionMap := make(map[string]Division)
	for _, tab := range responseData.DisplayedLists.Tabs {
		// Skip non-division tabs
		if tab.ID == "ALL" || tab.ID == "COMBINED" || tab.ID == "SCHEDULE" ||
			tab.ID == "SEASON_STATS" || tab.ID == "PLAYOFFS" {
			continue
		}

		divisionMap[tab.ID] = Division{
			ID:    tab.ID,
			Name:  tab.Name,
			Teams: make([]string, 0),
		}
	}

	// Process teams and standings table
	for _, table := range responseData.TableList {
		if table.TableType == "H2hPointsBased1" {
			// This is the standings table
			for _, row := range table.Rows {
				if len(row.Cells) < 10 || len(row.FixedCells) < 2 {
					continue
				}

				teamCell := row.FixedCells[1]
				teamID := teamCell.TeamID
				if teamID == "" {
					continue
				}

				teamInfo := responseData.FantasyTeamInfo[teamID]

				rank, _ := strconv.Atoi(row.FixedCells[0].Content)
				wins, _ := strconv.Atoi(row.Cells[0].Content)
				losses, _ := strconv.Atoi(row.Cells[1].Content)
				ties, _ := strconv.Atoi(row.Cells[2].Content)
				winPct, _ := strconv.ParseFloat(row.Cells[3].Content, 64)
				gamesBack, _ := strconv.ParseFloat(row.Cells[5].Content, 64)
				waiverOrder, _ := strconv.Atoi(row.Cells[6].Content)
				pointsFor, _ := strconv.ParseFloat(row.Cells[7].Content, 64)
				pointsAgainst, _ := strconv.ParseFloat(row.Cells[8].Content, 64)

				team := TeamStanding{
					TeamID:        teamID,
					Name:          teamInfo.Name,
					ShortName:     teamInfo.ShortName,
					LogoURL:       teamInfo.LogoURL512,
					Rank:          rank,
					Wins:          wins,
					Losses:        losses,
					Ties:          ties,
					WinPct:        winPct,
					DivRecord:     row.Cells[4].Content,
					GamesBack:     gamesBack,
					WaiverOrder:   waiverOrder,
					PointsFor:     pointsFor,
					PointsAgainst: pointsAgainst,
					Streak:        row.Cells[9].Content,
				}

				standings.Teams = append(standings.Teams, team)
			}
		} else if table.TableType == "H2hPointsBased2" {
			// These are the matchup tables
			period := 0
			date := ""

			// Parse period number and date from caption
			if strings.HasPrefix(table.Caption, "Scoring Period ") {
				parts := strings.Split(table.Caption, " ")
				if len(parts) >= 3 {
					period, _ = strconv.Atoi(parts[2])
				}
			}

			// Extract date from subCaption (e.g., "(Sat Apr 19, 2025)")
			if len(table.SubCaption) > 2 {
				date = strings.Trim(table.SubCaption, "()")
			}

			for _, row := range table.Rows {
				if len(row.Cells) < 8 {
					continue
				}

				awayTeamID := row.Cells[0].TeamID
				awayPoints, _ := strconv.ParseFloat(row.Cells[1].Content, 64)
				awayAdj, _ := strconv.ParseFloat(row.Cells[2].Content, 64)
				awayTotal, _ := strconv.ParseFloat(row.Cells[3].Content, 64)

				homeTeamID := row.Cells[4].TeamID
				homePoints, _ := strconv.ParseFloat(row.Cells[5].Content, 64)
				homeAdj, _ := strconv.ParseFloat(row.Cells[6].Content, 64)
				homeTotal, _ := strconv.ParseFloat(row.Cells[7].Content, 64)

				matchup := Matchup{
					ScoringPeriod: period,
					Date:          date,
					AwayTeam: MatchTeam{
						TeamID:     awayTeamID,
						Points:     awayPoints,
						Adjustment: awayAdj,
						Total:      awayTotal,
					},
					HomeTeam: MatchTeam{
						TeamID:     homeTeamID,
						Points:     homePoints,
						Adjustment: homeAdj,
						Total:      homeTotal,
					},
				}

				standings.Matchups = append(standings.Matchups, matchup)
			}
		}
	}

	// Add divisions from the map to the result
	for _, div := range divisionMap {
		standings.Divisions = append(standings.Divisions, div)
	}

	return standings, nil
}

func (c *Client) GetStandings() (*LeagueStandings, error) {
	var requestPayload = FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getStandings",
				Data: map[string]string{
					"leagueId": c.LeagueID,
				},
			},
		},
	}

	jsonStr, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://www.fantrax.com/fxpa/req?leagueId="+c.LeagueID, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var response StandingsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	standings, err := ProcessStandings(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to process standings: %w", err)
	}

	return standings, nil
}
