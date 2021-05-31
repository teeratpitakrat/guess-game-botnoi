package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var guessMap map[string]string
var re *regexp.Regexp
var token string
var channelId string

func main() {
	guessMap = make(map[string]string)
	re = regexp.MustCompile(`^\d{2}$`)

	token = os.Getenv("DISCORD_TOKEN")
	channelId = os.Getenv("DISCORD_CHANNEL_ID")

	log.Println("Creating a new discord connector")
	log.Printf("Token: %s\n", token)
	log.Printf("ChannelID: %s\n", channelId)

	if token == "" || channelId == "" {
		log.Println("Token or ChannelID cannot be empty")
		return
	}

	// Create a new Discord session using the provided bot token
	discordSession, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discordSession.AddHandler(messageCreate)

	// Open the websocket and begin listening
	err = discordSession.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
		return
	}

	discordSession.ChannelMessageSend(channelId, "เริ่มทายได้เลยค่า ใครทายซ้ำให้พิมพ์ใหม่ อย่า edit ข้อความเก่านะคะ เดี๋ยวบอทงง")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discordSession.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Process only messages in the specified channel
	if m.ChannelID == channelId {
		// If the message is "ping" reply with "Pong!"
		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
		}

		for _, guess := range re.FindAllString(m.Content, -1) {
			log.Println(fmt.Sprintf("%s guessed %s", m.Author.ID, guess))
			// Check for duplicate
			if _, exists := guessMap[guess]; exists {
				firstGuesser := guessMap[guess]
				if firstGuesser == m.Author.ID {
					reply_msg := fmt.Sprintf("<@%s> รู้แล้วค่ะว่าทายเลข %s ทายรอบเดียวก็พอค่ะ", m.Author.ID, guess)
					log.Println(reply_msg)
					s.ChannelMessageSend(m.ChannelID, reply_msg)
				} else {
					reply_msg := fmt.Sprintf("<@%s> ทายเลข %s ซ้ำกับ <@%s> ค่า", m.Author.ID, guess, firstGuesser)
					log.Println(reply_msg)
					s.ChannelMessageSend(m.ChannelID, reply_msg)
				}
			} else {
				// Check if m.Author.ID makes a new guess, if yes, delete old guess
				for guess, author := range guessMap {
					if author == m.Author.ID {
						delete(guessMap, guess)
					}
				}
				// Add to map
				guessMap[guess] = m.Author.ID
			}
		}
		// Print current map
		log.Println("Current map:")
		for guess, guesser := range guessMap {
			log.Println(fmt.Sprintf("%s guessed %s", guesser, guess))
		}
		log.Println("End of Current map\n")
	}
}
