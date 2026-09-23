package game

import (
	"encoding/json"
	"log"
)

type GuessEvent struct{
	PlayerID string
	Text string
}

type Room struct {
	State BaseRoom

	Join      chan *Player
	Leave     chan *Player
	Broadcast chan BroadcastData
	Guess     chan GuessEvent
	manager   *GameManager
	done      chan struct{}
}

func (r *Room) BroadcastState() {
	event := GameEvent{
		Type:    EventState,
		Payload: r.State,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Println("Error crushing JSON: ", err)
		return
	}

	r.Broadcast <- BroadcastData{
		SenderID: "SERVER",
		Payload:  eventBytes,
	}
}

func NewRoom(id string, gameType GameType, host string,manager *GameManager) *Room {
	return &Room{
		State: BaseRoom{
			ID:      id,
			Type:    gameType,
			Host:    host,
			Status:  StatusWaiting,
			Players: make(map[string]*Player),
			GuessedBy: make(map[string]bool),
			Scores: make(map[string]int),
		},
		Join:      make(chan *Player),
		Leave:     make(chan *Player),
		Broadcast: make(chan BroadcastData, 256),
		Guess: 	   make(chan GuessEvent, 64),
		manager:   manager,
		done:      make(chan struct{}),
	}
}

func (r *Room) Run() {

	log.Printf("Room %s is now running\n", r.State.ID)

	for {
		select {
		// player joins
		case player := <-r.Join:
			if existing, ok := r.State.Players[player.ID]; ok{
				log.Printf("Player %s rejoining - closing old connection", player.Username)
				close((existing.Send)) // killing old write pump 

			}

			r.State.Players[player.ID] = player
			r.State.Scores[player.ID] = 0

			//  game running and they're  new not rejoining
			if r.State.Status == StatusStarted {
				alreadyInOrder := false
				for _, id := range r.State.TurnOrder {
					if id == player.ID {
						alreadyInOrder = true
						break
					}
				}

				if !alreadyInOrder {
					r.State.TurnOrder = append(r.State.TurnOrder, player.ID)
					log.Printf("Added %s to turn order", player.Username)
				}
			}
			
			r.BroadcastState()

		case player := <-r.Leave:
			existing, ok := r.State.Players[player.ID]
			if !ok{
				break // player already removed 
			}

			if existing.Conn != player.Conn{
				log.Printf("Ignoring leave for %s - newer connection exists", player.Username)
				break 
			}

			if _, ok := r.State.Players[player.ID]; ok {
				delete(r.State.Players, player.ID)
				close(player.Send) // close their inbox
				log.Printf("Player %s left room %s\n", player.Username, r.State.ID)
				r.BroadcastState()
			}

			if len(r.State.Players)  == 0 || r.State.Status == StatusFinished {
				r.manager.DeleteRoom(r.State.ID)
				return 
			}

		case data := <-r.Broadcast:
			for _, player := range r.State.Players {

				if player.ID == data.SenderID && data.EventType == "DRAW" {
					continue
				}

				select {
				case player.Send <- data.Payload:
				default:
					close(player.Send)
					delete(r.State.Players, player.ID)
				}
			}	
		}
	}
}
