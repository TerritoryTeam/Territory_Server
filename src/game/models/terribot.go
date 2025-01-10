package models

type TerribotPosition struct {
	RoomID string
	X      int
	Y      int
}

type Terribot struct {
	ID   string
	Body []byte
	TerribotPosition
}
