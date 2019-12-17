package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	DmID string
	BannedWordsFileName string
	BannedWords [][]string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&DmID, "d", "", "DM ID")
	flag.StringVar(&BannedWordsFileName, "w", "bannedWords.csv", "Banned Words File")
	flag.Parse()
}

func main() {
	bannedWordsFile, err := os.Open(BannedWordsFileName)
	if err != nil {
		fmt.Println("error opening banned words file,", err)
		return
	}

	bannedWordsReader := csv.NewReader(bannedWordsFile)
	bannedWordsReader.FieldsPerRecord = -1

	BannedWords, err = bannedWordsReader.ReadAll()
	if err != nil {
		fmt.Println("error reading banned words file,", err)
		return
	}

	for ci, line := range BannedWords {
		for ri, word := range line {
			BannedWords[ci][ri] = strings.ToUpper(word)
		}
	}

	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageUpdate)
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-kill

	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	checkWords(s, m.Message)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	checkWords(s, m.Message)
}

func checkWords (s *discordgo.Session, m *discordgo.Message) {
	message := strings.ToUpper(m.Content)
	for _, line := range BannedWords {
		for _, word := range line {
			if strings.Contains(message, string(word)) {
				err := s.ChannelMessageDelete(m.ChannelID, m.ID)
				if err != nil {
					fmt.Println("error deleting message,", err)
					errMessage := fmt.Sprintf("Travdog Error: Error deleting message,", err)
					s.ChannelMessageSend(DmID, errMessage)
				}
			}
		}
	}
}