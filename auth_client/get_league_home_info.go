package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ============================================================
// Raw API Response Types
// ============================================================

// LeagueHomeInfoRawResponse represents the top-level response from getLeagueHomeInfo
type LeagueHomeInfoRawResponse struct {
	Responses []LeagueHomeInfoRawResponseItem `json:"responses"`
}

// LeagueHomeInfoRawResponseItem represents a single response item
type LeagueHomeInfoRawResponseItem struct {
	Data LeagueHomeInfoRawData `json:"data"`
}

// LeagueHomeInfoRawData contains all the data from the response
type LeagueHomeInfoRawData struct {
	Settings            LeagueHomeInfoRawSettings            `json:"settings"`
	FantasyTeams        []LeagueHomeInfoRawTeam              `json:"fantasyTeams"`
	Standings           LeagueHomeInfoRawStandings           `json:"standings"`
	Matchups            LeagueHomeInfoRawMatchups            `json:"matchups"`
	PendingTransactions LeagueHomeInfoRawPendingTransactions `json:"pendingTransactions"`
}

// LeagueHomeInfoRawPendingTransactions contains the pending-trade envelope
// embedded in the league home info response.
type LeagueHomeInfoRawPendingTransactions struct {
	Sets      []LeagueHomeInfoRawPendingSet      `json:"pendingTransactionSets"`
	ScorerMap map[string]LeagueHomeInfoRawScorer `json:"scorerMap"`
}

// LeagueHomeInfoRawPendingSet groups one or more pending player moves under a
// single trade ID. Each set typically corresponds to one proposed trade.
type LeagueHomeInfoRawPendingSet struct {
	ID           string                       `json:"id"`
	Transactions []LeagueHomeInfoRawPendingTx `json:"transactions"`
}

// LeagueHomeInfoRawPendingTx is a single player move within a pending trade.
type LeagueHomeInfoRawPendingTx struct {
	ScorerID     string `json:"scorerId"`
	SourceTeamID string `json:"sourceTeamId"`
	DestTeamID   string `json:"destinationTeamId"`
}

// LeagueHomeInfoRawScorer is the per-player metadata that the pending trade
// payload references via ScorerID.
type LeagueHomeInfoRawScorer struct {
	Name          string `json:"name"`
	PosShortNames string `json:"posShortNames"`
}

// LeagueHomeInfoRawSettings contains league settings
type LeagueHomeInfoRawSettings struct {
	LeagueName        string `json:"leagueName"`
	SportID           string `json:"sportId"`
	PremiumLeagueType string `json:"premiumLeagueType"`
	LeagueDisplayYear string `json:"leagueDisplayYear"`
	LogoURL           string `json:"logoUrl"`
	LogoUploaded      bool   `json:"logoUploaded"`
}

// LeagueHomeInfoRawTeam contains fantasy team info
type LeagueHomeInfoRawTeam struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ShortName    string `json:"shortName"`
	Commissioner bool   `json:"commissioner"`
	LogoID       string `json:"logoId"`
	LogoURL128   string `json:"logoUrl128"`
	LogoURL256   string `json:"logoUrl256"`
}

// LeagueHomeInfoRawStandings contains standings data
type LeagueHomeInfoRawStandings struct {
	Header     []LeagueHomeInfoRawStandingsHeader           `json:"header"`
	StatsTable []map[string][]LeagueHomeInfoRawStandingsRow `json:"statsTable"`
}

// LeagueHomeInfoRawStandingsHeader contains header column definitions
type LeagueHomeInfoRawStandingsHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LeagueHomeInfoRawStandingsRow contains a single team's standings row
type LeagueHomeInfoRawStandingsRow struct {
	TeamID        string `json:"teamId"`
	Team          string `json:"team"`
	Rank          int    `json:"rank"`
	Score         string `json:"score"`
	WinPercentage string `json:"winPercentage"`
	GamesBack     string `json:"gamesBack"`
	Points        string `json:"points"`
	Commish       bool   `json:"commish,omitempty"`
}

// LeagueHomeInfoRawMatchups contains matchup data
type LeagueHomeInfoRawMatchups struct {
	TitlePeriodInfo string                  `json:"titlePeriodInfo"`
	Games           []LeagueHomeInfoRawGame `json:"games"`
	NoMatchupsMsg   string                  `json:"noMatchupsMsg"`
	Live            bool                    `json:"live"`
}

// LeagueHomeInfoRawGame contains a single matchup game
type LeagueHomeInfoRawGame struct {
	AwayTeamID    string `json:"awayTeamId"`
	AwayTeamName  string `json:"awayTeamName"`
	AwayTeamScore string `json:"awayTeamScore"`
	HomeTeamID    string `json:"homeTeamId"`
	HomeTeamName  string `json:"homeTeamName"`
	HomeTeamScore string `json:"homeTeamScore"`
}

// ============================================================
// Processed Types (clean, easy to use)
// ============================================================

// LeagueHomeInfo represents the processed league home info
type LeagueHomeInfo struct {
	Settings      LeagueSettings      `json:"settings"`
	Teams         []LeagueTeam        `json:"teams"`
	Standings     []DivisionStandings `json:"standings"`
	Matchups      LeagueMatchups      `json:"matchups"`
	PendingTrades []PendingTrade      `json:"pendingTrades,omitempty"`
}

// PendingTrade is a single player move within a pending (proposed but not yet
// accepted) trade. Multiple PendingTrade entries with the same TradeID make up
// one proposed trade.
type PendingTrade struct {
	// TradeID groups player moves that belong to the same proposed trade.
	TradeID string `json:"tradeId"`
	// PlayerName is the display name of the player being moved.
	PlayerName string `json:"playerName"`
	// Position is the player's position(s) as displayed by Fantrax (e.g.
	// "SP", "3B,INF,OF").
	Position string `json:"position"`
	// FromTeam is the resolved fantasy team name the player is moving from.
	// Falls back to the source team ID if the team is not in the team list.
	FromTeam string `json:"fromTeam"`
	// ToTeam is the resolved fantasy team name the player is moving to.
	// Falls back to the destination team ID if the team is not in the team list.
	ToTeam string `json:"toTeam"`
}

// LeagueSettings contains league configuration
type LeagueSettings struct {
	LeagueName        string `json:"leagueName"`
	SportID           string `json:"sportId"`
	PremiumLeagueType string `json:"premiumLeagueType"`
	Year              string `json:"year"`
	LogoURL           string `json:"logoUrl"`
	LogoUploaded      bool   `json:"logoUploaded"`
}

// LeagueTeam contains fantasy team info
type LeagueTeam struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ShortName    string `json:"shortName"`
	Commissioner bool   `json:"commissioner"`
	LogoID       string `json:"logoId,omitempty"`
	LogoURL128   string `json:"logoUrl128"`
	LogoURL256   string `json:"logoUrl256"`
}

// DivisionStandings contains standings for a single division
type DivisionStandings struct {
	DivisionName string            `json:"divisionName"`
	Teams        []TeamStandingRow `json:"teams"`
}

// TeamStandingRow contains a single team's standings info
type TeamStandingRow struct {
	TeamID        string `json:"teamId"`
	TeamName      string `json:"teamName"`
	Rank          int    `json:"rank"`
	Record        string `json:"record"`
	WinPercentage string `json:"winPercentage"`
	GamesBack     string `json:"gamesBack"`
	Points        string `json:"points"`
	Commissioner  bool   `json:"commissioner"`
}

// LeagueMatchups contains matchup info for the current period
type LeagueMatchups struct {
	PeriodInfo    string        `json:"periodInfo"`
	Games         []MatchupGame `json:"games"`
	NoMatchupsMsg string        `json:"noMatchupsMsg,omitempty"`
	Live          bool          `json:"live"`
}

// MatchupGame contains a single matchup game
type MatchupGame struct {
	AwayTeamID    string `json:"awayTeamId"`
	AwayTeamName  string `json:"awayTeamName"`
	AwayTeamScore string `json:"awayTeamScore"`
	HomeTeamID    string `json:"homeTeamId"`
	HomeTeamName  string `json:"homeTeamName"`
	HomeTeamScore string `json:"homeTeamScore"`
}

// ============================================================
// API Functions
// ============================================================

// GetLeagueHomeInfoRaw fetches the raw league home info response
func (c *Client) GetLeagueHomeInfoRaw() ([]byte, error) {
	requestPayload := FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getLeagueHomeInfo",
				Data:   map[string]interface{}{},
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// GetLeagueHomeInfo fetches and processes the league home info
func (c *Client) GetLeagueHomeInfo() (*LeagueHomeInfo, error) {
	rawBody, err := c.GetLeagueHomeInfoRaw()
	if err != nil {
		return nil, err
	}

	var rawResponse LeagueHomeInfoRawResponse
	if err := json.Unmarshal(rawBody, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return processLeagueHomeInfo(&rawResponse)
}

// processLeagueHomeInfo converts the raw response to the processed format
func processLeagueHomeInfo(raw *LeagueHomeInfoRawResponse) (*LeagueHomeInfo, error) {
	if len(raw.Responses) == 0 {
		return nil, fmt.Errorf("no response data found")
	}

	data := raw.Responses[0].Data

	result := &LeagueHomeInfo{
		Settings: LeagueSettings{
			LeagueName:        data.Settings.LeagueName,
			SportID:           data.Settings.SportID,
			PremiumLeagueType: data.Settings.PremiumLeagueType,
			Year:              data.Settings.LeagueDisplayYear,
			LogoURL:           data.Settings.LogoURL,
			LogoUploaded:      data.Settings.LogoUploaded,
		},
		Teams:     make([]LeagueTeam, 0, len(data.FantasyTeams)),
		Standings: make([]DivisionStandings, 0),
		Matchups: LeagueMatchups{
			PeriodInfo:    data.Matchups.TitlePeriodInfo,
			NoMatchupsMsg: data.Matchups.NoMatchupsMsg,
			Live:          data.Matchups.Live,
			Games:         make([]MatchupGame, 0, len(data.Matchups.Games)),
		},
	}

	// Process teams
	for _, team := range data.FantasyTeams {
		result.Teams = append(result.Teams, LeagueTeam{
			ID:           team.ID,
			Name:         team.Name,
			ShortName:    team.ShortName,
			Commissioner: team.Commissioner,
			LogoID:       team.LogoID,
			LogoURL128:   team.LogoURL128,
			LogoURL256:   team.LogoURL256,
		})
	}

	// Process standings by division
	for _, divisionMap := range data.Standings.StatsTable {
		for divisionName, rows := range divisionMap {
			division := DivisionStandings{
				DivisionName: divisionName,
				Teams:        make([]TeamStandingRow, 0, len(rows)),
			}
			for _, row := range rows {
				division.Teams = append(division.Teams, TeamStandingRow{
					TeamID:        row.TeamID,
					TeamName:      row.Team,
					Rank:          row.Rank,
					Record:        row.Score,
					WinPercentage: row.WinPercentage,
					GamesBack:     row.GamesBack,
					Points:        row.Points,
					Commissioner:  row.Commish,
				})
			}
			result.Standings = append(result.Standings, division)
		}
	}

	// Process matchups
	for _, game := range data.Matchups.Games {
		result.Matchups.Games = append(result.Matchups.Games, MatchupGame{
			AwayTeamID:    game.AwayTeamID,
			AwayTeamName:  game.AwayTeamName,
			AwayTeamScore: game.AwayTeamScore,
			HomeTeamID:    game.HomeTeamID,
			HomeTeamName:  game.HomeTeamName,
			HomeTeamScore: game.HomeTeamScore,
		})
	}

	// Process pending trades. Each PendingTransactionSet groups one or more
	// player moves under a single trade ID; flatten into a slice keyed by
	// TradeID so callers can group or list as needed. Resolve fantasy-team
	// names through the FantasyTeams list; fall back to the raw team ID if
	// the team isn't found (defensive — should always resolve).
	if len(data.PendingTransactions.Sets) > 0 {
		teamName := func(id string) string {
			for _, ft := range data.FantasyTeams {
				if ft.ID == id {
					return ft.Name
				}
			}
			return id
		}
		for _, set := range data.PendingTransactions.Sets {
			for _, tx := range set.Transactions {
				scorer := data.PendingTransactions.ScorerMap[tx.ScorerID]
				result.PendingTrades = append(result.PendingTrades, PendingTrade{
					TradeID:    set.ID,
					PlayerName: scorer.Name,
					Position:   scorer.PosShortNames,
					FromTeam:   teamName(tx.SourceTeamID),
					ToTeam:     teamName(tx.DestTeamID),
				})
			}
		}
	}

	return result, nil
}

// GetPendingTrades fetches the league home info and returns just the pending
// trades. Pending trades are proposed but not yet accepted moves; multiple
// PendingTrade entries with the same TradeID make up one proposed trade. For
// callers who want the full LeagueHomeInfo, use GetLeagueHomeInfo instead.
func (c *Client) GetPendingTrades() ([]PendingTrade, error) {
	info, err := c.GetLeagueHomeInfo()
	if err != nil {
		return nil, err
	}
	return info.PendingTrades, nil
}
