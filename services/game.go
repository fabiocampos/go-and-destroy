package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/fabiocampos/go-and-destroy/models"
	"github.com/gorilla/websocket"
)

const playerWidth = 20
const playerHeight = 20
const gameWidth = 310
const gameHeight = 530

type GameService struct {
	Players     []*models.Player
	Shots       []*models.Shot
	Connections []*websocket.Conn
}

// NewService creates a new service
func NewGameService() *GameService {
	return &GameService{Players: make([]*models.Player, 0), Shots: make([]*models.Shot, 0)}
}

//Process the game state
func (s *GameService) RunGame() {
	isGameLooping := true
	for isGameLooping {
		if len(s.Players) > 0 {
			s.MoveShot()
			s.ClearDeadPlayers()
			s.CreateResponse()
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Write message back the game state to browser
func (s *GameService) CreateResponse() {
	gameResponse := models.GameResponse{Players: s.Players, Shots: s.Shots}
	encondedPlayers, _ := json.Marshal(gameResponse)

	for i, connection := range s.Connections {
		if err := connection.WriteMessage(websocket.TextMessage, []byte(encondedPlayers)); err != nil {
			s.RemovePlayer(connection)
			fmt.Printf("Disconnected: %v", err)
			s.Connections = append(s.Connections[:i], s.Connections[i+1:]...)
		}
	}
}

func (s *GameService) AddPlayer(conn *websocket.Conn) {
	position := models.Position{X: random(0, gameWidth), Y: random(0, gameHeight), FaceDirection: "right"}

	color := models.Color{Red: random(1, 255), Green: random(1, 254), Blue: random(1, 253)}
	player := &models.Player{ID: conn.RemoteAddr().String(), Position: position, Color: color, Status: "ALIVE"}
	fmt.Println("New player: %v", player)

	//Add a New Player and the respective websocket connection
	s.Players = append(s.Players, player)
	s.Connections = append(s.Connections, conn)
}

func (s *GameService) RemovePlayer(conn *websocket.Conn) {
	fmt.Printf("Removing player : %s", conn.RemoteAddr().String())
	for i, player := range s.Players {
		if player.ID == conn.RemoteAddr().String() {
			s.Players = append(s.Players[:i], s.Players[i+1:]...)
		}
	}
	return
}

func (s *GameService) ProcessAction(action string, playerID string) {
	player := s.GetPlayerByID(playerID)
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
		s.createShot(player)
		break
	default:
		fmt.Printf("unknown command: %s.", action)
	}
}

func (s *GameService) createShot(player *models.Player) {
	for _, shot := range s.Shots {
		if shot.PlayerID == player.ID {
			return
		}
	}
	shotSpeed := 3
	xPosition := player.Position.X + playerWidth/2
	yPosition := player.Position.Y + playerHeight/2
	position := models.Position{X: xPosition, Y: yPosition, FaceDirection: player.FaceDirection}
	shot := &models.Shot{PlayerID: player.ID, Position: position, Speed: shotSpeed}
	s.Shots = append(s.Shots, shot)
}

func (s *GameService) MoveShot() {
	for i, shot := range s.Shots {
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

		shouldDestroy = shouldDestroy || s.processShotColision(shot)
		if shouldDestroy {
			s.DestroyShot(i)
		}
	}
}

func (s *GameService) processShotColision(shot *models.Shot) bool {
	targetDestroyed := false
	for _, player := range s.Players {
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

func (s *GameService) ClearDeadPlayers() {
	for i, player := range s.Players {
		if player.Status == "DEAD" {
			s.Players = append(s.Players[:i], s.Players[i+1:]...)
		}
	}
}

func (s *GameService) DestroyShot(index int) {
	s.Shots = append(s.Shots[:index], s.Shots[index+1:]...)
}

func (s *GameService) GetPlayerByID(playerID string) *models.Player {
	for _, player := range s.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
