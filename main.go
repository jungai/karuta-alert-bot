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
	grabStatus = "off"
)

// reg
var (
	dropPattern, _     = regexp.Compile(`(^\<\@\d{18}\>) is dropping \d cards!$`)
	grabPattern, _     = regexp.Compile(`(^\<\@\d{18}\>) took the \*\*.*\*\* card .\*!`)
	dailyPattern, _    = regexp.Compile(`you earned a daily reward`)
	purchasePattern, _ = regexp.Compile(`(^\<\@\d{18}\>), please follow this link to complete your purchase`)
	workPattern, _     = regexp.Compile(`\*\*Your workers have finished their tasks.\*\*`)
	checkIsDigit, _    = regexp.Compile(`^[0-9]+$`)
)

var (
	simpleDontBuyMessage = "อย่าเติมเลยค้าบบ 😅"
	hardDontBuyMessage   = "ก็บอกว่าอย่าเติมมม 😡"
	dropMessage          = "**Drop** currently available 😗"
	grabMessage          = "**Grab** currently available 😉"
	dailyMessage1        = "see ya for next daily is available 🥳"
	dailyMessage2        = "**Daily** currently available 😉"
	workMessage1         = "I will notify you when work already avaliable 🌝"
	workMessage2         = "**Work** currently available 🥳"
)

func getUser(id string) string {
	return fmt.Sprintf("<@%v>", id)
}

func isValidPrefix(p string) bool {
	return p == prefix || p == alias
}

func embedHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:     "Help",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://media3.giphy.com/media/26Ff4P2zcsiIi6fQY/giphy.gif?cid=ecf05e47ocriucm3ha0ikjwwirjxjshcdv4jwajouf0uln4t&rid=giphy.gif&ct=g"},
		Color:     58,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v status**", prefix),
				Value: "get currently status(drop, grab, alias)",
			},
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v drop <value>**", prefix),
				Value: "set drop status (value are 'on' or 'off)",
			},
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v grab <value>**", prefix),
				Value: "set grab status (value are 'on' or 'off)",
			},
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v count <number> msg?** ", prefix),
				Value: "timer (1,2,3,4,5,....) for minute",
			},
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v count <number>hr msg?** ", prefix),
				Value: "timer (1,2,3,4,5,....) for hour",
			},
			&discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("**%v alias <alias>**", prefix),
				Value: "change alias",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "source code -> https://github.com/jungai/nonnmonn-v2",
		},
	}
}

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		fmt.Println("Cannot connect to server")

		return
	}

	discord.AddHandler(messageCreate)
	discord.AddHandler(messageUpdate)

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
		user := strings.Split(m.Content, " ")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", user[0], dropMessage))

		return

	case grabPattern.MatchString(m.Content) && grabStatus == On:
		time.Sleep(time.Minute * 30)
		user := strings.Split(m.Content, " ")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", user[0], grabMessage))

		return

	case dailyPattern.MatchString(m.Content):
		s.ChannelMessageSend(m.ChannelID, dailyMessage1)
		time.Sleep(time.Second * 84_600)
		user := strings.Split(m.Content, " ")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", user[0], dailyMessage2))

		return

	case purchasePattern.MatchString(m.Content):
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, hardDontBuyMessage)

		return
	}

	// Check Embed
	if len(m.Embeds) > 0 {
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
	case "help":
		s.ChannelMessageSendEmbed(m.ChannelID, embedHelp())

		return

	case "drop", "grab":
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
		time.Sleep(time.Minute * 30)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v already 30 min 😏", getUser(m.Author.ID)))

		return

	case "cg":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v see ya in 10 min", getUser(m.Author.ID)))
		time.Sleep(time.Minute * 10)
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

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if len(m.Embeds) > 0 {
		switch {
		case m.Embeds[0].Title == "Work" && workPattern.MatchString(m.Embeds[0].Description):
			user := strings.Split(m.Embeds[0].Description, ",")
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", user[0], workMessage1))
			time.Sleep(time.Second * 42_300) // 11.45 hour
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v %v", user[0], workMessage2))

			return
		}

	}
}
