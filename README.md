# server-guard
 A discord server guard which uses emojis and bot detection (WIP)

# features
- bot detection system
- custom captcha system\
    1.) This captcha system is like no other containing an emoji captcha, it uses images of emojis and requires a user to react with the emoji shown in the image to verify themselves.\
    2.) This captcha system uses an algorithm coded by my friend HelloThere#1337 on discord to warp images and add random lines along with random pixel changes to the image, this is to make the image harder to solve with basic OCR.

# Config options
- SusLevel: how many checks the account has to fail before being served a punishment. Low is 1 max is 3.
- Punishment: the punishment served to the user on failing captcha of failing bot checks. This can be either "kick" or "ban".
- Token: The discord bot token.
- ModChannel: The channel where mod logs will be posted, if someone fails a captcha or gets detected as a bot it will show here.
- Check*: Checks for *, for example, if checkAvatar is true then it will check if they have an avatar or not.
- VerifiedRole: Role to give the user when they pass the captcha.
- CaptchaMessage: The message that the user reacts to when trying to be served a captcha.
- CaptchaTries: The amount of tries a user has to solve a captcha. 
- CaptchaReaction: The emoji which the user has to react when trying to be served a captcha.
- AccountAgeMS: The minimum age that an account must be before it wont be flagged as a bot. This is Unix Millisecond.
- BotLimit: The amount of users that can join with the same name in *JoinTime MS before they all get detected as a bot and punished.
- JoinTime: The amount of time in MilliSeconds it takes after join to wipe the *Config.PreviousJoined users.
- DetectRaids: Boolean option that chooses rather the bot will punish a user because of another user being punished with their username and/or the user will be punished for an influx of users joining with their name, this can be configured with the BotLimit and JoinTime option.

# Notice
**Please use this with membership screening on with Email and Phone verification required.**\
If you want to add more emojis, they must be 60x60 and contain more than 2 colors.
