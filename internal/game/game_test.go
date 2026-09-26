package game


import (
	"strings"
	"testing"
)


// ------------------ BUILD MASKTESTS -------------------------
func TestBuildMask_SingleWork(t *testing.T){

	// input -> "cat"   output -> "_ _ _"
	result := buildMask("cat")

	if result != "_ _ _"{
		t.Errorf("expected _ _ _ got %s", result)
	}
}


func TestBiuldMask_PreservesSpaces(t *testing.T){
	// hot dog -> "_ _ _  _ _ _"
	result := buildMask("hot dog") 
	if result != "_ _ _   _ _ _"{
		t.Errorf("Error expected '_ _ _   _ _ _' , got %s", result)
	}
}



func TestBuildMask_SingleWord(t *testing.T){
	// "s" -> "_"
	var words = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, char := range words{
		result := buildMask(string(char))
		if result != "_"{
			t.Errorf("Error expected '_' but got %s ", result)
		}

	}
}


func TestBuildMask_LongerWord(t *testing.T) {
	// "elephant" has 8 letters → 8 underscores separated by spaces
	result := buildMask("elephant")
	parts := strings.Split(result, " ")
	if len(parts) != 8 {
		t.Errorf("expected 8 parts, got %d — full mask: '%s'", len(parts), result)
	}
	for _, p := range parts {
		if p != "_" {
			t.Errorf("every part should be '_', got '%s'", p)
		}
	}
}

func TestBuildMask_DoesNotRevealLetters(t *testing.T) {
	word := "watermelon"
	result := buildMask(word)
	for _, ch := range word {
		if strings.ContainsRune(result, ch) {
			t.Errorf("mask revealed the letter '%c', full mask: '%s'", ch, result)
		}
	}
}

// ─── GenerateRoomId tests ────────────────────────────────────────────────────

func TestGenerateRoomId_Length(t *testing.T) {
	id := GenerateRoomId()
	if len(id) != 5 {
		t.Errorf("expected room ID of length 5, got %d: '%s'", len(id), id)
	}
}

func TestGenerateRoomId_OnlyValidChars(t *testing.T) {
	const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for range 100 { // generate 100 IDs and check all of them
		id := GenerateRoomId()
		for _, ch := range id {
			if !strings.ContainsRune(validChars, ch) {
				t.Errorf("invalid character '%c' in room ID '%s'", ch, id)
			}
		}
	}
}

func TestGenerateRoomId_Uniqueness(t *testing.T) {
	// not guaranteed mathematically, but 1000 IDs should never collide
	seen := make(map[string]bool)
	for range 1000 {
		id := GenerateRoomId()
		if seen[id] {
			t.Errorf("duplicate room ID generated: '%s'", id)
		}
		seen[id] = true
	}
}

// GameManager tests ───────────────────────────────────────────────────────

func TestNewGameManager_StartsEmpty(t *testing.T) {
	gm := NewGameManager()
	if len(gm.Rooms) != 0 {
		t.Errorf("new manager should have no rooms, got %d", len(gm.Rooms))
	}
}

func TestGetRoom_ReturnsErrorWhenNotFound(t *testing.T) {
	gm := NewGameManager()
	_, err := gm.GetRoom("XXXXX")
	if err == nil {
		t.Error("expected error for non-existent room, got nil")
	}
}

func TestCreateRoom_CanBeRetrieved(t *testing.T) {
	gm := NewGameManager()
	gm.CreateRoom("ABC01", GameTypeClassic, "player-1")

	room, err := gm.GetRoom("ABC01")
	if err != nil {
		t.Fatalf("expected to find room, got error: %v", err)
	}
	if room.State.ID != "ABC01" {
		t.Errorf("expected room ID 'ABC01', got '%s'", room.State.ID)
	}
	if room.State.Status != StatusWaiting {
		t.Errorf("new room should be waiting, got '%s'", room.State.Status)
	}
	if room.State.Host != "player-1" {
		t.Errorf("expected host 'player-1', got '%s'", room.State.Host)
	}
}

func TestDeleteRoom_RemovesIt(t *testing.T) {
	gm := NewGameManager()
	gm.CreateRoom("DEL01", GameTypeClassic, "player-1")
	gm.DeleteRoom("DEL01")

	_, err := gm.GetRoom("DEL01")
	if err == nil {
		t.Error("expected error after deleting room, got nil")
	}
}

func TestCreateRoom_MultipleRooms(t *testing.T) {
	gm := NewGameManager()
	gm.CreateRoom("R0001", GameTypeClassic, "p1")
	gm.CreateRoom("R0002", GameTypeClassic, "p2")
	gm.CreateRoom("R0003", GameTypeClassic, "p3")

	if len(gm.Rooms) != 3 {
		t.Errorf("expected 3 rooms, got %d", len(gm.Rooms))
	}
}

//  NewRoom tests ───────────────────────────────────────────────────────────

func TestNewRoom_InitialState(t *testing.T) {
	gm := NewGameManager()
	room := NewRoom("ROOM1", GameTypeClassic, "host-player", gm)

	if room.State.Status != StatusWaiting {
		t.Errorf("room should start as waiting, got '%s'", room.State.Status)
	}
	if room.State.Players == nil {
		t.Error("players map should be initialized, got nil")
	}
	if room.State.Scores == nil {
		t.Error("scores map should be initialized, got nil")
	}
	if room.State.GuessedBy == nil {
		t.Error("guessedBy map should be initialized, got nil")
	}
	if room.Broadcast == nil {
		t.Error("broadcast channel should be initialized")
	}
	if room.Guess == nil {
		t.Error("guess channel should be initialized")
	}
}