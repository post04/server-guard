package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// OnMessage listens for mass ping and raid attacks
func (c *Config) OnMessage(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.Bot {
		return
	}

	if c.AntiMassPing && c.AntiMassPingLimit > len(msg.Mentions) {
		creationDate, err := discordgo.SnowflakeTimestamp(msg.Author.ID)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.PreviousJoins[msg.Author.Username] = append(c.PreviousJoins[msg.Author.Username], &JoinedUser{UserID: msg.Author.ID, JoinedAt: creationDate, Username: msg.Author.Username, Bot: true, Punished: true, GuildID: msg.GuildID, AvatarURL: msg.Author.AvatarURL(""), Discrim: msg.Author.Discriminator})
		if c.Punishment == "ban" || c.Punishment == "bann" {
			err := s.GuildBanCreateWithReason(msg.GuildID, msg.Author.ID, "User involved in mass pinging", 0)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if c.Punishment == "kick" {
			err := s.GuildMemberDeleteWithReason(msg.GuildID, msg.Author.ID, "User involved in mass pinging")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if c.ModChannel != "" && c.Punishment != "" {
			if c.Punishment == "ban" {
				c.Punishment = "bann"
			}
			_, err := s.ChannelMessageSendEmbed(c.ModChannel, &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%sed %s#%s", c.Punishment, msg.Author.Username, msg.Author.Discriminator),
				Description: fmt.Sprintf("Reason: Mass ping\nPings: %v", len(msg.Mentions)),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: msg.Author.AvatarURL(""),
				},
			})
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
