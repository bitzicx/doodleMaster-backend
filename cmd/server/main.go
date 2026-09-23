package main

import (
	"log"
	"net/http"
	"time"

	"encoding/json"

	"doodleMaster/internal/game"

	"github.com/gorilla/websocket"

	"os"
)

// global manager
var manager = game.NewGameManager()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow every one to connect
	},
}



type CreateRoomRequest struct {
	Username string `json:"username"`
	PlayerID string `json:"playerId"`
}

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRoomRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var roomID string

	for {
		roomID = game.GenerateRoomId()
		_, err := manager.GetRoom(roomID)
		if err != nil {
			break
		}
	}

	// create room otherwise
	manager.CreateRoom(roomID, game.GameTypeClassic, req.PlayerID)
	log.Printf("Room %s created successfully created by %s \n", roomID, req.Username)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(roomID))
}


type JoinRoomRequest struct {
	RoomID   string `json:"roomId"`
	Username string `json:"username"`
	PlayerID string `json:"playerId"`
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JoinRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := manager.GetRoom(req.RoomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Room found"))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	username := r.URL.Query().Get("username")
	playerID := r.URL.Query().Get("playerId")

	if roomID == "" || username == "" || playerID == "" {
		log.Println("Connection rejected: missing roomId, username, or playerId")
		return
	}

	room, err := manager.GetRoom(roomID)
	if err != nil {
		log.Println("Connection rejected: room doesn't exist")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v\n", err)
		return
	}

	log.Printf("Player %s connected to room %s\n", username, roomID)

	player := &game.Player{
		ID:       playerID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	room.Join <- player
	go writePump(player)
	readPump(player, room) // blocks until disconnect
}


func handleQuickJoin(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return 
	}
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return 
	}

	roomId := manager.FindOrCreateRoom(req.PlayerID)
	log.Printf("Quick Join: Player %s in room %s", req.Username, roomId)
	w.WriteHeader((http.StatusOK))
	w.Write([]byte(roomId))
}



// read from personal inbox and send to phone
func writePump(player *game.Player) {
	defer player.Conn.Close()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select{
		case message, ok := <- player.Send:
			if !ok{
				player.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return 
			}
			player.Conn.WriteMessage(websocket.TextMessage, message)
		case <- ticker.C:
			if err :=  player.Conn.WriteMessage(websocket.PingMessage, nil); err != nil{
				log.Printf("Ping failed for player %s: %v", player.ID, err)
				return 
			} 
		}
	}
}


// listen from phone and Broadcast to room member
func readPump(player *game.Player, room *game.Room) {
	defer func() {
		room.Leave <- player
		player.Conn.Close()
	}()

	player.Conn.SetReadDeadline(time.Now().Add(60*time.Second))

	player.Conn.SetPongHandler(func(string) error{
		player.Conn.SetReadDeadline(time.Now().Add(60*time.Second))
		log.Printf("Pong receivef rom player %s", player.ID)
		return nil 
	})

	for {
		_, message, err := player.Conn.ReadMessage()
		if err != nil {
			log.Printf("player %s disconnected: %v", player.ID, err)
			break
		}

		var event game.GameEvent

		// decoding
		err = json.Unmarshal(message, &event)

		if err != nil {
			log.Println("Eroor, Inavalied massige: ", string(message))
			return
		}
		event.SenderID = player.ID

		enrichedMessage, err := json.Marshal(event)
		if err != nil {
			log.Print("Error marshaling event: ", err)
			return
		}

		log.Printf("player %s just trigger %s event", player.Username, event.Type)

		if event.Type == "START_GAME" {
			room.State.Status = "RUNNING"
			room.BroadcastState()
			go room.StartGameLoop()
			log.Printf("Room %s Current Status is now: %s", room.State.ID, room.State.Status)

			continue
		}

		if event.Type == "CHAT"{
			payloadMap, ok := event.Payload.(map[string]interface{})
			if !ok {
				log.Println("CHAT payload is not an object")
				continue
			}

			text, ok := payloadMap["message"].(string)
			if !ok || text == "" {
				log.Println("CHAT payload missing text field")
				continue
			}

			log.Printf("player %s guessed: %s", player.ID, text)

			if room.State.Status == "RUNNING" {
				select {
				case room.Guess <- game.GuessEvent{PlayerID: player.ID, Text: text}:
				default:
				}
				
				continue
			}

			
		}

		room.Broadcast <- game.BroadcastData{
			EventType: string(event.Type),
			SenderID: player.ID,
			Payload:  enrichedMessage,
		}
	}

} 

func main(){

	port := os.Getenv("PORT")
	if port == ""{
		port = "8080"
	}

	// endpoint for continuous pings
	http.HandleFunc("/health", func(w http.ResponseWriter, r * http.Request){
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})


	http.HandleFunc("/create", handleCreateRoom)
	http.HandleFunc("/quickjoin", handleQuickJoin)
	http.HandleFunc("/join", handleJoinRoom)
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("server running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v\n", err)
	}

}
