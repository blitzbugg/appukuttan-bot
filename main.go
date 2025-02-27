package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var spotifyClient *spotify.Client

func main() {

	err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file:", err)
        return
    }
    // Discord setup
    token := os.Getenv("DISCORD_BOT_TOKEN")
    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        fmt.Println("Error creating Discord session:", err)
        return
    }

    // Spotify setup with Client Credentials Flow
    config := &clientcredentials.Config{
        ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
        ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
        TokenURL:     spotifyauth.TokenURL,
    }
    spotifyToken, err := config.Token(context.Background())
    if err != nil {
        fmt.Println("Error getting Spotify token:", err)
        return
    }
    httpClient := spotifyauth.New().Client(context.Background(), spotifyToken)
    spotifyClient = spotify.New(httpClient)

    dg.AddHandler(messageCreate)
    err = dg.Open()
    if err != nil {
        fmt.Println("Error opening connection:", err)
        return
    }

    fmt.Println("Appukuttan is now running. Press CTRL+C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc

    dg.Close()
}


func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }

    if strings.HasPrefix(m.Content, "!") {
        command := strings.TrimPrefix(m.Content, "!")
        switch strings.ToLower(command) {
		case "sad":
			suggestions := getSpotifySuggestions("malayalam sad", 5)
			s.ChannelMessageSend(m.ChannelID, "Here are 5 sad Malayalam songs from Spotify:\n"+suggestions)
		case "happy":
			suggestions := getSpotifySuggestions("malayalam happy", 5)
			s.ChannelMessageSend(m.ChannelID, "Here are 5 happy Malayalam songs from Spotify:\n"+suggestions)
		case "romantic":
			suggestions := getSpotifySuggestions("malayalam romantic", 5)
			s.ChannelMessageSend(m.ChannelID, "Here are 5 romantic Malayalam songs from Spotify:\n"+suggestions)
		case "angry":
			suggestions := getSpotifySuggestions("malayalam angry", 5)
			s.ChannelMessageSend(m.ChannelID, "Here are 5 angry Malayalam songs from Spotify:\n"+suggestions)
		case "chill":
			suggestions := getSpotifySuggestions("malayalam chill", 5)
			s.ChannelMessageSend(m.ChannelID, "Here are 5 chill Malayalam songs from Spotify:\n"+suggestions)
		default:
			s.ChannelMessageSend(m.ChannelID, "I can suggest Malayalam songs based on your mood! Try !sad, !happy, !romantic, !angry, or !chill.")
		}
    }
}

func getSpotifySuggestions(query string, limit int) string {
    results, err := spotifyClient.Search(context.Background(), query, spotify.SearchTypeTrack)
    if err != nil {
        return "Oops, I couldn’t fetch songs right now!"
    }

    var response strings.Builder
    count := 0
    for _, track := range results.Tracks.Tracks {
        if count >= limit {
            break
        }
        response.WriteString(fmt.Sprintf("%d. %s - %s (%s)\n", count+1, track.Name, track.Artists[0].Name, track.ExternalURLs["spotify"]))
        count++
    }
    return response.String()
}