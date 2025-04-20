# Fantrax API Bindings for Go

This repo contains simple API bindings for the Fantrax Beta API for Golang.
It is current as of documentation provided from Fantrax in April 2025.

All credit to Fantrax for the development of the API. Please use these bindings responsibly.
Contributions/bug fixes are welcomed!

## Endpoints

Currently, there are bindings for the following endpoints:

### Retrieve Player IDs
This is used to retrieve the Fantrax IDs that identify every Player. All other API calls use these
IDs to refer to players.
- URL: `https://www.fantrax.com/fxea/general/getPlayerIds?sport=NFL`
- Request Parameters:
  - `sport` (required) - one of NFL, MLB, NHL, NBA, NCAAF, NCAAB, PGA, NASCAR, EPL

### Retrieve Player Info
This is used to retrieve ADP (Average Draft Pick) info for all players in the specified sport, with
optional filters.
- URL: `https://www.fantrax.com/fxea/general/getAdp`
- Request Parameters:
  - `sport` (required) - one of NFL, MLB, NHL, NBA, NCAAF, NCAAB, PGA, NASCAR, EPL
  - `position` (optional) - standard position abbreviations (e.g. QB, WR for football)
  - `showAllPositions` (optional) - whether to show default position, or all Fantrax positions of player. Can be
    “true” or “false”
  - `start` (optional) - start index for returned elements
  - `limit` (optional) - max number of returned elements
  - `order` (optional) - sort key for returned elements
- Request Body Example: `{"sport":"NFL","position":"QB","start":1,"limit":5,"order":"NAME"}`

### Retrieve League List
Retrieve the list of leagues, including name and ID of each league, as well as the name(s)
and ID(s) that the user owns in each league.
- URL: `https://www.fantrax.com/fxea/general/getLeagues`
- Request Parameters:
  - `userSecretId` (required) – the Secret ID shown on the Fantrax User Profile screen

### Retrieve League Info
Retrieve information about a specific league. This includes all the team names/IDs, matchups,
players in pool with info, and many of the league configuration settings.
- URL: `https://www.fantrax.com/fxea/general/getLeagueInfo?leagueId=[LeagueID]`
- Request Parameters: None

### Retrieve Draft Pick Info
Retrieve future and current draft picks in a specific league.
- URL: `https://www.fantrax.com/fxea/general/getDraftPicks?leagueId=[LeagueID]`
- Request Parameters: None

### Retrieve Draft Results
Retrieve the draft results of a specific league. This can be retrieved live during a draft.
- URL: `https://www.fantrax.com/fxea/general/getTeamRosters?leagueId=[LeagueID]&period=6`
- Request Parameters:
  - `period` (optional) - The lineup period for which the rosters are returned

### Retrieve League Standings
Retrieve the current standings of the league. This includes basic standings data, such as the
rank, points, W-L-T, games back, win %, etc. Individual stats are not yet included, but will be
in a future release.
- URL: `https://www.fantrax.com/fxea/general/getStandings?leagueId=[LeagueID]`
- Request Parameters: None


## Example Usage

```go
package main

import (
	fantrax "github.com/pmurley/go-fantrax"
	log "github.com/sirupsen/logrus"
)

const leagueId string = "my-league-id"

func main() {
	client, err := fantrax.NewClient(true)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.GetDraftResults(leagueId)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.GetLeagueInfo(leagueId)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.GetTeamRosters(leagueId, fantrax.WithPeriod(3))
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.GetPlayerIds(fantrax.MLB)
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.GetPlayerInfo(fantrax.MLB,
		fantrax.WithStart(1),
		fantrax.WithLimit(100),
		fantrax.WithPosition("SS"),
		fantrax.WithOrder("ADP"),
		fantrax.WithShowAllPositions(true),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```