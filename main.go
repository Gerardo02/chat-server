package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type clientMessage struct {
	UserName     string `json:"user_name"`
	Message      string `json:"message"`
	FirstMessage bool   `json:"first_message"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	clientsConns = make(map[string]*websocket.Conn)
)

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
		_, message, err := conn.ReadMessage()
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

		// Send ready and listening message to client
		log.Printf("Received: %s\n", messageBody.Message)
		if messageBody.FirstMessage {
			expectedUserName = messageBody.UserName
			response, _ := json.Marshal(clientMessage{
				UserName:     "Server",
				Message:      "Welcome to the chat room, " + expectedUserName,
				FirstMessage: false,
			})

			log.Println("Received ready from client")
			if err := conn.WriteMessage(1, response); err != nil {
				log.Println("Error writing message:", err)
			}

			broadcast(clientsConns, clientMessage{
				UserName:     "Server",
				Message:      expectedUserName + " just joined the chat, say hi :)",
				FirstMessage: false,
			}, expectedUserName)

			clientsConns[expectedUserName] = conn
			continue
		}

		// Respond with latest message to everyone but the author of the message
		broadcast(clientsConns, messageBody, expectedUserName)
	}

	delete(clientsConns, expectedUserName)
	broadcast(clientsConns, clientMessage{
		UserName:     "Server",
		Message:      expectedUserName + " exited the chat room :(",
		FirstMessage: false,
	}, expectedUserName)
}

func broadcast(conns map[string]*websocket.Conn, message clientMessage, sender string) {
	for k, v := range conns {
		if k == sender {
			continue
		}

		response, _ := json.Marshal(message)
		if err := v.WriteMessage(1, response); err != nil {
			log.Printf("Error writing message to: %s\n", k)
			log.Println(err)
			continue
		}
	}
}
