package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/color"
)

// Config file information
type Config struct {
	Token   string `json:"token"`   // User/bot token
	Agent   string `json:"agent"`   // User-agent for request(s)
	Prefix  string `json:"prefix"`  // Selfbot prefix
	Command string `json:"command"` // Selfbot delete command
	Verbose bool   `json:"verbose"` // Verbose mode
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // Alphabet for generation
var config Config                                                      // Access struct
var timers = []int{200, 300, 400, 500}                                 // 2-5 ms

func init() {
	rand.Seed(time.Now().UnixNano()) // Random seed for random numbers
}

func main() {
	// Get config
	configRead()

	// Login
	client, err := discordgo.New(config.Token)
	checkErr(err, false)

	// All events
	client.AddHandler(messageEvent) // Messages

	// All intents (server & dms)
	client.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)

	// Other
	client.UserAgent = config.Agent // Set useragent

	// Start connection
	err = client.Open()
	checkErr(err, false)

	// Multiple lines are neater
	art := `___  _ ____ ___  ____ _    
|  \ | [__  |  \ |___ |    
|__/ | ___] |__/ |___ |___ 
`
	color.Printf(colors("[Y]%s[E]\n[R]By Perpdox[E]\n\n"), art)
	color.Printf(colors("[B]Username: %s[E]\n"), client.State.User)
	color.Printf(colors("[G]User ID: %s[E]\n\n"), client.State.User.ID)

	// Block forever
	runtime.Goexit()

	// End connection
	client.Close()
}

// Read config file
func configRead() {
	// Open config file
	file, err := os.Open("config.json")
	checkErr(err, false)

	// Read file bytes
	content, err := io.ReadAll(file)
	checkErr(err, false)

	// Unmarshal into struct
	json.Unmarshal(content, &config)
}

// Check error(s)
func checkErr(err error, safety bool) {
	// Safety is for non exits
	if safety {
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	} else {
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(0)
		}
	}
}

// Color replacement
func colors(text string) string {
	text = strings.ReplaceAll(text, "[Y]", "<yellow>")
	text = strings.ReplaceAll(text, "[R]", "<red>")
	text = strings.ReplaceAll(text, "[G]", "<green>")
	text = strings.ReplaceAll(text, "[B]", "<blue>")
	text = strings.ReplaceAll(text, "[E]", "</>")
	return text
}

// Generate random message
func generateContent() string {
	// Random length between 0-20 (+5)
	length := rand.Intn(20) + 5
	messageContent := make([]byte, length)

	// Each byte index
	for m := range messageContent {
		// Assign byte a random letter
		// Chosen through a random index from letters
		messageContent[m] = letters[rand.Intn(len(letters))]
	}

	// Return generated string
	return string(messageContent)
}

// Any message creation event
func messageEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Message has command and is by author
	if strings.HasPrefix(m.Content, config.Prefix+config.Command) && m.Author.ID == s.State.User.ID {
		message := strings.Split(m.Content, " ") // Split content by spaces

		channelName, err := s.Channel(m.ChannelID) // Channel name
		checkErr(err, true)
		guildName, err := s.Guild(m.GuildID) // Guild name
		checkErr(err, true)

		// Number check
		if len(message) > 1 {
			amount, err := strconv.Atoi(message[1]) // Convert to int
			checkErr(err, true)

			messages, err := s.ChannelMessages(m.ChannelID, amount, "", "", "") // Get messages
			checkErr(err, true)

			// Go through each message
			for _, message := range messages {
				// Author message only
				if message.Author.ID == s.State.User.ID {
					s.ChannelMessageEdit(m.ChannelID, message.ID, generateContent())             // Edit message
					time.Sleep(time.Millisecond * time.Duration(timers[rand.Intn(len(timers))])) // Sleep

					s.ChannelMessageDelete(m.ChannelID, message.ID) // Delete message

					// Verbose prints information
					if config.Verbose {
						color.Printf(colors("[Y][*][E] Edited & deleted message in [B][%s] #%s[E] > [G]%s[E]\n"), guildName.Name, channelName.Name, message.Content)
					}
				}
			}
		} else {
			// Forever until no more messages
			var messageID string // Last message ID to continue from
			for {
				messages, err := s.ChannelMessages(m.ChannelID, 100, messageID, "", "") // Get messages
				checkErr(err, true)

				// Go through each message
				for _, message := range messages {
					// Author message only
					if message.Author.ID == s.State.User.ID {
						s.ChannelMessageEdit(m.ChannelID, message.ID, generateContent())             // Edit message
						time.Sleep(time.Millisecond * time.Duration(timers[rand.Intn(len(timers))])) // Sleep
						s.ChannelMessageDelete(m.ChannelID, message.ID)                              // Delete message

						messageID = message.ID // Store last message ID

						// Verbose prints information
						if config.Verbose {
							color.Printf(colors("[Y][*][E] Edited & deleted message in [B][%s] #%s[E] > [G]%s[E]\n"), guildName.Name, channelName.Name, message.Content)
						}
					}
				}
			}
		}
	}
}

/*
Note: If you're planning to delete more than 100 messages, do not specify
a number unless it's 100. The reason for this is because, Discord only
alllows fetching up to 100 at once.

The reason for storing the last message ID, is because we can only fetch
100 messages at a time. So every message, we should store the last message
-ID and continue where we left off.
*/
