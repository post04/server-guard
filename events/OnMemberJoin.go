package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Config stores everything in `config.json`
type Config struct {
	Token            string `json:"token"`
	ModChannel       string `json:"modChannel"`
	Punishment       string `json:"punishment"`
	SusLevel         int    `json:"susLevel"`
	CheckBadges      bool   `json:"checkBadges"`
	CheckAccountAge  bool   `json:"checkAccountAge"`
	CheckAvatar      bool   `json:"checkAvatar"`
	VerifiedRole     string `json:"verifiedRole"`
	CaptchaMessageID string `json:"captchaMessage"`
	CaptchaReaction  string `json:"captchaReaction"`
	CaptchaTries     int    `json:"captchaTries"`

	Emojis           []string `json:"emojis"`
	CaptchaListeners map[string]*Captcha
}

// OnMemberJoin waits for a member to join then does stuff to decide if they're legit or not
func (c *Config) OnMemberJoin(s *discordgo.Session, member *discordgo.GuildMemberAdd) {
	susLevel := 0
	hasAvatar := true
	oldAccount := true
	hasBadges := true
	if member.User.Avatar == "" && c.CheckAvatar {
		susLevel++
		hasAvatar = false
	}
	creationDate, err := discordgo.SnowflakeTimestamp(member.User.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	creationMs := creationDate.UnixNano() / 1000000
	//86400000
	if time.Now().UnixNano()/1000000 > creationMs+(86400000*7) && c.CheckAccountAge {
		oldAccount = false
		susLevel++
	}
	if member.User.PublicFlags == 0 && c.CheckBadges {
		hasBadges = false
		susLevel++
	}
	if susLevel >= c.SusLevel {
		description := ""
		if c.CheckAvatar {
			description += fmt.Sprintf("Has Avatar: %t\n", hasAvatar)
		}
		if c.CheckBadges {
			description += fmt.Sprintf("Has Badges: %t\n", hasBadges)
		}
		if c.CheckAccountAge {
			description += fmt.Sprintf("Aged Account: %t\n", oldAccount)
			minuteStr := fmt.Sprint(creationDate.Minute())
			if len(minuteStr) == 1 {
				minuteStr = "0" + minuteStr
			}
			description += "Account Age: " + fmt.Sprintf("%v/%s/%v @ %v:%v", creationDate.Day(), creationDate.Month().String(), creationDate.Year(), creationDate.Hour(), minuteStr)
		}
		if c.Punishment == "ban" {
			err := s.GuildBanCreateWithReason(member.GuildID, member.User.ID, fmt.Sprintf("User had a suslevel of %v", susLevel), 0)
			if err != nil {
				fmt.Println(err)
				return
			}

		}
		if c.Punishment == "kick" {
			err := s.GuildMemberDeleteWithReason(member.GuildID, member.User.ID, fmt.Sprintf("User had a suslevel of %v", susLevel))
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		if c.ModChannel != "" {
			_, err := s.ChannelMessageSendEmbed(c.ModChannel, &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%sed %s#%s", c.Punishment, member.User.Username, member.User.Discriminator),
				Description: description,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: member.User.AvatarURL(""),
				},
			})
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
