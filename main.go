package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func input(reader *bufio.Reader) string {
	raw, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Something went wrong when reading stdin: " + err.Error())
		return ""
	}
	switch runtime.GOOS {
	case "windows":
		return strings.Replace(raw, "\r\n", "", -1)
	default:
		return strings.Replace(raw, "\n", "", 1)
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please input the spam words:")
	words := input(reader)
	fmt.Println("Please input the spam delay:")
	delayRaw := input(reader)
	delay, err := strconv.Atoi(delayRaw)
	if err != nil {
		fmt.Println("Invalid delay.")
		return
	}
	fmt.Println("Please input your discord token:")
	token := input(reader)
	client, err1 := discordgo.New(token)
	err2 := client.Open()
	if err1 != nil || err2 != nil {
		fmt.Println("Could not open session")
		fmt.Println(err1, err2)
	}
	fmt.Println("------------------------")
	var channelMap map[int]*discordgo.Channel
	channelMap = make(map[int]*discordgo.Channel)
	var guildMap map[int]map[int]*discordgo.Channel
	guildMap = make(map[int]map[int]*discordgo.Channel)
	for guildIndex, guild := range client.State.Guilds {
		fmt.Print("GuildNumber: ", guildIndex, " [", guild.Name, ":", guild.ID, "]", "\n")
		for channelIndex, channel := range guild.Channels {
			fmt.Print("\tChannelNumber: ", channelIndex, " [", channel.Name, ":", channel.ID, "]\n")
			channelMap[channelIndex] = channel
		}
		guildMap[guildIndex] = channelMap
		channelMap = nil
	}
	fmt.Println("Please select channels to spam in, separated by a comma. Syntax is the following: GuildNumber-ChannelNumber : 1-12,0-15")
	spamIn := input(reader)
	spamDivided := strings.Split(spamIn, ",")
	channels := channelParser(spamDivided, guildMap)
	fmt.Println("Spamming in...")
	for _, channel := range channels {
		fmt.Println(channel.Name)
	}
	spam(client, channels, words, delay)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-quit
	fmt.Println("Quitting...")
	client.Close()
}

func spam(client *discordgo.Session, channels []*discordgo.Channel, word string, delay int) {
	for {
		for _, channel := range channels {
			client.ChannelMessageSend(channel.ID, word)
		}
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func channelParser(spamDivided []string, guildMap map[int]map[int]*discordgo.Channel) []*discordgo.Channel {
	var allChannels []*discordgo.Channel
	for _, item := range spamDivided {
		raw := strings.Split(item, "-")
		guildNumber := raw[0]
		guildChannel := raw[1]
		cIndex, err := strconv.Atoi(guildChannel)
		if err != nil {
			return nil
		}
		gIndex, err := strconv.Atoi(guildNumber)
		if err != nil {
			return nil
		}
		if channels, ok := guildMap[gIndex]; ok {
			for channelNumber, channel := range channels {
				if channelNumber == cIndex {
					allChannels = append(allChannels, channel)
				}
			}
		}
	}
	return allChannels
}
