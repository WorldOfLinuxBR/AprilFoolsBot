package main

import (
	"os"
	"errors"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Env struct {
	MongoURI string
	DCToken string
	GuildID string
	OwnerID string
}
/*

*/
func init() {
	if err := godotenv.Load(); err != nil {
		log.Info("Error loading .env file")
	}
	log.Println(".env file loaded")
}

func GetEnv() (Env, error) {
	Envv := Env{
		MongoURI: os.Getenv("MONGODB_URI"),
		DCToken: os.Getenv("DISCORD_TOKEN"),
		GuildID: os.Getenv("GUILD_ID"),
		OwnerID: os.Getenv("OWNER_ID"),
	}

	if Envv.GuildID == "" || Envv.OwnerID == "" {
		return Envv, errors.New("You must set 'GUILD_ID' and 'OWNER_ID' environmental variables.")
	}

	if Envv.MongoURI == "" {
		return Envv, errors.New("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	if Envv.DCToken == "" {
		return Envv, errors.New("You must set your 'DISCORD_TOKEN' environmental variable.")
	}

	return Envv, nil
}

