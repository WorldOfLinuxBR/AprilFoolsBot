package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

var DB Database
var Envvars *Env
var gOwnerId string
var gGuildID string

func main() {
	var err error
	Envvars, err = GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	gOwnerId = Envvars.OwnerID
	gGuildID = Envvars.GuildID



	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := DB.Connect(ctx); err != nil {
		log.Fatal("Error connecting to the database.")
	}
	fmt.Println("Connected to MongoDB!")

	defer func() {
		DB.Disconnect(ctx)
		fmt.Println("\nDisconnected from MongoDB.")
	}()

	DB.SwitchTo("Discord", "Users")

	discord, err := discordgo.New("Bot " + Envvars.DCToken)
	if err != nil {
		log.Fatal("Discord: ", err)
	}

	discord.AddHandler(messageCreate)

	discord.Identify.Intents = discordgo.IntentsAll

	if err := discord.Open(); err != nil {
		log.Fatal("Discord: ", err)
	}

	fmt.Println("Logged in as", discord.State.User.String(), "Press CTRL + C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!aprilfools" && m.Author.ID == gOwnerId {
		s.ChannelMessageSend(m.ChannelID, "A pegadinha vai começar alek. Vê o terminal.")

		members, err := s.GuildMembers(gGuildID, "", 1000)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Lista de usuários: %s", err.Error()))
		}

		for _, v := range members {
			name := v.Nick

			if name == "" {
				name = v.User.Username
			}

			oldname := name

			// Check if the user's name already contains "@(Whatever the user has set)"
			if strings.Contains(name, "@" + Envvars.AppendName) {
				continue
			}

			// Check if the user's name contains "@"
			if strings.Contains(name, "@") {
				name = strings.Split(name, "@")[0] + "@" + Envvars.AppendName
			} else {
				name = name + "@" + Envvars.AppendName
			}

			// Update the user's nickname
			err := s.GuildMemberNickname(gGuildID, v.User.ID, name)
			if err != nil {
				log.Error(err)
			}

			log.Infof("%s -> %s", oldname, name)
		}

		s.ChannelMessageSend(m.ChannelID, "KKKKKKKKKKKKKKKKKKKKKKKK MUDEI FOI (QUAIS) TUDO! :wedoalittletrolling:")
	}

	if m.Content == "!backupUsernames" && m.Author.ID == gOwnerId {
		res, err := DB.Collection.DeleteMany(context.TODO(), bson.M{})
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error deleting collection data: %v", err))
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d Documentos deletados.", res.DeletedCount))

		msg, _ := s.ChannelMessageSend(m.ChannelID, ":warning: Status: 0/?")

		members, err := s.GuildMembers(gGuildID, "", 1000)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error getting users: %s", err.Error()))
		}

		for i, v := range members {
			name := v.User.Username
			if v.Nick != "" {
				name = v.Nick

			}
			_, err := DB.Collection.InsertOne(context.TODO(), bson.M{
				"uid":      v.User.ID,
				"username": name,
			})

			if err != nil {
				log.Error(err)
			}
			s.ChannelMessageEdit(msg.ChannelID, msg.ID, fmt.Sprintf(":warning: Status: %d/%d", i+1, len(members)))
		}

	}

	if m.Content == "!undoAprilFools" && m.Author.ID == gOwnerId {
		// Restore usernames from the database
		cur, err := DB.Collection.Find(context.Background(), bson.M{})
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error finding collection data: %v", err))
			return
		}
		defer cur.Close(context.Background())

		var count int
		msg, _ := s.ChannelMessageSend(m.ChannelID, ":warning: **status**: 0 usernames restaurados.")
		for cur.Next(context.Background()) {
			s.ChannelMessageEdit(msg.ChannelID, msg.ID, fmt.Sprintf(":warning: **status**: %d usernames restaurados.", count))
			var user struct {
				UID      string `bson:"uid"`
				Username string `bson:"username"`
			}
			if err := cur.Decode(&user); err != nil {
				log.Error(err)
				continue
			}

			member, err := s.GuildMember(gGuildID, user.UID)
			if err != nil {
				log.Error(err)
				continue
			}

			// Restore the username
			if member.Nick != user.Username || member.User.Username != user.Username {
				if err := s.GuildMemberNickname(gGuildID, user.UID, user.Username); err != nil {
					log.Error(err)
					continue
				}
				count++

			}

		}

		if err := cur.Err(); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error iterating collection data: %v", err))
			return
		}
	}
}
