package game

import (
	"time"
)

type OpCode int32

const (
	OpCode_OPCODE_JOIN  OpCode = 1
	OpCode_OPCODE_LEAVE OpCode = 2
)

const (
	JoinMessage_JOINTYPE_NEWUSER string = "NewUser"
	JoinMessage_JOINTYPE_REJOIN  string = "ReJoin"
	JoinMessage_JOINTYPE_LEAVE   string = "Leave"
)

type JoinLeaveMessage struct {
	UserId       string    `json:"userId"`
	UserName     string    `json:"userName"`
	JoinType     string    `json:"joinType"`
	JoinDatetime time.Time `json:"joinDatetime"`
}

type RoomUpdateMessage struct {
	room Room `json:"room"`
}
