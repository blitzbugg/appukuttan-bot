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

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Start HTTP server (required for Render)
	go startHTTPServer()

	// Setup Spotify client
	setupSpotify()

	// Setup Discord bot
	dg := setupDiscord()

	// Wait for shutdown signal
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

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord connection:", err)
		os.Exit(1)
	}

	return dg
}

// -------------------- MESSAGE HANDLER --------------------

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!") {
		command := strings.TrimPrefix(m.Content, "!")
		command = strings.ToLower(command)

		var mood string

		switch command {
		case "sad":
			mood = "malayalam sad"
		case "happy":
			mood = "malayalam happy"
		case "romantic":
			mood = "malayalam romantic"
		case "angry":
			mood = "malayalam angry"
		case "chill":
			mood = "malayalam chill"
		default:
			s.ChannelMessageSend(m.ChannelID,
				"I can suggest Malayalam songs based on your mood! Try !sad, !happy, !romantic, !angry, or !chill.")
			return
		}

		suggestions := getSpotifySuggestions(mood, 5)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Here are 5 %s songs:\n%s", command, suggestions))
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
		return "Oops, I couldn’t fetch songs right now!"
	}

	tracks := results.Tracks.Tracks
	if len(tracks) == 0 {
		return "No songs found!"
	}

	// Shuffle results
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