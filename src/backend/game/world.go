package game

import (
	"github.com/google/uuid"
)

type TerritoryWorld struct {
	ID             string
	WidthCapacity  int
	HeightCapacity int
	Size           int
	IsolatedRooms  map[string]*Room
	OwningRooms    map[string]*Room
	Rooms          map[string]*Room
	MaxTickToLive  int64
}

func NewTerritoryWorld(widthCapacity int, heightCapacity int, maxTickToLive int64) *TerritoryWorld {
	return &TerritoryWorld{
		ID:             uuid.New().String(),
		WidthCapacity:  widthCapacity,
		HeightCapacity: heightCapacity,
		Size:           0,
		IsolatedRooms:  make(map[string]*Room),
		OwningRooms:    make(map[string]*Room),
		Rooms:          make(map[string]*Room),
		MaxTickToLive:  maxTickToLive,
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
