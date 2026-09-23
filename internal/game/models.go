package game

import "github.com/gorilla/websocket"


// draw point representation 
type DrawPoint struct{
	X float32 `json:"x"`
	Y float32 `json:"y"`
}


type BroadcastData struct{
	EventType string
	SenderID string
	Payload []byte

}


// player represents a user currently connected to the server
type Player struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	Score    int             `json:"score"`
	Conn     *websocket.Conn `json:"-"`
	Send     chan []byte     `json:"-"`
}


// Guess represents a chat attempt at the word
type Guess struct {
	PlayersId string `json:"player_id"`
	GuessText string `json:"guesstext"`
	GuessedAt int64 `json:"guessed_at"`
}



type GameType string

const (
	GameTypeClassic GameType = "CLASSIC_TURN_BASED"
	GameTypeSpeedAI GameType = "SPEED_AI_SHOWDOWN"
)


type RoomStatus string

const (
    StatusWaiting  RoomStatus = "waiting"
    StatusStarted  RoomStatus = "started"
    StatusFinished RoomStatus = "finished"
)

// BaseRoom is the in-memory arena
type BaseRoom struct {
    Host        string             `json:"host"`
    ID          string             `json:"id"`
    Status      RoomStatus         `json:"status"`
    Type        GameType           `json:"type"`
    Players     map[string]*Player `json:"players"`
    CurrentTurn string             `json:"current_turn"`
    CurrentWord string             `json:"current_word"` 
    WordMask    string             `json:"word_mask"`    // "_ _ _ _" sent to guessers
    Round       int                `json:"round"`
    StartedAt   int64              `json:"started_at"`
    GuessedBy   map[string]bool    `json:"guessed_by"`  
    Scores      map[string]int     `json:"scores"`      
    TurnOrder   []string           `json:"turn_order"`  
}


type EventType string 

const (
    EventDraw      EventType = "DRAW"
    EventChat      EventType = "CHAT"
    EventState     EventType = "STATE_UPDATE"
    EventTurnStart EventType = "TURN_START"
    EventTurnEnd   EventType = "TURN_END"
    EventGameOver  EventType = "GAME_OVER"
    EventGuess     EventType = "GUESS"
)
type GameEvent struct {
	Type EventType `json:"type"`
	SenderID string    `json:"sender_id"`
	Payload any `json:"payload"`
}


type TurnStartPayload struct {
    DrawerID   string `json:"drawer_id"`
    DrawerName string `json:"drawer_name"`
    Word       string `json:"word"`       // only set for the drawer
    WordMask   string `json:"word_mask"`  // always set
    Duration   int    `json:"duration"`   // seconds
}

type GuessPayload struct {
    Text string `json:"text"`
}

type TurnEndPayload struct {
    Word      string         `json:"word"`
    Scores    map[string]int `json:"scores"`
    GuessedBy map[string]bool `json:"guessed_by"`
}



type Doodle struct {
	Name string
}

var doodles = []Doodle{
	{Name: "chair"},
	{Name: "table"},
	{Name: "bed"},
	{Name: "sofa"},
	{Name: "door"},
	{Name: "window"},
	{Name: "lamp"},
	{Name: "fan"},
	{Name: "clock"},
	{Name: "mirror"},
	{Name: "cup"},
	{Name: "mug"},
	{Name: "plate"},
	{Name: "bowl"},
	{Name: "spoon"},
	{Name: "fork"},
	{Name: "knife"},
	{Name: "bottle"},
	{Name: "glass"},
	{Name: "bucket"},
	{Name: "umbrella"},
	{Name: "backpack"},
	{Name: "pencil"},
	{Name: "pen"},
	{Name: "eraser"},
	{Name: "ruler"},
	{Name: "book"},
	{Name: "notebook"},
	{Name: "scissors"},
	{Name: "key"},
	{Name: "lock"},
	{Name: "wallet"},
	{Name: "phone"},
	{Name: "camera"},
	{Name: "laptop"},
	{Name: "television"},
	{Name: "remote"},
	{Name: "headphones"},
	{Name: "light bulb"},
	{Name: "candle"},

	{Name: "apple"},
	{Name: "banana"},
	{Name: "orange"},
	{Name: "watermelon"},
	{Name: "strawberry"},
	{Name: "cherry"},
	{Name: "grape"},
	{Name: "lemon"},
	{Name: "pineapple"},
	{Name: "coconut"},
	{Name: "carrot"},
	{Name: "potato"},
	{Name: "tomato"},
	{Name: "mushroom"},
	{Name: "corn"},
	{Name: "egg"},
	{Name: "bread"},
	{Name: "pizza"},
	{Name: "burger"},
	{Name: "hot dog"},
	{Name: "sandwich"},
	{Name: "donut"},
	{Name: "cake"},
	{Name: "ice cream"},
	{Name: "lollipop"},
	{Name: "cookie"},
	{Name: "candy"},
	{Name: "chocolate"},
	{Name: "popcorn"},
	{Name: "french fries"},
	{Name: "noodles"},
	{Name: "rice"},
	{Name: "soup"},
	{Name: "cheese"},
	{Name: "milk"},
	{Name: "juice"},
	{Name: "coffee"},
	{Name: "tea"},
	{Name: "water bottle"},
	{Name: "cupcake"},

	{Name: "cat"},
	{Name: "dog"},
	{Name: "mouse"},
	{Name: "rat"},
	{Name: "rabbit"},
	{Name: "horse"},
	{Name: "cow"},
	{Name: "pig"},
	{Name: "sheep"},
	{Name: "goat"},
	{Name: "chicken"},
	{Name: "duck"},
	{Name: "bird"},
	{Name: "fish"},
	{Name: "shark"},
	{Name: "whale"},
	{Name: "dolphin"},
	{Name: "octopus"},
	{Name: "crab"},
	{Name: "turtle"},
	{Name: "snake"},
	{Name: "frog"},
	{Name: "monkey"},
	{Name: "lion"},
	{Name: "tiger"},
	{Name: "bear"},
	{Name: "elephant"},
	{Name: "giraffe"},
	{Name: "zebra"},
	{Name: "kangaroo"},
	{Name: "penguin"},
	{Name: "owl"},
	{Name: "butterfly"},
	{Name: "bee"},
	{Name: "spider"},

	{Name: "tree"},
	{Name: "flower"},
	{Name: "rose"},
	{Name: "sun"},
	{Name: "moon"},
	{Name: "star"},
	{Name: "cloud"},
	{Name: "rainbow"},
	{Name: "mountain"},
	{Name: "volcano"},
	{Name: "river"},
	{Name: "lake"},
	{Name: "island"},
	{Name: "palm tree"},
	{Name: "cactus"},
	{Name: "leaf"},
	{Name: "grass"},
	{Name: "rock"},
	{Name: "fire"},
	{Name: "snowman"},
	{Name: "snowflake"},
	{Name: "lightning"},
	{Name: "raindrop"},
	{Name: "nest"},

	{Name: "car"},
	{Name: "bus"},
	{Name: "truck"},
	{Name: "taxi"},
	{Name: "van"},
	{Name: "bicycle"},
	{Name: "motorcycle"},
	{Name: "scooter"},
	{Name: "train"},
	{Name: "subway"},
	{Name: "airplane"},
	{Name: "helicopter"},
	{Name: "rocket"},
	{Name: "boat"},
	{Name: "ship"},
	{Name: "sailboat"},
	{Name: "hot-air balloon"},
	{Name: "skateboard"},
	{Name: "roller skate"},
	{Name: "tractor"},

	{Name: "person"},
	{Name: "baby"},
	{Name: "man"},
	{Name: "woman"},
	{Name: "boy"},
	{Name: "girl"},
	{Name: "face"},
	{Name: "eye"},
	{Name: "nose"},
	{Name: "ear"},
	{Name: "mouth"},
	{Name: "hand"},
	{Name: "foot"},
	{Name: "finger"},
	{Name: "hair"},
	{Name: "mustache"},
	{Name: "beard"},
	{Name: "hat"},
	{Name: "glasses"},
	{Name: "crown"},

	{Name: "smiley face"},
	{Name: "heart"},
	{Name: "star"},
	{Name: "ghost"},
	{Name: "robot"},
	{Name: "alien"},
	{Name: "monster"},
	{Name: "crown"},
	{Name: "treasure chest"},
	{Name: "sword"},
	{Name: "shield"},
	{Name: "magic wand"},
	{Name: "balloon"},
	{Name: "gift"},
	{Name: "bomb"},
	{Name: "dice"},
	{Name: "playing card"},
	{Name: "musical note"},
	{Name: "guitar"},
	{Name: "trophy"},
}