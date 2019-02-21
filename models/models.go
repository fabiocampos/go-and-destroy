package models

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
