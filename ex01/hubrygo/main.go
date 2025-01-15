package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type ChatRoom struct {
	Name     string     `json:"name"`
	Messages []*Message `json:"messages"`
}

var chatRooms = make(map[string]*ChatRoom)

func getMethod(w http.ResponseWriter, r *http.Request) {
	pageName := r.URL.Path[1:]
	if pageName == "" {
		pageName = "home"
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
	var availableChatRooms []string
	for roomName := range chatRooms {
		availableChatRooms = append(availableChatRooms, roomName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableChatRooms)
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Name string `json:"name"`
	}

	// Décoder le corps de la requête JSON
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	roomName := requestData.Name
	if _, exists := chatRooms[roomName]; !exists {
		chatRooms[roomName] = &ChatRoom{Name: roomName, Messages: []*Message{}}
		fmt.Printf("Room created: %v\n", roomName)
	}
	sendAllChatRoom(w, r)
}

func sendOldMessages(w http.ResponseWriter, r *http.Request) {
	fmt.Println("sendOldMessages called") // Ajoutez cette ligne pour vérifier si la fonction est appelée

	var requestData struct {
		RoomName string `json:"roomName"`
	}

	// Décoder le corps de la requête JSON
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	roomName := requestData.RoomName
	fmt.Printf("roomName: %v\n", roomName)
	if room, exists := chatRooms[roomName]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room.Messages)
	} else {
		http.Error(w, "Room not found", http.StatusNotFound)
	}
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		RoomName string `json:"roomName"`
		Author   string `json:"author"`
		Content  string `json:"content"`
	}

	// Décoder le corps de la requête JSON
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Ajouter des instructions de débogage
	fmt.Printf("Received message for room: %v\n", requestData.RoomName)
	fmt.Printf("Author: %v\n", requestData.Author)
	fmt.Printf("Content: %v\n", requestData.Content)

	roomName := requestData.RoomName
	if room, exists := chatRooms[roomName]; exists {
		room.Messages = append(room.Messages, &Message{Author: requestData.Author, Content: requestData.Content})
		fmt.Printf("Message added: %v to room %v\n", room.Messages[len(room.Messages)-1], roomName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "Message sent successfully"})
	} else {
		http.Error(w, "Room not found", http.StatusNotFound)
	}
}

func postMethod(w http.ResponseWriter, r *http.Request) {
	pageName := r.URL.Path[1:]
	fmt.Printf("pageName: %v\n", pageName)
	switch pageName {
	case "chatRoom":
		sendAllChatRoom(w, r)
		return
	case "room":
		createRoom(w, r)
		return
	case "joinChatRoom":
		// TODO
	case "getOldMessages":
		fmt.Printf("getOldMessages\n")
		sendOldMessages(w, r)
	case "sendMessage":
		fmt.Printf("sendMessage\n")
		sendMessage(w, r)
	default:
		http.Error(w, "No Page for this", http.StatusMethodNotAllowed)
	}
}

func selectMethod(w http.ResponseWriter, r *http.Request) {
	fmt.Print("r.Method: ", r.Method, "\n")
	switch r.Method {
	case "GET":
		getMethod(w, r)
		return
	case "POST":
		postMethod(w, r)
		return
	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func main() {
	// Route pour servir les fichiers HTML
	http.HandleFunc("/", selectMethod)

	// Lancer le serveur
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
