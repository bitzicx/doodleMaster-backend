package game

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"
)

func GenerateRoomId() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, 5)

	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(result)

}

type GameManager struct{
	Rooms map[string]*Room

	mu sync.RWMutex
}

// creating newGame manager when the server stars
func NewGameManager()*GameManager{
	return &GameManager{
		Rooms: make(map[string]*Room),
	}
}

const MaxPlayersPerRoom = 8

func (manager *GameManager) FindOrCreateRoom(host string) string {
    manager.mu.Lock()
    defer manager.mu.Unlock()

    // look for a waiting room with space
    for id, room := range manager.Rooms {
        if room.State.Status == StatusWaiting &&
            len(room.State.Players) < MaxPlayersPerRoom {
            return id
        }
    }

    // no room found — create a new one
    roomID := GenerateRoomId()
    room := NewRoom(roomID, GameTypeClassic, host, manager)
    manager.Rooms[roomID] = room
    go room.Run()

    return roomID
}

func (manager *GameManager)DeleteRoom(id string){
	manager.mu.Lock()
	defer manager.mu.Unlock()
	delete(manager.Rooms, id)
	log.Printf("Room %s deleted", id)
}


// create new room function 
func (manager *GameManager) CreateRoom(id string, gameType GameType, host string)*Room{

	manager.mu.Lock() //lock before creating 
	defer manager.mu.Unlock() // automatically unlock after finishing

	room := NewRoom(id, gameType, host, manager)
	manager.Rooms[id] = room

	go room.Run()

	return room

}


func (manager *GameManager) GetRoom(id string) (*Room, error){
	
	manager.mu.RLock()  // read lock, allowing multiple people to read at once 
	defer manager.mu.RUnlock()

	room, exists := manager.Rooms[id]

	if !exists{
		return nil, errors.New("room not found")
	}
	return room, nil
}