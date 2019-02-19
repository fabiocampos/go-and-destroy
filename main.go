package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type GameResponse struct {
	Players []*Player `json:"players"`
	Shots   []*Shot   `json:"shots"`
}

type Player struct {
	ID       string `json:"id"`
	Position `json:"position"`
	Color    `json:"color"`
	Status   string `json:"status"`
}

type Position struct {
	X             int    `json:"x"`
	Y             int    `json:"y"`
	FaceDirection string `json:"faceDirection"`
}

type Color struct {
	Red   int `json:"red"`
	Green int `json:"green"`
	Blue  int `json:"blue"`
}

type Shot struct {
	PlayerID string `json:"playerId"`
	Position `json:"position"`
	Speed    int `json:"speed"`
	Status   string
}

type Game struct {
	Connections map[string]*websocket.Conn
	Players     []*Player `json:"players"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}
var players []*Player
var shots []*Shot
var connections []*websocket.Conn
var isGameLooping bool

const playerWidth = 20
const playerHeight = 20

const gameWidth = 310
const gameHeight = 530

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func processAction(action string, playerID string) {
	player := getPlayerByID(playerID)
	speed := 5
	switch action {
	case "right":
		if player.Position.X < gameWidth {
			player.Position.X += speed
			player.FaceDirection = action
		}
		break
	case "left":
		if player.Position.X > 0 {
			player.Position.X -= speed
			player.FaceDirection = action
		}
		break
	case "up":
		if player.Position.Y > 0 {
			player.Position.Y -= speed
			player.FaceDirection = action
		}
		break
	case "down":
		if player.Position.Y < gameHeight {
			player.Position.Y += speed
			player.FaceDirection = action
		}
		break
	case "shot":
		createShot(player)
		break
	default:
		fmt.Printf("unknown command: %s.", action)
	}
}

func createShot(player *Player) {
	for _, shot := range shots {
		if shot.PlayerID == player.ID {
			return
		}
	}
	shotSpeed := 3
	xPosition := player.Position.X + playerWidth/2
	yPosition := player.Position.Y + playerHeight/2
	position := Position{X: xPosition, Y: yPosition, FaceDirection: player.FaceDirection}
	shot := &Shot{PlayerID: player.ID, Position: position, Speed: shotSpeed}
	shots = append(shots, shot)
}

func moveShot() {
	for i, shot := range shots {
		action := shot.Position.FaceDirection
		speed := shot.Speed
		shouldDestroy := false
		switch action {
		case "right":
			if shot.Position.X <= gameWidth {
				shot.Position.X += speed
			} else {
				shouldDestroy = true
			}
			break
		case "left":
			if shot.Position.X >= 0 {
				shot.Position.X -= speed
			} else {
				shouldDestroy = true
			}
			break
		case "up":
			if shot.Position.Y > 0 {
				shot.Position.Y -= speed
			} else {
				shouldDestroy = true
			}
			break
		case "down":
			if shot.Position.Y <= gameHeight {
				shot.Position.Y += speed
			} else {
				shouldDestroy = true
			}
			break
		}

		shouldDestroy = shouldDestroy || processShotColision(shot)
		if shouldDestroy {
			destroyShot(i)
		}
	}
}

func processShotColision(shot *Shot) bool {
	targetDestroyed := false
	for _, player := range players {
		colisionMaxX := player.Position.X + playerWidth
		colisionMaxY := player.Position.Y + playerHeight
		shotMaxX := shot.Position.X + (playerWidth / 3)
		shotMaxY := shot.Position.Y + (playerHeight / 3)

		if shot.PlayerID != player.ID &&
			player.Status == "ALIVE" &&
			shotMaxX >= player.Position.X &&
			shot.Position.X <= colisionMaxX &&
			shotMaxY >= player.Position.Y &&
			shot.Position.Y <= colisionMaxY {
			player.Status = "DEAD"
			targetDestroyed = true
			break
		}
	}
	return targetDestroyed
}

func clearDeadPlayers() {
	for i, player := range players {
		if player.Status == "DEAD" {
			players = append(players[:i], players[i+1:]...)
		}
	}
}

func destroyShot(index int) {
	shots = append(shots[:index], shots[index+1:]...)
}

func getPlayerByID(playerID string) *Player {
	for _, player := range players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

//Process the game state
func runGameCicle() {
	isGameLooping = true
	for isGameLooping {
		if len(players) > 0 {
			moveShot()
			createResponse()
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Write message back the game state to browser
func createResponse() {
	gameResponse := GameResponse{Players: players, Shots: shots}
	encondedPlayers, _ := json.Marshal(gameResponse)

	for i, connection := range connections {
		if err := connection.WriteMessage(websocket.TextMessage, []byte(encondedPlayers)); err != nil {
			fmt.Printf("Disconnected: %v", err)
			connections = append(connections[:i], connections[i+1:]...)
		}
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	players = make([]*Player, 0)
	shots = make([]*Shot, 0)
	isGameLooping = false

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/content/", http.StripPrefix("/content/", fs))

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("WebsocketError: %v", err)
		}
		position := Position{X: random(0, gameWidth), Y: random(0, gameHeight), FaceDirection: "right"}

		color := Color{Red: random(1, 255), Green: random(1, 254), Blue: random(1, 253)}
		player := &Player{ID: conn.RemoteAddr().String(), Position: position, Color: color, Status: "ALIVE"}
		fmt.Println("New player: %v", player)
		players = append(players, player)
		connections = append(connections, conn)

		for {
			// Read message from browser
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("Removing player : %s", conn.RemoteAddr().String())
				for i, player := range players {
					if player.ID == conn.RemoteAddr().String() {
						players = append(players[:i], players[i+1:]...)
					}
				}
				conn.Close()
				return
			}
			//fmt.Printf("%s sent: %s of type: %v\n", conn.RemoteAddr(), string(msg), msgType)
			processAction(string(msg), conn.RemoteAddr().String())
		}
	})
	go runGameCicle()
	http.ListenAndServe(":"+port, nil)

}
