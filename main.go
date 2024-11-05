package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type clientMessage struct {
	UserName string `json:"user_name"`
	Message  string `json:"message"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// var messages []clientMessage

func main() {
	router := chi.NewRouter()

	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("estamos listos compas"))
	})

	router.Get("/ws/chat", chatHandler)

	serve := &http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	log.Println("listening http://localhost:8080")
	serve.ListenAndServe()
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP request to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer conn.Close() // Ensure the connection is closed when the function ends

	var expectedUserName string

	// Listen for messages from the client and respond
	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		messageBody := clientMessage{}

		err = json.Unmarshal(message, &messageBody)
		if err != nil {
			log.Println("No pudimos decifrar el json de tu mensaje compadre")
			break
		}

		log.Printf("Received: %s\n", messageBody.Message)
		if messageBody.Message == "READY" {
			expectedUserName = messageBody.UserName
			log.Println("Received ready from client")
			if err := conn.WriteMessage(1, []byte("Welcome to the chat room, "+expectedUserName)); err != nil {
				log.Println("Error writing message:", err)
			}
			continue
		}

		// Store incomming messages
		// messages = append(messages, messageBody)

		// Respond with a confirmation message
		if messageBody.UserName != expectedUserName {
			response, _ := json.Marshal(messageBody)
			if err := conn.WriteMessage(messageType, response); err != nil {
				log.Println("Error writing message:", err)
				break
			}
		}
	}
}
