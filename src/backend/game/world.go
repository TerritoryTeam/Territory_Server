package game

type TerritoryWorld struct {
	Capacity      int
	Rooms         map[string]*Room
	MaxTickToLive int64
}

func NewTerritoryWorld(capacity int, maxTickToLive int64) *TerritoryWorld {
	return &TerritoryWorld{
		Capacity:      capacity,
		Rooms:         make(map[string]*Room),
		MaxTickToLive: maxTickToLive,
	}
}
