package models

import (
	"math/rand"
	"strconv"
	"time"
)

const (
	smoothLevel   int = 1
	fillThreshold int = 55
	wallThreshold int = 200
	caveThreshold int = 50
	boardSize     int = 1
)

type RoomBase struct {
	Name        string
	DisplayName string
	Width       int
	Height      int
	OwnerID     string
}

type Room struct {
	RoomBase
	RoomPosition
	Grid [][]byte
}

type RoomPosition struct {
	WorldID string
	RoomX   int
	RoomY   int
}

func (r *Room) SetWorld(worldID string, roomX, roomY int) {
	r.RoomPosition.WorldID = worldID
	r.RoomPosition.RoomX = roomX
	r.RoomPosition.RoomY = roomY

	r.Name = "W" + strconv.Itoa(roomX) + "H" + strconv.Itoa(roomY)
	r.DisplayName = "LinkedRoom"
}

func NewRoom(width, height int, ownerId string, seed string) *Room {
	grid := generateRoomGrid(width, height, seed)

	return &Room{
		RoomBase: RoomBase{
			Name:        "Isolated Room",
			DisplayName: "Isolated Room",
			Width:       width,
			Height:      height,
			OwnerID:     ownerId,
		},
		Grid: grid,
		RoomPosition: RoomPosition{
			WorldID: "",
			RoomX:   0,
			RoomY:   0,
		},
	}
}

func generateRoomGrid(width, height int, seed string) [][]byte {

	grid := make([][]byte, height)
	for rowIndex := 0; rowIndex < height; rowIndex++ {
		grid[rowIndex] = make([]byte, width)
	}

	randomFillRoomGrid(grid, width, height, seed)

	for smoothIteration := 0; smoothIteration < smoothLevel; smoothIteration++ {
		smoothRoomGrid(grid, width, height)
	}

	return grid
}

func randomFillRoomGrid(grid [][]byte, width, height int, seed string) {
	if seed == "" {
		seed = strconv.Itoa(int(time.Now().UnixNano()))
	}

	seedHash := int(hashString(seed))
	rand.Seed(int64(seedHash))

	for rowIndex := 0; rowIndex < height; rowIndex++ {
		for colIndex := 0; colIndex < width; colIndex++ {
			if rowIndex == 0 || rowIndex == height-1 || colIndex == 0 || colIndex == width-1 {
				// Create walls around the room
				grid[rowIndex][colIndex] = BlockIDWall
			} else {
				if rand.Intn(100) < fillThreshold {
					grid[rowIndex][colIndex] = BlockIDWall
				} else {
					grid[rowIndex][colIndex] = BlockIDEmpty
				}
			}
		}
	}
}

func smoothRoomGrid(grid [][]byte, width, height int) {
	for rowIndex := 0; rowIndex < height; rowIndex++ {
		for colIndex := 0; colIndex < width; colIndex++ {
			neighborWallTiles := getSurroundWallCount(grid, colIndex, rowIndex, width, height)
			if neighborWallTiles > 4 {
				grid[rowIndex][colIndex] = BlockIDWall
			} else if neighborWallTiles < 4 {
				grid[rowIndex][colIndex] = BlockIDEmpty
			}
		}
	}
}

func getSurroundWallCount(grid [][]byte, x, y, width, height int) int {
	wallCount := 0
	for neighborY := y - 1; neighborY <= y+1; neighborY++ {
		for neighborX := x - 1; neighborX <= x+1; neighborX++ {
			if neighborX >= 0 && neighborX < width && neighborY >= 0 && neighborY < height {
				if neighborX != x || neighborY != y {
					if grid[neighborY][neighborX] == BlockIDWall {
						wallCount++
					}
				}
			} else {
				wallCount++
			}
		}
	}

	return wallCount
}

func hashString(s string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		hash = hash ^ uint32(s[i])
		hash = hash * 16777619
	}
	return hash
}
