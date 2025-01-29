package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Message struct {
	Author   string   `json:"author"`
	Content  string   `json:"content"`
	Time     string   `json:"time"`
	Image    string   `json:"imageId"`
	RoomId   string   `json:"roomId"`
	TabImage []string `json:"tabImage"`
}

type ChatRoom struct {
	Name         string     `json:"name"`
	Key          string     `json:"key"`
	Messages     []*Message `json:"messages"`
	LastActivity time.Time  `json:"lastActivity"`
}

var chatRooms = make(map[string]*ChatRoom)
var clients = make(map[chan *Message]bool)
var roomClients = make(map[chan *ChatRoom]bool)
var clientsMutex sync.Mutex
var chatRoomsMutex sync.Mutex

func getMethod(w http.ResponseWriter, r *http.Request) {
	pageName := r.URL.Path[1:]
	if pageName == "" {
		pageName = "index"
	}
	pageName = "template/" + pageName + ".html"
	file, error := os.Open(pageName)
	if error != nil {
		fmt.Printf("404 not found\n")
		file, error = os.Open("template/404.html")
		if error != nil {
			os.Exit(1)
		}
	}
	fileInfo, error := file.Stat()
	if error != nil {
		os.Exit(1)
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func sendAllChatRoom(w http.ResponseWriter, _ *http.Request) {
	type RoomInfo struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}

	var availableChatRooms []RoomInfo
	for _, room := range chatRooms {
		availableChatRooms = append(availableChatRooms, RoomInfo{Name: room.Name, Key: room.Key})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableChatRooms)
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	roomName := requestData.Name
	key := requestData.Key
	if _, exists := chatRooms[key]; !exists {
		chatRooms[key] = &ChatRoom{
			Name:         roomName,
			Key:          key,
			Messages:     []*Message{},
			LastActivity: time.Now(),
		}
		// fmt.Printf("Room created: %v\n", chatRooms[key])
		notifyRoomClients(chatRooms[key])
	}
	// creer un dossier pour les images
	os.Mkdir("images/"+key, 0777)
	sendAllChatRoom(w, r)
}

func sendOldMessages(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("sendOldMessages called")
	var requestData struct {
		RoomKey string `json:"roomKey"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	// fmt.Printf("Request body: %s\n", body)

	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// fmt.Printf("Received request for old messages for room: %v\n", requestData)
	roomKey := requestData.RoomKey
	// fmt.Printf("roomKey: %v\n", roomKey)
	if room, exists := chatRooms[roomKey]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.Messages)
	} else {
		http.Error(w, "Room not found", http.StatusNotFound)
	}
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("sendMessage called\n")
	var requestData struct {
		Author   string   `json:"author"`
		Content  string   `json:"content"`
		Image    string   `json:"imageId"`
		RoomId   string   `json:"roomId"`
		ImageTab []string `json:"tabImage"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Request body: %s\n", body)

	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	fmt.Printf("Received message: %v\n", requestData)

	roomKey := requestData.RoomId
	if room, exists := chatRooms[roomKey]; exists {
		message := &Message{
			Author:   requestData.Author,
			Content:  requestData.Content,
			Time:     time.Now().Format(time.RFC3339),
			Image:    requestData.Image,
			RoomId:   roomKey,
			TabImage: requestData.ImageTab,
		}
		room.Messages = append(room.Messages, message)
		room.LastActivity = time.Now()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "Message sent successfully"})
		notifyClients(message)
	} else {
		http.Error(w, "Room not found", http.StatusNotFound)
	}
}

func notifyClients(message *Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		client <- message
	}
}

func notifyRoomClients(room *ChatRoom) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range roomClients {
		client <- room
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("sseHandler called\n")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	messageChan := make(chan *Message)
	clientsMutex.Lock()
	clients[messageChan] = true
	clientsMutex.Unlock()
	defer func() {
		clientsMutex.Lock()
		delete(clients, messageChan)
		clientsMutex.Unlock()
		close(messageChan)
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for {
		select {
		case message := <-messageChan:
			data, _ := json.Marshal(message)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func roomSseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	roomChan := make(chan *ChatRoom)
	clientsMutex.Lock()
	roomClients[roomChan] = true
	clientsMutex.Unlock()
	defer func() {
		clientsMutex.Lock()
		delete(roomClients, roomChan)
		clientsMutex.Unlock()
		close(roomChan)
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for {
		select {
		case room := <-roomChan:
			data, _ := json.Marshal(room)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func clearRoom() {
	for {
		fmt.Println("Running clearRoom")
		time.Sleep(1 * time.Minute)
		chatRoomsMutex.Lock()
		for roomKey, room := range chatRooms {
			if time.Since(room.LastActivity) > 10*time.Minute {
				delete(chatRooms, roomKey)
				fmt.Printf("Room %s cleared due to inactivity\n", roomKey)
			}
		}
		chatRoomsMutex.Unlock()
	}
}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("joinRoom called\n")
	roomKey := r.URL.Path[len("/rooms/"):]
	_, exists := chatRooms[roomKey]
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	pageName := "template/chat.html"
	file, error := os.Open(pageName)
	if error != nil {
		fmt.Printf("404 not found\n")
		file, error = os.Open("template/404.html")
		if error != nil {
			os.Exit(1)
		}
	}
	fileInfo, error := file.Stat()
	if error != nil {
		os.Exit(1)
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("uploadFile called")
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error retrieving the file")
		fmt.Println(err)
		return
	}
	roomId := "images/" + r.FormValue("roomId") + "/"
	defer file.Close()
	fmt.Printf("Uploaded file: %+v\n", handler.Filename)
	f, err := os.OpenFile(roomId+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error creating the file")
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	w.Header().Set("Content-Type", "application/json")
	imagePath := roomId + handler.Filename
	json.NewEncoder(w).Encode(map[string]string{"status": "File uploaded successfully", "imagePath": imagePath})
}

func main() {
	go clearRoom()
	http.HandleFunc("/", getMethod)
	http.HandleFunc("GET /rooms/{id}/messages", sseHandler)
	http.HandleFunc("/roomEvents", roomSseHandler)
	http.HandleFunc("GET /rooms/{id}", joinRoom)
	http.HandleFunc("POST /room", createRoom)
	http.HandleFunc("POST /rooms/{id}/messages", sendMessage)
	http.HandleFunc("POST /getOldMessages", sendOldMessages)
	http.HandleFunc("GET /chatRoom", sendAllChatRoom)
	http.HandleFunc("POST /upload", uploadFile)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	fmt.Println("Serveur démarré sur : https://localhost:8081")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()
	err := http.ListenAndServeTLS(":8081", "cert.csr", "cert.key", nil)
	if err != nil {
		fmt.Println(err)
	}
}

// Generate private key (.key)

// # Key considerations for algorithm "RSA" ≥ 2048-bit
// openssl genrsa -out server.key 2048

// # Key considerations for algorithm "ECDSA" ≥ secp384r1
// # List ECDSA the supported curves (openssl ecparam -list_curves)
// openssl ecparam -genkey -name secp384r1 -out server.key

// Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)

// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
