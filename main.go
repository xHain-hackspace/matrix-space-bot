package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const (
	envHomeserver  = "SPACE_BOT_HOMESERVER"
	envPantalaimon = "SPACE_BOT_PANTALAIMON"
	envUsername    = "SPACE_BOT_USERNAME"
	envPassword    = "SPACE_BOT_PASSWORD"
	envSpaceID     = "SPACE_BOT_SPACE_ID"
	envRoomID      = "SPACE_BOT_ROOM_ID"
)

const (
	commandPrefix = "!space-bot"
)

func main() {
	homeserver := os.Getenv(envHomeserver)
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	roomID := os.Getenv(envRoomID)
	spaceID := os.Getenv(envSpaceID)

	if homeserver == "" {
		log.Fatalf("%s is undefined", envHomeserver)
	}
	if username == "" {
		log.Fatalf("%s is undefined", envUsername)
	}
	if password == "" {
		log.Fatalf("%s is undefined", envPassword)
	}
	if spaceID == "" || !strings.HasPrefix(spaceID, "!") {
		log.Fatalf("%s is undefined, or wrong format", envSpaceID)
	}
	if roomID == "" || !(strings.HasPrefix(roomID, "!") || strings.HasPrefix(roomID, "#")) {
		log.Fatalf("%s is undefined, or wrong format", envRoomID)
	}
	fmt.Println("Logging into", homeserver, "as", username)
	client, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		panic(err)
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: username},
		Password:         password,
		StoreCredentials: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Login successful")

	space := id.RoomID(spaceID)
	client.JoinRoomByID(space)
	// resolve roomid if alias
	var room id.RoomID
	if strings.HasPrefix(roomID, "!") {
		room = id.RoomID(roomID)
	} else {
		resp, err := client.ResolveAlias(id.RoomAlias(roomID))
		if err != nil {
			log.Fatalf("Error: Could not find bot room: %s\n", err)
		}
		room = resp.RoomID
	}
	client.JoinRoomByID(room)

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	ignore := mautrix.OldEventIgnorer{
		UserID: client.UserID,
	}
	ignore.Register(syncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		// ignore messages from us
		if evt.Sender == client.UserID {
			return
		}
		var respondSuccess = func() {
			if _, err := client.SendReaction(room, evt.ID, "✅"); err != nil {
				log.Printf("Error: could not send positive reaction %s\n", err)
			}
		}
		var respondError = func(msg string) {
			if _, err := client.SendReaction(room, evt.ID, "❌"); err != nil {
				log.Printf("Error: could not send negative reaction: %s\n", err)
			}
			content := struct {
				Body    string `json:"body"`
				MsgType string `json:"msgtype"`
			}{
				Body:    msg,
				MsgType: "m.text",
			}
			if _, err := client.SendMessageEvent(room, event.EventMessage, &content); err != nil {
				log.Printf("Error: could not send response: %s\n", err)
				return
			}
		}
		body := evt.Content.AsMessage().Body
		msg := strings.Split(body, " ")
		if !strings.HasPrefix(body, commandPrefix) || len(msg) != 3 {
			respondError("Not a valid command: " + body)
			log.Printf("Not a valid command: %s\n", body)
			return
		}
		cmd := msg[1]
		switch cmd {
		case "add":
			if msg[2] == "" {
				log.Println("Error: no room specified")
				respondError("No room specified")
				return
			}
			content := struct {
				Via []string `json:"via"`
			}{
				Via: []string{homeserver},
			}
			addedRoomId := msg[2]
			if strings.HasPrefix(addedRoomId, "#") {
				rsp, err := client.ResolveAlias(id.RoomAlias(addedRoomId))
				if err != nil {
					respondError("Could not find room: " + addedRoomId)
					return
				}
				addedRoomId = rsp.RoomID.String()
			}
			_, err := client.SendStateEvent(space, event.StateSpaceChild, addedRoomId, content)
			if err != nil {
				log.Printf("Error: Could not update to space room: %s\n", err)
				respondError("Could not update room " + addedRoomId + " to be part of space " + space.String())
			} else {
				respondSuccess()
			}

		default:

		}
	})

	err = client.Sync()
	if err != nil {
		panic(err)
	}

}
