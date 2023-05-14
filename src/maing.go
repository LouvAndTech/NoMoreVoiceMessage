package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	BotToken       *string
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

/* === Global variables === */
var s *discordgo.Session

// init is called before main
func init() {
	log.Println("Starting...")
	//Get the bot token from the environment variables
	if os.Getenv("TOKEN") == "" {
		log.Fatalf("You need to pass the bot token as an argument")
	}
	log.Println("Bot token:", os.Getenv("TOKEN"))
	BotToken = flag.String("token", os.Getenv("TOKEN"), "Bot access token")
	flag.Parse()

	// Initialize the bot
	log.Println("Initializing bot...")
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	// Add Handeler for the ready event
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	s.AddHandler(messageCreate)
	s.AddHandler(messageReactionAdd)

	// In this example, we only care about receiving message events.
	s.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildMessageReactions

	err := s.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}

// On any message on the server
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	//Log the whole message
	//log.Printf("Message: \nAuthor: %v#%v\nContent: %v\nAt : %v\nType: %v\nAttachement: %v", m.Author.Username, m.Author.Discriminator, m.Content, m.Timestamp, m.Type, m.Attachments[0].)

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	//if the content is empty and there is an attachment
	if m.Content == "" && len(m.Attachments) > 0 {
		//If the attachment is a voice message
		if m.Attachments[0].ContentType == "audio/ogg" {
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "🇫🇷")
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "🇺🇸")
		}
	}
}

func messageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore all reactions created by the bot itself
	if r.UserID == s.State.User.ID {
		return
	}
	m, _ := s.ChannelMessage(r.ChannelID, r.MessageID)
	// only care if the reaction is on a voice message and if it's with a flag
	if m.Attachments[0].ContentType == "audio/ogg" && (r.Emoji.Name == "🇫🇷" || r.Emoji.Name == "🇺🇸") {
		//Send a message to say that we are converting the voice message
		prov, _ := s.ChannelMessageSendReply(m.ChannelID, "I see you wanna convert a voice message ! Please wait while I convert it to text...", m.Reference())
		//Check the emoji
		var txt string
		var err error
		switch r.Emoji.Name {
		case "🇫🇷":
			//Convert using the french model
			txt, err = ToText(m.Attachments[0].URL, "fr_fr")
		case "🇺🇸":
			//Convert using the english model
			txt, err = ToText(m.Attachments[0].URL, "en_us")
		default:
			return
		}

		//Remove the reactions
		_ = s.ChannelMessageDelete(prov.ChannelID, prov.ID)
		if err != nil {
			log.Println("Error while converting voice to text:", err)
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Sorry, I couldn't convert your voice message to text.", m.Reference())
			return
		}
		//Send the text
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Here is the text:\n```"+txt+"```", m.Reference())
	}
}
