package database

import (
	"github.com/pmurley/go-fantrax"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	Players *map[string]fantrax.Player `json:"players"`
	Teams   *fantrax.LeagueRosters     `json:"teams"`
}

func NewDatabase() *Database {
	return &Database{
		Players: new(map[string]fantrax.Player),
	}
}

func (db *Database) Update(c *fantrax.Client) {
	playerIDs, err := c.GetPlayerIds(fantrax.MLB)
	if err != nil {
		log.Fatal(err)
	}
	db.Players = playerIDs

	teamRosters, err := c.GetTeamRosters()
	if err != nil {
		log.Fatal(err)
	}
	db.Teams = teamRosters
}
