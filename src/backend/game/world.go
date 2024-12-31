package game

import (
	"github.com/google/uuid"
)

type TerritoryWorld struct {
	ID                 string
	WidthCapacity      int
	HeightCapacity     int
	Size               int
	IsolatedRooms      map[string]*Room
	OwningRooms        map[string]*Room
	Rooms              map[string]*Room
	MaxWorldTickToLive int
	// The number of tick between each tick of the world. For example, if the value is 60, the world will tick every 60 game ticks.
	TickRatio int
	WorldTick int
}

func NewTerritoryWorld(widthCapacity int, heightCapacity int, maxTickToLive int) *TerritoryWorld {
	return &TerritoryWorld{
		ID:                 uuid.New().String(),
		WidthCapacity:      widthCapacity,
		HeightCapacity:     heightCapacity,
		Size:               0,
		IsolatedRooms:      make(map[string]*Room),
		OwningRooms:        make(map[string]*Room),
		Rooms:              make(map[string]*Room),
		MaxWorldTickToLive: maxTickToLive,
		TickRatio:          10,
		WorldTick:          0,
	}
}

func (w *TerritoryWorld) Capacity() int {
	return w.WidthCapacity * w.HeightCapacity
}

func (w *TerritoryWorld) CreateNewRoom() *Room {
	if w.Capacity() <= len(w.Rooms) {
		return nil
	}

	room := NewRoom(50, 50, "", "")

	rowIndex := w.Size / w.WidthCapacity
	columnIndex := w.Size % w.WidthCapacity

	room.SetWorld(w.ID, rowIndex, columnIndex)
	w.Rooms[room.Name] = room
	w.IsolatedRooms[room.Name] = room

	return room
}

func (w *TerritoryWorld) TakeRoom(userID string) {
	if len(w.IsolatedRooms) == 0 {
		return
	}

	for _, room := range w.IsolatedRooms {
		if room.OwnerID == "" {
			room.OwnerID = userID
			delete(w.IsolatedRooms, room.Name)
			w.OwningRooms[room.Name] = room
			return
		}
	}
}
