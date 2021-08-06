# server-guard
 A discord server guard which uses emojis and bot detection (WIP)

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

# Notice
This is a WIP meaning not everything is done and updates will come in the future.\
I plan on adding checks for flood when members join and when reactions are added.\
**Please use this with membership screening on with Email and Phone verification required.**\
If you want to add more emojis, they must be 60x60 or it will look weird.