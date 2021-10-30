package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	prefix = "mon"
)

// enum
const (
	On  = "on"
	Off = "off"
)

// state
var (
	alias      = "mon"
	dropStatus = "on"
	grabStatus = "on"
)

// reg
var (
	dropPattern, _     = regexp.Compile(`(^\<\@\d{18}\>) is dropping \d cards!$`)
	grabPattern, _     = regexp.Compile(`(^\<\@\d{18}\>) took the \*\*.*\*\* card .\*!`)
	dailyPattern, _    = regexp.Compile(`you earned a daily reward`)
	purchasePattern, _ = regexp.Compile(`(^\<\@\d{18}\>), please follow this link to complete your purchase`)
	checkIsDigit, _    = regexp.Compile(`^[0-9]+$`)
)

var (
	simpleDontBuyMessage = "อย่าเติมเลยค้าบบ 😅"
	hardDontBuyMessage   = "ก็บอกว่าอย่าเติมมม 😡"
	dropMessage          = "**Drop** currently available 😗"
	grabMessage          = "**Grab** currently available 😉"
	dailyMessage1        = "see ya for next daily is available 🥳"
	dailyMessage2        = "**Daily** currently available 😉"
)

func getUser(id string) string {
	return fmt.Sprintf("<@%v>", id)
}

func isValidPrefix(p string) bool {
	return p == prefix || p == alias
}

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		fmt.Println("Cannot connect to server")

		return
	}

	discord.AddHandler(messageCreate)

	// Identity
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// for karuta bot here
	switch {
	case dropPattern.MatchString(m.Content) && dropStatus == On:
		time.Sleep(time.Minute * 30)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", getUser(m.Author.ID), dropMessage))

		return

	case grabPattern.MatchString(m.Content) && grabStatus == On:
		time.Sleep(time.Minute * 30)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", getUser(m.Author.ID), grabMessage))

		return

	case dailyPattern.MatchString(m.Content):
		s.ChannelMessageSend(m.ChannelID, dailyMessage1)
		time.Sleep(time.Second * 84_600)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", getUser(m.Author.ID), dailyMessage2))

		return

	case purchasePattern.MatchString(m.Content):
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, hardDontBuyMessage)

		return
	}

	// Check Embed
	if len(m.Embeds) != 0 {
		switch {
		case m.Embeds[0].Title == "Purchase Gems":
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			s.ChannelMessageSend(m.ChannelID, simpleDontBuyMessage)

			return
		}
	}

	// user command
	rawMonCommand := strings.Split(m.Content, " ")

	monCommand := make([]string, 4)
	copy(monCommand, rawMonCommand)

	prefix, command, value1, value2 := monCommand[0], monCommand[1], monCommand[2], monCommand[3]

	// check prefix
	if !isValidPrefix(prefix) {
		return
	}

	switch command {
	case "drop":
	case "grab":
		if value1 != On && value1 != Off {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%v** is not valid <on | off>", command))

			return
		}

		if command == "drop" {
			dropStatus = value1
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%v** is **%v**", command, value1))

			return
		}

		if command == "grab" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%v** is **%v**", command, value1))
			grabStatus = value1

			return
		}

		return

	case "alias":
		alias = value1
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**alias** is set to -> **%v**", value1))

		return
	case "cd":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in 30 min", getUser(m.Author.ID)))
		time.Sleep(time.Second * 3)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v already 30 min 😏", getUser(m.Author.ID)))

		return

	case "cg":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in 10 min", getUser(m.Author.ID)))
		time.Sleep(time.Second * 2)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v already 10 min 😏", getUser(m.Author.ID)))

		return

	case "vi":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in 2 hours", getUser(m.Author.ID)))
		time.Sleep(time.Hour * 2)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v vi kub", getUser(m.Author.ID)))

		return

	case "count":
		if condi := strings.Split(value1, "hr"); checkIsDigit.MatchString(condi[0]) && len(condi) == 2 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in %v hours", getUser(m.Author.ID), condi[0]))
			realNumber, _ := strconv.ParseInt(condi[0], 10, 32)
			time.Sleep(time.Hour * time.Duration(realNumber))

			if value2 == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v already %v hours 😏", getUser(m.Author.ID), value1))

				return
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v -> %v", getUser(m.Author.ID), value2))

			return

		}

		if checkIsDigit.MatchString(value1) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in %v min", getUser(m.Author.ID), value1))
			realNumber, _ := strconv.ParseInt(value1, 10, 32)
			time.Sleep(time.Minute * time.Duration(realNumber))

			if value2 == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v already %v min 😏", getUser(m.Author.ID), value1))

				return
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v -> %v", getUser(m.Author.ID), value2))

			return

		}
	}

}
