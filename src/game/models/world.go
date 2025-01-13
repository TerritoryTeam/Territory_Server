package models

import (
	"sync"

	"github.com/google/uuid"
)

type void struct{}

type TerritoryWorld struct {
	ID             string
	WidthCapacity  int
	HeightCapacity int
	Size           int
	// Room Name
	IsolatedRooms map[string]void
	// User Id to Room Name
	OwningRooms map[string]string
	// Room Name to Room
	Rooms              map[string]*Room
	MaxWorldTickToLive int
	// The number of tick between each tick of the world. For example, if the value is 60, the world will tick every 60 game ticks.
	TickRatio int
	WorldTick int
	worldLock sync.Mutex
}

func NewTerritoryWorld(widthCapacity int, heightCapacity int, maxTickToLive int) *TerritoryWorld {
	return &TerritoryWorld{
		ID:                 uuid.New().String(),
		WidthCapacity:      widthCapacity,
		HeightCapacity:     heightCapacity,
		Size:               0,
		IsolatedRooms:      make(map[string]void),
		OwningRooms:        make(map[string]string),
		Rooms:              make(map[string]*Room),
		MaxWorldTickToLive: maxTickToLive,
		TickRatio:          10,
		WorldTick:          0,
	}
}

func (w *TerritoryWorld) Capacity() int {
	return w.WidthCapacity * w.HeightCapacity
}

func (w *TerritoryWorld) InitializeRooms() {
	defer w.worldLock.Unlock()

	w.worldLock.Lock()

	for rowIndex := 0; rowIndex < w.HeightCapacity; rowIndex++ {
		for columnIndex := 0; columnIndex < w.WidthCapacity; columnIndex++ {
			room := NewRoom(50, 50, "", "")
			room.SetWorld(w.ID, rowIndex, columnIndex)
			w.Rooms[room.Name] = room
			w.IsolatedRooms[room.Name] = void{}
		}
	}
}
