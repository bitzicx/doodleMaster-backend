package game



import(
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"time"
    "fmt" 
)



func (r *Room)StartGameLoop(){
    
	// time.Sleep(3*time.Second)

    r.State.TurnOrder = make([]string, 0, len(r.State.Players))
    for id := range r.State.Players {
        r.State.TurnOrder = append(r.State.TurnOrder, id)
		r.State.Scores[id] = 0 // initial socre
	}
 
    for j := range 3{
        for i , playerId := range r.State.TurnOrder{
            drawer, ok := r.State.Players[playerId]

            if !ok {
                continue 
            }
            r.State.Round = i+1
            r.runTurn(drawer)
            time.Sleep(2*time.Second)
        }
        j++
    }

	r.State.Status = StatusFinished
	r.BroadcastState()
}


func (r *Room) runTurn(drawer *Player) {
    // pick word
    word := doodles[rand.Intn(len(doodles))].Name
    mask := buildMask(word)

    // update state
    r.State.CurrentTurn = drawer.ID
    r.State.CurrentWord = word
    r.State.WordMask = mask
    r.State.GuessedBy = make(map[string]bool)

    r.broadcastTurnStart(drawer, word, mask)
    r.BroadcastState()

    time.Sleep(2*time.Second)


    for len(r.Guess) > 0 {
        <-r.Guess
    }

    const turnSeconds = 40

    turnDuration := 40 * time.Second
    timer := time.NewTimer(turnDuration)
    ticker := time.NewTicker(1*time.Second)

    defer timer.Stop()
    defer ticker.Stop()

    nonDrawers := len(r.State.Players) - 1
    remaining := turnSeconds

    for {
        select {
        case <-timer.C:
            // time's up
            r.broadcastTurnEnd(word)
            return

        case <- ticker.C:
            remaining--
            r.broadcastTimer(remaining)
            if _, drawerStillHere := r.State.Players[drawer.ID]; !drawerStillHere {
                log.Printf("Drawer %s left — ending turn early", drawer.Username)
                r.broadcastTurnEnd(word)
                return
            }

        case guess := <-r.Guess:
            // ignore the drawer guessing their own word
            if guess.PlayerID == drawer.ID {
                continue
            }


            log.Printf("Player %s guessed %s word", guess.PlayerID, guess.Text)

            if strings.EqualFold(strings.TrimSpace(guess.Text), word) {

            // ignore players who already guessed correctly
                if r.State.GuessedBy[guess.PlayerID] {
                    continue
                }
			    log.Printf("Correctly guessed by : %s", guess.PlayerID)
                
                // correct guess
                r.State.GuessedBy[guess.PlayerID] = true

                score := 1 // you can make this time-based later
                r.State.Scores[guess.PlayerID] += score
                r.State.Scores[drawer.ID] += 1 // drawer earns per correct guess

                r.broadcastCorrectGuess(guess.PlayerID)
                // broadcast the updated state so everyone sees scores tick up
                r.BroadcastState()



                // end turn early if everyone guessed
                if len(r.State.GuessedBy) >= nonDrawers {
                    timer.Stop()
                    r.broadcastTurnEnd(word)
                    return
                }
            }else{

                var event GameEvent
                event.Type = "CHAT"
                event.SenderID = guess.PlayerID
                event.Payload = map[string]string{
                        "message":   fmt.Sprintf("%s", guess.Text),
                    }

                b, _ := json.Marshal(event)
                r.Broadcast <- BroadcastData{
                    EventType: "CHAT",
                    SenderID:  "SERVER",
                    Payload:   b,
                } 
            }
        }
    }
}

func (r *Room) broadcastCorrectGuess(playerID string) {
    player, ok := r.State.Players[playerID]
    if !ok {
        return
    }

    var event GameEvent
    event.Type = "CHAT"
    event.SenderID = "SERVER"
    event.Payload = map[string]string{
            "message":   fmt.Sprintf("%s guessed the word!", player.Username),
        }

    b, _ := json.Marshal(event)
    r.Broadcast <- BroadcastData{
        EventType: "CORRECT_GUESS",
        SenderID:  "SERVER",
        Payload:   b,
    }
}

func buildMask(word string) string {
    parts := strings.Split(word, "")
    for i, ch := range parts {
        if ch != " " {
            parts[i] = "_"
        }
    }
    return strings.Join(parts, " ")
}


func (r *Room) broadcastTurnStart(drawer *Player, word, mask string) {
    for _, p := range r.State.Players {
        var payload TurnStartPayload
        if p.ID == drawer.ID {
            payload = TurnStartPayload{
                DrawerID:   drawer.ID,
                DrawerName: drawer.Username,
                Word:       word, // only the drawer gets this
                WordMask:   mask,
                Duration:   40,
            }
        } else {
            payload = TurnStartPayload{
                DrawerID:   drawer.ID,
                DrawerName: drawer.Username,
                Word:       "",  // guessers never see the word
                WordMask:   mask,
                Duration:   40,
            }
        }

        event := GameEvent{Type: EventTurnStart, Payload: payload}
        b, err := json.Marshal(event)
        if err != nil {
            continue
        }
        select {
        case p.Send <- b:
        default:
        }
    }
}

func (r *Room) broadcastTurnEnd(word string) {
    event := GameEvent{
        Type: EventTurnEnd,
        Payload: TurnEndPayload{
            Word:      word,
            Scores:    r.State.Scores,
            GuessedBy: r.State.GuessedBy,
        },
    }
    b, _ := json.Marshal(event)
    r.Broadcast <- BroadcastData{
        EventType: string(EventTurnEnd),
        SenderID:  "SERVER",
        Payload:   b,
    }
}

func (r *Room) broadcastGameOver() {
    event := GameEvent{
        Type:    EventGameOver,
        Payload: r.State.Scores,
    }
    b, _ := json.Marshal(event)
    r.Broadcast <- BroadcastData{
        EventType: string(EventGameOver),
        SenderID:  "SERVER",
        Payload:   b,
    }
    log.Printf("Room %s game over. Scores: %v", r.State.ID, r.State.Scores)
}


func (r *Room) broadcastTimer(remaining int) {
    event := GameEvent{
        Type:    "TIMER",
        Payload: map[string]int{"remaining": remaining},
    }
    b, _ := json.Marshal(event)
    r.Broadcast <- BroadcastData{
        EventType: "TIMER",
        SenderID:  "SERVER",
        Payload:   b,
    }
}