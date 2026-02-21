package models

// LeagueSetupMatchups is the top-level result of parsing the league setup page.
// It contains all matchup data, team metadata, division structure, and form
// configuration needed to POST changes back to the league setup endpoint.
type LeagueSetupMatchups struct {
	Teams      []LeagueSetupTeam      // All teams with IDs, names, short names
	Divisions  []LeagueSetupDivision  // Division structure with team assignments
	Matchups   map[int][]MatchupPair  // Period number -> list of matchup pairs
	FormConfig LeagueSetupFormConfig  // All config values needed to POST back
}

// LeagueSetupTeam represents a team parsed from addTeam() JS calls on the
// league setup page. Teams with multiple owners will have multiple entries in
// the Owners slice.
type LeagueSetupTeam struct {
	TeamID    string
	Name      string
	ShortName string
	Owners    []TeamOwner
}

// TeamOwner represents a single owner of a team, parsed from addTeam() JS calls.
type TeamOwner struct {
	Email          string
	UserID         string // Original userId from addTeam(); "NULL" if owner hasn't joined
	IsCommissioner bool
	JoinedLeague   bool
}

// LeagueSetupDivision represents a division with its assigned teams, parsed
// from divisionName_ inputs and __removeTeamFromDivision() JS calls.
type LeagueSetupDivision struct {
	DivisionID string
	Name       string
	TeamIDs    []string
}

// MatchupPair represents a single away vs home matchup within a scoring period.
// A HomeTeamID of "-1" indicates a bye.
type MatchupPair struct {
	AwayTeamID string
	HomeTeamID string
}

// LeagueSetupFormConfig holds all the form field values from the league setup
// page that need to be echoed back unchanged when POSTing matchup changes.
type LeagueSetupFormConfig struct {
	// HiddenFields stores values from <input type="hidden"> elements (name -> value)
	HiddenFields map[string]string
	// SelectFields stores the selected value from <select> elements (name -> selected value)
	SelectFields map[string]string
	// CheckboxFields stores checkbox shadow fields prefixed with _ (name -> "on")
	CheckboxFields map[string]string
	// TeamNames maps teamId -> team name from teamName_{teamId} inputs
	TeamNames map[string]string
	// TeamShortNames maps teamId -> short name from teamShortName_{teamId} inputs
	TeamShortNames map[string]string
	// OwnerEmailFields stores the computed teamOwnerEmail form field keys and values.
	// Only owners where !IsCommissioner && !JoinedLeague generate email input fields.
	// Key format: "teamOwnerEmail,{email},{teamId},{userId}" -> email value.
	OwnerEmailFields map[string]string
	// DivisionNames maps divisionId -> division name for divisionName_{divId} POST fields.
	DivisionNames map[string]string
	// Divisions stores the ~~divisions values for POST reconstruction.
	// Each entry is one ~~divisions form field: "{divId}={teamId1}|{teamId2}|..."
	Divisions []string
}
