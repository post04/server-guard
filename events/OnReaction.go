package events

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Captcha is a struct that holds all captcha information
type Captcha struct {
	UserID       string
	Emoji        string
	GuildID      string
	MessageID    string
	WrongAnswers []string
	Tries        int
}

// OnReaction listens for Captcha reactions to get a Captcha
func (c *Config) OnReaction(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	// The reaction is on a Captcha, this can only be in a dm anyways
	if cap, ok := c.CaptchaListeners[reaction.UserID]; ok {
		if reaction.Emoji.Name == cap.Emoji && reaction.MessageID == cap.MessageID {
			c.CaptchaListeners[reaction.UserID].WrongAnswers = append(c.CaptchaListeners[reaction.UserID].WrongAnswers, reaction.Emoji.Name)
			err := s.GuildMemberRoleAdd(cap.GuildID, reaction.UserID, c.VerifiedRole)
			if err != nil {
				fmt.Println(err)
				return
			}
			_, err = s.ChannelMessageSend(reaction.ChannelID, "You answered correctly!")
			if err != nil {
				fmt.Println(err)
			}
			if c.ModChannel != "" {
				user, err := s.State.Member(cap.GuildID, reaction.UserID)
				if err != nil {
					user, err = s.GuildMember(cap.GuildID, reaction.UserID)
					if err != nil {
						fmt.Println(err)
						return
					}
				}

				_, err = s.ChannelMessageSendEmbed(c.ModChannel, &discordgo.MessageEmbed{
					Title:       "Verified " + user.User.Username + "#" + user.User.Discriminator,
					Description: fmt.Sprintf("Captcha Information\n\nCorrect Emoji: %s\nResponse Emojis: %s", cap.Emoji, strings.Join(c.CaptchaListeners[reaction.UserID].WrongAnswers, ", ")),
				})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			delete(c.CaptchaListeners, reaction.UserID)
			return
		}
		if cap.MessageID == reaction.MessageID && reaction.Emoji.Name != cap.Emoji {
			user, err := s.State.Member(cap.GuildID, reaction.UserID)
			if err != nil {
				user, err = s.GuildMember(cap.GuildID, reaction.UserID)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			c.CaptchaListeners[reaction.UserID].WrongAnswers = append(c.CaptchaListeners[reaction.UserID].WrongAnswers, reaction.Emoji.Name)
			c.CaptchaListeners[reaction.UserID].Tries = cap.Tries + 1
			if c.CaptchaListeners[reaction.UserID].Tries >= c.CaptchaTries {
				if c.Punishment == "ban" {
					c.Punishment = "bann"
				}
				_, err := s.ChannelMessageSend(reaction.ChannelID, "You failed the captcha and ran out of tries! You will now be "+c.Punishment+"ed!")
				if err != nil {
					fmt.Println(err)
					return
				}
				if c.Punishment == "ban" || c.Punishment == "bann" {
					err = s.GuildBanCreateWithReason(cap.GuildID, reaction.UserID, "Failed captcha", 0)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				if c.Punishment == "kick" {
					err = s.GuildMemberDeleteWithReason(cap.GuildID, cap.UserID, "failed captcha")
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				if c.ModChannel != "" {
					_, err = s.ChannelMessageSendEmbed(c.ModChannel, &discordgo.MessageEmbed{
						Title:       fmt.Sprintf("%s#%s failed captcha %v times", user.User.Username, user.User.Discriminator, c.CaptchaTries),
						Description: fmt.Sprintf("Captcha Information\n\nAnswers: %s\nCorrect Answer: %s", strings.Join(c.CaptchaListeners[reaction.UserID].WrongAnswers, ", "), cap.Emoji),
					})
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				delete(c.CaptchaListeners, reaction.UserID)
				return
			}
			_, err = s.ChannelMessageSend(reaction.ChannelID, fmt.Sprintf("You have failed the captcha! You have %v more tries!", (c.CaptchaTries-c.CaptchaListeners[reaction.UserID].Tries)))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		return
	}
	// End of dm Captcha listener

	// Start of Captcha system
	if reaction.Emoji.Name == c.CaptchaReaction && reaction.MessageID == c.CaptchaMessageID {
		err := s.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, reaction.Emoji.Name, reaction.UserID)
		if err != nil {
			fmt.Println(err)
		}
		user, err := s.State.Member(reaction.GuildID, reaction.UserID)
		if err != nil {
			user, err = s.GuildMember(reaction.GuildID, reaction.UserID)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		image, emoji, err := c.NewCaptcha()
		if err != nil {
			fmt.Println(err)
			return
		}
		dm, err := s.UserChannelCreate(reaction.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "Please react with the emoji seen in the image below!")
		if err != nil {
			fmt.Println(err)
			return
		}
		msg, err := s.ChannelFileSend(dm.ID, "test.jpg", image)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.CaptchaListeners[user.User.ID] = &Captcha{
			UserID:       user.User.ID,
			Emoji:        emoji,
			GuildID:      reaction.GuildID,
			MessageID:    msg.ID,
			Tries:        0,
			WrongAnswers: []string{},
		}
	}
}
