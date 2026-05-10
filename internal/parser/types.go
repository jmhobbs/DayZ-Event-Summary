package parser

type LogLine struct {
	Timestamp string
	Type      string
	Subject   *Player
	Object    *Player
}

type Player struct {
	ID       string
	Name     string
	Position *Position
}

type Position struct {
	X float64
	Y float64
	Z float64
}
