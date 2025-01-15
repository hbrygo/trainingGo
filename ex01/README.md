# ex01
## Webchat

## Goal

Write a webserver which allows users to communicate in chat rooms in real-time.

The server's home page will allow the user to set a username, and to either create a room or join a room by entering its ID.
When creating or joining a room, users will be redirected to another page where they can see the ID of the room and a chatbox when they will view all previous messages in the room (including those from before they joined). They will also be able to send messages themselves, which will then appear in the chat attached to their chosen username.
All messages sent by any user in the room will instantly be viewable by all other users in the room without having to reload the page.

If a room is inactive (no new messages sent) for more than the inactivity timeout defined when starting the server, the room  and its messages will automatically be deleted.

## Instructions

- The program should be written in Go.
- The program can only use Go's standard library, including the external library (`golang.org/x/*`).
- The program will be compiled with:
```sh
go build -o webchat ./cmd/webchat
```
- The program should not panic.
- To provide real-time capabilities, you will use and implement SSE(Server-Sent Events).
- The room ID will have to be a randomly generated 16 characters hexadecimal string.
- The room deletion timeout will have to be set as a flag in the [time.ParseDuration](https://pkg.go.dev/time#ParseDuration) format when launching the server:
```sh
./webchat --timeout 1h
```

## Endpoints
- Homepage `GET /`
- Create room `POST /rooms/`
- Join room `GET /rooms/:id/`
- Send message to room: `POST /rooms/:id/messages/`
- Receive real-time messages from room (SSE): `GET /rooms/:id/messages/`

## Bonus
Complete as many or as little as you want.

- When deleting a room, archive its content by saving it to a file in a format of your choice.
- When creating a room, allow users to set its name and an optional public mode. Then display all active public rooms ands their name on the homepage.
- Anything else ... be creative!
