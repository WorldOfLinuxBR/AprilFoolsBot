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
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Info("Error loading .env file")
	}
	log.Println(".env file loaded")
}

var gOwnerId string
var gGuildID string
var coll *mongo.Collection

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	dcToken := os.Getenv("DISCORD_TOKEN")
	GuildID := os.Getenv("GUILD_ID")
	OwnerID := os.Getenv("OWNER_ID")

	if GuildID == "" || OwnerID == "" {
		log.Fatal("You must set 'GUILD_ID' and 'OWNER_ID' environmental variables.")
	}

	gOwnerId = OwnerID
	gGuildID = GuildID

	if mongoURI == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	if dcToken == "" {
		log.Fatal("You must set your 'DISCORD_TOKEN' environmental variable.")
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	MongoDB, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPIOptions))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	defer func() {
		if err := MongoDB.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("\nDisconnected from MongoDB.")
	}()

	coll = MongoDB.Database("Discord").Collection("Users")

	discord, err := discordgo.New("Bot " + dcToken)
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

			// Check if the user's name already contains "@Windows 12"
			if strings.Contains(name, "@Windows 12") {
				continue
			}

			// Check if the user's name contains "@"
			if strings.Contains(name, "@") {
				name = strings.Split(name, "@")[0] + "@Windows 12"
			} else {
				name = name + "@Windows 12"
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
		res, err := coll.DeleteMany(context.TODO(), bson.M{})
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
			_, err := coll.InsertOne(context.TODO(), bson.M{
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
		cur, err := coll.Find(context.Background(), bson.M{})
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error finding collection data: %v", err))
			return
		}
		defer cur.Close(context.Background())

		var count int
        msg, _ := s.ChannelMessageSend(m.ChannelID, ":warning: **status**: 0 usernames restaurados.")
		for cur.Next(context.Background()) {
      s.ChannelMessageEdit(msg.ChannelID,msg.ID, fmt.Sprintf(":warning: **status**: %d usernames restaurados.", count))
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
