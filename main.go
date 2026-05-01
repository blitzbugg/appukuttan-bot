package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var spotifyClient *spotify.Client

// Sarcastic responses for plain (non-command) messages
var sarcasticReplies = []string{
	"Kalajittu podey! paatu keeken vannekkunu 😏 `!help` nokk",
	"Over aaki chalam aakathadey! Try `!help` instead 😒",
	"Eeekaantha chandrikeee..  baaki kelkaan`/songs` or `!help`😎",
	"Ammachi angane alla! `!help` nokk🙄",
	"Mahadevaa.. ee message-um command alla! `!help` nokk 😤",
	"Enthaadaa kuttaa...😌`!help` nokk",
}

const helpMessage = "🎵 **Appukuttan - Malayalam Mood Bot**\n\n" +
	"Here are the commands you can use:\n\n" +
	"🎭 **Mood Commands**\n" +
	"> `!sad`      — Sad Malayalam songs\n" +
	"> `!happy`    — Happy Malayalam songs\n" +
	"> `!romantic` — Romantic Malayalam songs\n" +
	"> `!angry`    — Angry Malayalam songs\n" +
	"> `!chill`    — Chill Malayalam songs\n\n" +
	"🎛️ **Interactive**\n" +
	"> `/songs` — Pick a mood from a dropdown menu\n\n" +
	"❓ **Help**\n" +
	"> `!help` or `/help` — Show this message"

func main() {
	go func() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Bot is running 🚀"))
    })
    http.ListenAndServe(":8080", nil)
}()
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	go startHTTPServer()
	setupSpotify()

	dg := setupDiscord()

	fmt.Println("Appukuttan is now running. Press CTRL+C to exit.")
	waitForShutdown(dg)
}

// -------------------- HTTP SERVER --------------------

func startHTTPServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bot is running"))
	})

	fmt.Println("HTTP server running on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("HTTP server error:", err)
	}
}

// -------------------- SPOTIFY SETUP --------------------

func setupSpotify() {
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(context.Background())
	if err != nil {
		fmt.Println("Error getting Spotify token:", err)
		os.Exit(1)
	}

	httpClient := spotifyauth.New().Client(context.Background(), token)
	spotifyClient = spotify.New(httpClient)
}

// -------------------- DISCORD SETUP --------------------

func setupDiscord() *discordgo.Session {
	token := os.Getenv("DISCORD_BOT_TOKEN")

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		os.Exit(1)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(interactionCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord connection:", err)
		os.Exit(1)
	}

	registerSlashCommands(dg)

	return dg
}

// -------------------- SLASH COMMAND REGISTRATION --------------------

func registerSlashCommands(dg *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Show all available Appukuttan commands",
		},
		{
			Name:        "songs",
			Description: "Pick a mood and get Malayalam song suggestions",
		},
	}

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			fmt.Printf("Error creating slash command %s: %v\n", cmd.Name, err)
		} else {
			fmt.Printf("Registered slash command: /%s\n", cmd.Name)
		}
	}
}

// -------------------- INTERACTION HANDLER --------------------

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	// Slash commands
	case discordgo.InteractionApplicationCommand:
		switch i.ApplicationCommandData().Name {
		case "help":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: helpMessage,
				},
			})

		case "songs":
			// Reply with a select menu dropdown
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🎵 Pick a mood to get Malayalam song suggestions:",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID:    "mood_select",
									Placeholder: "Choose your mood...",
									Options: []discordgo.SelectMenuOption{
										{Label: "😢 Sad", Value: "sad", Description: "Melancholic Malayalam tracks"},
										{Label: "😄 Happy", Value: "happy", Description: "Upbeat feel-good songs"},
										{Label: "💕 Romantic", Value: "romantic", Description: "Love & romance vibes"},
										{Label: "😤 Angry", Value: "angry", Description: "High energy intense songs"},
										{Label: "😎 Chill", Value: "chill", Description: "Relaxed laid-back tracks"},
									},
								},
							},
						},
					},
				},
			})
		}

	// Select menu interaction
	case discordgo.InteractionMessageComponent:
		if i.MessageComponentData().CustomID == "mood_select" {
			selected := i.MessageComponentData().Values[0] // e.g. "sad"
			mood := "malayalam " + selected
			suggestions := getSpotifySuggestions(mood, 5)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Here are 5 %s songs:\n%s", selected, suggestions),
				},
			})
		}
	}
}

// -------------------- MESSAGE HANDLER --------------------

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Plain message (no "!" prefix) — reply with sarcasm
	if !strings.HasPrefix(m.Content, "!") {
		rand.Seed(time.Now().UnixNano())
		reply := sarcasticReplies[rand.Intn(len(sarcasticReplies))]
		s.ChannelMessageSend(m.ChannelID, reply)
		return
	}

	command := strings.ToLower(strings.TrimPrefix(m.Content, "!"))

	switch command {
	case "help":
		s.ChannelMessageSend(m.ChannelID, helpMessage)

	case "sad", "happy", "romantic", "angry", "chill":
		mood := "malayalam " + command
		suggestions := getSpotifySuggestions(mood, 5)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Here are 5 %s songs:\n%s", command, suggestions))

	default:
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("❓ Unknown command `!%s`. Type `!help` to see what I can do!", command))
	}
}

// -------------------- SPOTIFY LOGIC --------------------

func getSpotifySuggestions(query string, limit int) string {
	results, err := spotifyClient.Search(
		context.Background(),
		query,
		spotify.SearchTypeTrack,
		spotify.Limit(20),
	)
	if err != nil {
		return "Oops, I couldn't fetch songs right now!"
	}

	tracks := results.Tracks.Tracks
	if len(tracks) == 0 {
		return "No songs found!"
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tracks), func(i, j int) {
		tracks[i], tracks[j] = tracks[j], tracks[i]
	})

	var response strings.Builder
	count := 0

	for _, track := range tracks {
		if count >= limit {
			break
		}

		response.WriteString(fmt.Sprintf(
			"%d. %s - %s (%s)\n",
			count+1,
			track.Name,
			track.Artists[0].Name,
			track.ExternalURLs["spotify"],
		))
		count++
	}

	return response.String()
}

// -------------------- GRACEFUL SHUTDOWN --------------------

func waitForShutdown(dg *discordgo.Session) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	fmt.Println("Shutting down bot...")
	dg.Close()
}