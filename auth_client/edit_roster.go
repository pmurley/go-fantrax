package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pmurley/go-fantrax/models"
)

// Status ID constants
const (
	StatusActive  = "1" // Active roster
	StatusReserve = "2" // Reserve/Bench
	StatusIR      = "3" // Injured Reserve
	StatusMinors  = "9" // Minors
)

// Position ID constants - represent slot types, not individual slots
// Note: Multiple roster slots can share the same position ID
// Not all leagues will have all position slot types
const (
	PosC    = "001" // Catcher
	Pos1B   = "002" // First Base
	Pos3B   = "004" // Third Base
	PosSS   = "005" // Shortstop
	PosMI   = "007" // Middle Infield (2B or SS)
	PosCF   = "010" // Center Field
	PosOF   = "012" // Outfield
	PosUtil = "014" // Utility
	PosSP   = "015" // Starting Pitcher
	PosRP   = "016" // Relief Pitcher (league may have multiple RP slots with this ID)
	PosP    = "017" // Pitcher (any pitcher - SP or RP eligible, league may have multiple P slots)
	PosRP2  = "043" // Relief Pitcher 2
	PosRP3  = "044" // Relief Pitcher 3
)

// RosterPosition represents a player's position and status on the roster
type RosterPosition struct {
	PosID string `json:"posId,omitempty"` // Position slot ID (e.g., "001", "002"). Optional for some players.
	StID  string `json:"stId"`            // Status ID: "1"=Active, "2"=Reserve, "9"=Minors/IR
}

// ConfirmOrExecuteTeamRosterChangesRequest represents the request payload for roster changes
type ConfirmOrExecuteTeamRosterChangesRequest struct {
	RosterLimitPeriod    int                       `json:"rosterLimitPeriod"`
	FantasyTeamID        string                    `json:"fantasyTeamId"`
	Daily                bool                      `json:"daily"`
	AdminMode            bool                      `json:"adminMode"`
	ApplyToFuturePeriods bool                      `json:"applyToFuturePeriods"`
	FieldMap             map[string]RosterPosition `json:"fieldMap"` // Map of playerID -> RosterPosition
}

// ConfirmOrExecuteTeamRosterChangesRaw executes roster changes and returns the raw API response
//
// This method sends the complete roster state (all players with their positions and statuses).
// It does NOT send just the changes - you must provide the full roster in fieldMap.
//
// Parameters:
//   - period: The roster period (week number)
//   - teamID: The fantasy team ID to edit
//   - fieldMap: Map of ALL player IDs on the roster to their positions and statuses
//   - applyToFuturePeriods: true = apply to current and future periods, false = current period only
//   - daily: Whether this is a daily league
//   - adminMode: true = commissioner editing another team, false = user editing own team
//
// Status IDs (StID):
//   - "1" = Active Roster
//   - "2" = Reserve/Bench
//   - "9" = Minors/Injured Reserve
//
// Position IDs (PosID) are league-specific. Some players may not have a PosID (only StID).
//
// Returns the raw API response including all fields, or an error if the request failed.
func (c *Client) ConfirmOrExecuteTeamRosterChangesRaw(
	period int,
	teamID string,
	fieldMap map[string]RosterPosition,
	applyToFuturePeriods bool,
	daily bool,
	adminMode bool,
) (*models.RosterChangeResponse, error) {

	requestPayload := FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "confirmOrExecuteTeamRosterChanges",
				Data: ConfirmOrExecuteTeamRosterChangesRequest{
					RosterLimitPeriod:    period,
					FantasyTeamID:        teamID,
					Daily:                daily,
					AdminMode:            adminMode,
					ApplyToFuturePeriods: applyToFuturePeriods,
					FieldMap:             fieldMap,
				},
			},
		},
	}

	// Build the reference URL
	refUrl := fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/team/roster#league-team-roster-confirm-dialog", c.LeagueID)

	// Get timezone from UserInfo if available, otherwise default to UTC
	timezone := "UTC"
	if c.UserInfo != nil && c.UserInfo.Timezone != "" {
		timezone = c.UserInfo.Timezone
	}

	// Build the full request with metadata
	fullRequest := map[string]interface{}{
		"msgs":   requestPayload.Msgs,
		"uiv":    3,
		"refUrl": refUrl,
		"dt":     1,
		"at":     0,
		"av":     "0.0",
		"tz":     timezone,
		"v":      "172.1.0",
	}

	jsonStr, err := json.Marshal(fullRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
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
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse models.RosterChangeResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &apiResponse, nil
}

// ConfirmOrExecuteTeamRosterChanges executes roster changes and returns a simplified result
//
// This is a convenience wrapper around ConfirmOrExecuteTeamRosterChangesRaw that parses
// the response into an easier-to-use format and checks for errors.
//
// See ConfirmOrExecuteTeamRosterChangesRaw for detailed parameter documentation.
//
// Returns a RosterChangeResult with success status, changes made, and any error messages.
func (c *Client) ConfirmOrExecuteTeamRosterChanges(
	period int,
	teamID string,
	fieldMap map[string]RosterPosition,
	applyToFuturePeriods bool,
	daily bool,
	adminMode bool,
) (*models.RosterChangeResult, error) {

	rawResponse, err := c.ConfirmOrExecuteTeamRosterChangesRaw(period, teamID, fieldMap, applyToFuturePeriods, daily, adminMode)
	if err != nil {
		return nil, err
	}

	// Parse the response into a simplified result
	result := &models.RosterChangeResult{}

	if len(rawResponse.Responses) == 0 {
		return nil, fmt.Errorf("API returned empty responses array")
	}

	responseData := rawResponse.Responses[0].Data

	// Check if this was a commissioner action
	result.IsCommissioner = responseData.Commissioner

	// Check for main error message (critical failure)
	if responseData.FantasyResponse.MainMsg != "" {
		result.Success = false
		result.ErrorMessage = responseData.FantasyResponse.MainMsg
		return result, nil
	}

	// Check if the change was allowed
	if !responseData.TextArray.Model.ChangeAllowed {
		result.Success = false
		result.ErrorMessage = "Change not allowed by league rules"
		result.Warnings = responseData.TextArray.Model.IllegalRosterMsgs
		return result, nil
	}

	// Check for errors via showConfirmWindow
	if responseData.FantasyResponse.ShowConfirmWindow {
		result.Success = false
		result.ErrorMessage = "API indicated error via showConfirmWindow"
		result.Warnings = responseData.TextArray.Model.IllegalRosterMsgs
		return result, nil
	}

	// Success!
	result.Success = true
	result.Changes = responseData.TextArray.Model.RosterAdjustmentInfo.LineupChanges
	result.Warnings = responseData.TextArray.Model.IllegalRosterMsgs
	result.TotalFee = responseData.TextArray.Model.RosterAdjustmentInfo.TotalFee

	return result, nil
}

// BuildFieldMapFromRoster extracts a fieldMap from a TeamRosterResponse
//
// This helper function iterates through all tables and rows in the roster response
// and builds the fieldMap required for ConfirmOrExecuteTeamRosterChanges.
//
// Typical workflow:
//   1. roster := client.GetTeamRosterInfoRaw(period, teamID)
//   2. fieldMap := BuildFieldMapFromRoster(roster)
//   3. Modify fieldMap (e.g., fieldMap["playerId"].StID = "2")
//   4. client.ConfirmOrExecuteTeamRosterChanges(period, teamID, fieldMap, ...)
//
// Note: The roster response contains multiple tables (usually one for hitters, one for pitchers).
// This function processes all tables to build the complete roster.
func BuildFieldMapFromRoster(roster *models.TeamRosterResponse) map[string]RosterPosition {
	fieldMap := make(map[string]RosterPosition)

	if len(roster.Responses) == 0 {
		return fieldMap
	}

	// Iterate through all tables (hitters, pitchers, etc.)
	for _, table := range roster.Responses[0].Data.Tables {
		// Iterate through all rows (players)
		for _, row := range table.Rows {
			// Skip rows without a scorer (empty slots, section headers, etc.)
			if row.Scorer.ScorerID == "" {
				continue
			}

			playerID := row.Scorer.ScorerID

			// Build RosterPosition - some players may not have posId
			position := RosterPosition{
				StID: row.StatusID,
			}

			if row.PosID != "" {
				position.PosID = row.PosID
			}

			fieldMap[playerID] = position
		}
	}

	return fieldMap
}

// RosterEditor provides a high-level interface for editing team rosters
// It handles the complexity of fieldMaps, status IDs, and position IDs.
//
// Usage:
//
//	editor, err := client.NewRosterEditor(period, teamID, adminMode, daily)
//	editor.MoveToActive(playerID, auth_client.PosSS)
//	editor.MoveToReserve(playerID)
//	result, err := editor.Apply(applyToFuturePeriods)
type RosterEditor struct {
	client      *Client
	period      int
	teamID      string
	adminMode   bool
	daily       bool
	rawRoster   *models.TeamRosterResponse
	fieldMap    map[string]RosterPosition
	playerNames map[string]string // playerID -> name (for helpful error messages)
	changesMade []string          // track what we've changed for logging
}

// PlayerInfo represents basic information about a player on the roster
type PlayerInfo struct {
	PlayerID   string
	Name       string
	StatusID   string // "1"=Active, "2"=Reserve, "3"=IR, "9"=Minors
	PositionID string // "001", "005", etc. (may be empty for non-active players)
}

// NewRosterEditor creates a new roster editor for the specified team and period
//
// This method fetches the current roster state from the API.
//
// Parameters:
//   - period: The roster period (week number). Pass 0 to auto-detect the current period.
//   - teamID: The fantasy team ID to edit (empty string = authenticated user's team)
//   - adminMode: true = commissioner editing another team, false = user editing own team
//   - daily: true = daily league, false = weekly league
//
// Best practice: Create editor, make changes, and call Apply() immediately.
// Do not hold the editor for long periods as roster state may change externally.
func (c *Client) NewRosterEditor(period int, teamID string, adminMode bool, daily bool) (*RosterEditor, error) {
	// Auto-detect current period if 0 is passed
	if period == 0 {
		currentPeriod, err := c.GetCurrentPeriod()
		if err != nil {
			return nil, fmt.Errorf("failed to auto-detect current period: %w", err)
		}
		period = currentPeriod
	}

	// Fetch current roster
	rawRoster, err := c.GetTeamRosterInfoRaw(fmt.Sprintf("%d", period), teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current roster: %w", err)
	}

	// Build initial fieldMap from current state
	fieldMap := BuildFieldMapFromRoster(rawRoster)

	// Build playerNames map for helpful error messages
	playerNames := make(map[string]string)
	for _, table := range rawRoster.Responses[0].Data.Tables {
		for _, row := range table.Rows {
			if row.Scorer.ScorerID != "" {
				playerNames[row.Scorer.ScorerID] = row.Scorer.Name
			}
		}
	}

	return &RosterEditor{
		client:      c,
		period:      period,
		teamID:      teamID,
		adminMode:   adminMode,
		daily:       daily,
		rawRoster:   rawRoster,
		fieldMap:    fieldMap,
		playerNames: playerNames,
		changesMade: []string{},
	}, nil
}

// MoveToActive moves a player to the Active roster at the specified position
//
// This method works for both:
//   - Moving Reserve/Minors/IR players to Active
//   - Changing the position of already-Active players
//
// Parameters:
//   - playerID: The player's ID
//   - positionID: The position slot type (use constants like PosSS, PosC, etc.)
//
// Returns an error if the player is not found on the roster.
func (e *RosterEditor) MoveToActive(playerID string, positionID string) error {
	pos, exists := e.fieldMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found on roster", playerID)
	}

	oldStatus := pos.StID
	oldPos := pos.PosID

	pos.StID = StatusActive
	pos.PosID = positionID
	e.fieldMap[playerID] = pos

	playerName := e.playerNames[playerID]
	if oldStatus == StatusActive && oldPos != "" {
		e.changesMade = append(e.changesMade, fmt.Sprintf("%s: %s → %s", playerName, positionName(oldPos), positionName(positionID)))
	} else {
		e.changesMade = append(e.changesMade, fmt.Sprintf("%s: %s → Active at %s", playerName, statusName(oldStatus), positionName(positionID)))
	}

	return nil
}

// MoveToReserve moves a player to the Reserve/Bench
//
// The posId is automatically cleared to let Fantrax assign an appropriate position.
func (e *RosterEditor) MoveToReserve(playerID string) error {
	pos, exists := e.fieldMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found on roster", playerID)
	}

	oldStatus := pos.StID
	pos.StID = StatusReserve
	pos.PosID = "" // Clear posId - let Fantrax assign
	e.fieldMap[playerID] = pos

	playerName := e.playerNames[playerID]
	e.changesMade = append(e.changesMade, fmt.Sprintf("%s: %s → Reserve", playerName, statusName(oldStatus)))
	return nil
}

// MoveToMinors moves a player to the Minors
//
// The posId is automatically cleared to let Fantrax assign an appropriate position.
func (e *RosterEditor) MoveToMinors(playerID string) error {
	pos, exists := e.fieldMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found on roster", playerID)
	}

	oldStatus := pos.StID
	pos.StID = StatusMinors
	pos.PosID = "" // Clear posId - let Fantrax assign
	e.fieldMap[playerID] = pos

	playerName := e.playerNames[playerID]
	e.changesMade = append(e.changesMade, fmt.Sprintf("%s: %s → Minors", playerName, statusName(oldStatus)))
	return nil
}

// MoveToIR moves a player to Injured Reserve
//
// The posId is automatically cleared to let Fantrax assign an appropriate position.
func (e *RosterEditor) MoveToIR(playerID string) error {
	pos, exists := e.fieldMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found on roster", playerID)
	}

	oldStatus := pos.StID
	pos.StID = StatusIR
	pos.PosID = "" // Clear posId - let Fantrax assign
	e.fieldMap[playerID] = pos

	playerName := e.playerNames[playerID]
	e.changesMade = append(e.changesMade, fmt.Sprintf("%s: %s → IR", playerName, statusName(oldStatus)))
	return nil
}

// GetPlayersByStatus returns all players with a specific status
func (e *RosterEditor) GetPlayersByStatus(statusID string) []PlayerInfo {
	var players []PlayerInfo
	for playerID, pos := range e.fieldMap {
		if pos.StID == statusID {
			players = append(players, PlayerInfo{
				PlayerID:   playerID,
				Name:       e.playerNames[playerID],
				StatusID:   pos.StID,
				PositionID: pos.PosID,
			})
		}
	}
	return players
}

// GetActivePlayers returns all active roster players
func (e *RosterEditor) GetActivePlayers() []PlayerInfo {
	return e.GetPlayersByStatus(StatusActive)
}

// GetReservePlayers returns all reserve/bench players
func (e *RosterEditor) GetReservePlayers() []PlayerInfo {
	return e.GetPlayersByStatus(StatusReserve)
}

// GetMinorsPlayers returns all players in the minors
func (e *RosterEditor) GetMinorsPlayers() []PlayerInfo {
	return e.GetPlayersByStatus(StatusMinors)
}

// GetIRPlayers returns all players on injured reserve
func (e *RosterEditor) GetIRPlayers() []PlayerInfo {
	return e.GetPlayersByStatus(StatusIR)
}

// GetPendingChanges returns a list of changes that have been queued but not yet applied
func (e *RosterEditor) GetPendingChanges() []string {
	return e.changesMade
}

// Apply commits all changes to the Fantrax API
//
// Parameters:
//   - applyToFuturePeriods: true = apply to current and future periods, false = current period only
//
// Returns the result of the roster change operation, or an error if the request failed.
func (e *RosterEditor) Apply(applyToFuturePeriods bool) (*models.RosterChangeResult, error) {
	return e.client.ConfirmOrExecuteTeamRosterChanges(
		e.period,
		e.teamID,
		e.fieldMap,
		applyToFuturePeriods,
		e.daily,
		e.adminMode,
	)
}

// statusName converts a status ID to a human-readable name
func statusName(statusID string) string {
	switch statusID {
	case StatusActive:
		return "Active"
	case StatusReserve:
		return "Reserve"
	case StatusIR:
		return "IR"
	case StatusMinors:
		return "Minors"
	default:
		return fmt.Sprintf("Status(%s)", statusID)
	}
}

// positionName converts a position ID to a human-readable name
func positionName(positionID string) string {
	switch positionID {
	case PosC:
		return "C"
	case Pos1B:
		return "1B"
	case Pos3B:
		return "3B"
	case PosSS:
		return "SS"
	case PosMI:
		return "MI"
	case PosCF:
		return "CF"
	case PosOF:
		return "OF"
	case PosUtil:
		return "Util"
	case PosSP:
		return "SP"
	case PosRP:
		return "RP"
	case PosP:
		return "P"
	case PosRP2:
		return "RP2"
	case PosRP3:
		return "RP3"
	default:
		return fmt.Sprintf("Pos(%s)", positionID)
	}
}
