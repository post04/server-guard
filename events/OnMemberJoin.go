package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Config stores everything in `config.json`
type Config struct {
	Token            string        `json:"token"`
	ModChannel       string        `json:"modChannel"`
	Punishment       string        `json:"punishment"`
	SusLevel         int           `json:"susLevel"`
	CheckBadges      bool          `json:"checkBadges"`
	CheckAccountAge  bool          `json:"checkAccountAge"`
	CheckAvatar      bool          `json:"checkAvatar"`
	VerifiedRole     string        `json:"verifiedRole"`
	CaptchaMessageID string        `json:"captchaMessage"`
	CaptchaReaction  string        `json:"captchaReaction"`
	CaptchaTries     int           `json:"captchaTries"`
	AccountAgeMS     int64         `json:"accountAgeMS"`
	BotLimit         int           `json:"botLimit"`
	JoinTime         time.Duration `json:"joinTime"`
	DetectRaids      bool          `json:"detectRaids"`

	Emojis           []string
	CaptchaListeners map[string]*Captcha
	PreviousJoins    map[string][]*JoinedUser
	Session          *discordgo.Session
}

// JoinedUser holds information about a user that joined including their name, userID, time of join, etc.
type JoinedUser struct {
	UserID    string
	JoinedAt  time.Time
	Username  string
	Discrim   string
	GuildID   string
	Bot       bool
	AvatarURL string
	Punished  bool
}

// PunishBots ranges thru all the bots and punishes any unpunished bots
func (c *Config) PunishBots(bots []*JoinedUser, s *discordgo.Session) {
	if !c.DetectRaids {
		return
	}
	// Range thru all the bots and punsish any bot that isn't already punished
	for i, bot := range bots {
		// Punished unpunished bots
		if !bot.Punished {
			b := bot
			b.Punished = true
			b.Bot = true
			c.PreviousJoins[bot.Username][i] = b
			//fmt.Printf("%+v\n", c.PreviousJoins[bot.Username][i])
			if c.Punishment == "ban" || c.Punishment == "bann" {
				err := s.GuildBanCreateWithReason(bot.GuildID, bot.UserID, "User involved in a bot raid", 0)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if c.Punishment == "kick" {
				err := s.GuildMemberDeleteWithReason(bot.GuildID, bot.UserID, "User involved in a bot raid")
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
					Title:       fmt.Sprintf("%sed %s#%s", c.Punishment, bot.Username, bot.Discrim),
					Description: fmt.Sprintf("Reason: detected as a bot due to influx of joins.\npunished users: %v", i+1),
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: bot.AvatarURL,
					},
				})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
	// Done ranging thru unpunished bots
}

// CheckBots checks all the bots every x time and bans raids
func (c *Config) CheckBots(GcCycle time.Duration) {
	if !c.DetectRaids {
		return
	}
	garbageCollector := time.NewTicker(GcCycle)
	for range garbageCollector.C {
		for _, bots := range c.PreviousJoins {
			// Detecting a bot from a flood
			if len(bots) > c.BotLimit {
				c.PunishBots(bots, c.Session)
				return
			}
			for _, bot := range bots {
				if bot.Bot {
					c.PunishBots(bots, c.Session)
					break
				}
			}
		}
	}
}

// ClearBots clears accounts that trigger raid detection after every x time.Duration
// special thanks to this commit on another project named dr-docso for providing the code for this :)
// https://github.com/post04/dr-docso/commit/fbfa4aa1b80b604b2c03f9c6c5782aadcd42e071#diff-92ef575f410b265e12970f2e513257450e9f390d36fdc623f095937b564d164f
func (c *Config) ClearBots(GcCycle time.Duration) {
	if !c.DetectRaids {
		return
	}
	garbageCollector := time.NewTicker(GcCycle)
	for range garbageCollector.C {
		for key, users := range c.PreviousJoins {
			final := []*JoinedUser{}
			for _, user := range users {
				if time.Since(user.JoinedAt) < c.JoinTime*time.Millisecond {
					final = append(final, user)
				}
			}
			c.PreviousJoins[key] = final
		}
	}
}

// OnMemberJoin waits for a member to join then does stuff to decide if they're legit or not
func (c *Config) OnMemberJoin(s *discordgo.Session, member *discordgo.GuildMemberAdd) {
	if member.User.Bot {
		return
	}
	creationDate, err := discordgo.SnowflakeTimestamp(member.User.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	creationMs := creationDate.UnixNano() / 1000000
	// try to detect a raid and ban the flooding users
	if bots, ok := c.PreviousJoins[member.User.Username]; ok && c.DetectRaids {

		// Detecting a bot from a flood
		if len(bots) > c.BotLimit {
			c.PreviousJoins[member.User.Username] = append(c.PreviousJoins[member.User.Username], &JoinedUser{UserID: member.User.ID, JoinedAt: creationDate, Username: member.User.Username, Bot: true, Punished: false, GuildID: member.GuildID, AvatarURL: member.User.AvatarURL(""), Discrim: member.User.Discriminator})
			return
		}
		for _, bot := range bots {
			if bot.Bot {
				c.PreviousJoins[member.User.Username] = append(c.PreviousJoins[member.User.Username], &JoinedUser{UserID: member.User.ID, JoinedAt: creationDate, Username: member.User.Username, Bot: true, Punished: false, GuildID: member.GuildID, AvatarURL: member.User.AvatarURL(""), Discrim: member.User.Discriminator})
				return
			}
		}
	}
	// see if the user is a bot using checks like pfp, badges, etc.
	susLevel := 0
	hasAvatar := true
	oldAccount := true
	hasBadges := true
	if member.User.Avatar == "" && c.CheckAvatar {
		susLevel++
		hasAvatar = false
	}
	//86400000
	if time.Now().UnixNano()/1000000 < creationMs+c.AccountAgeMS && c.CheckAccountAge {
		oldAccount = false
		susLevel++
	}
	if member.User.PublicFlags == 0 && c.CheckBadges {
		hasBadges = false
		susLevel++
	}
	if susLevel >= c.SusLevel {
		c.PreviousJoins[member.User.Username] = append(c.PreviousJoins[member.User.Username], &JoinedUser{UserID: member.User.ID, JoinedAt: creationDate, Username: member.User.Username, Bot: true, Punished: true, GuildID: member.GuildID, AvatarURL: member.User.AvatarURL(""), Discrim: member.User.Discriminator})
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
		if c.Punishment == "ban" || c.Punishment == "bann" {
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
			if c.Punishment == "ban" {
				c.Punishment = "bann"
			}
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
		return
	}
	c.PreviousJoins[member.User.Username] = append(c.PreviousJoins[member.User.Username], &JoinedUser{UserID: member.User.ID, JoinedAt: creationDate, Username: member.User.Username, Bot: false, GuildID: member.GuildID, AvatarURL: member.User.AvatarURL(""), Discrim: member.User.Discriminator})
}
