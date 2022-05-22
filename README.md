# Matrix Space Bot

## Compiling

The bot can be built with the command `go build main.go` from the repos root.

## Usage

You have to set a couple of Environment Variables to make the bot work:

| Variable               | Description                                                       |
| ---------------------- | ----------------------------------------------------------------- |
| `SPACE_BOT_HOMESERVER` | URL in the form of `matrix.example.org`                           |
| `SPACE_BOT_USERNAME`   | Username of the bot user                                          |
| `SPACE_BOT_PASSWORD`   | Password of the bot user                                          |
| `SPACE_BOT_SPACE_ID`   | ID of the Space this bot should manage, either as roomID or alias |
| `SPACE_BOT_ROOM_ID`    | ID of the Room that will be used to send commands to the bot      |
