package matches

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"territory.com/server/game/api"
	"territory.com/server/game/models"
)

type WorldMatch struct {
	marshaler   *protojson.MarshalOptions
	unmarshaler *protojson.UnmarshalOptions
}

type WorldMatchState struct {
	presences            map[string]runtime.Presence
	roomOwningUserIDs    map[string]*models.Room
	emptyGameTick        int
	ticksUntilNextUpdate int
	// Number of users currently in the process of connecting to the match.
	joinsInProgress int
	models.TerritoryWorld
}

func RegisterWorldMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &WorldMatch{
		marshaler: &protojson.MarshalOptions{
			UseEnumNumbers: true,
		},
		unmarshaler: &protojson.UnmarshalOptions{
			DiscardUnknown: false,
		},
	}, nil
}

func (m *WorldMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &WorldMatchState{
		emptyGameTick:        0,
		presences:            map[string]runtime.Presence{},
		roomOwningUserIDs:    map[string]*models.Room{},
		ticksUntilNextUpdate: 0,
		TerritoryWorld: *models.NewTerritoryWorld(
			5,
			5,
			100,
		),
	}

	capacity := state.Capacity()
	for i := 0; i < capacity; i++ {
		logger.Info("Creating new room: %d", i)
		state.CreateNewRoom()
	}

	tickRate := 1 // 1 tick per second = 1 MatchLoop func invocations per second
	label := "Territory World Match Demo"
	return state, tickRate, label
}

func (m *WorldMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	// Allow all users to join the match

	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid state object")
		return state, false, "invalid state"
	}

	// Check if it's a user attempting to rejoin after a disconnect.
	if presence, ok := worldState.presences[presence.GetSessionId()]; ok {
		if presence == nil {
			// User rejoining after a disconnect.
			worldState.joinsInProgress++
			return state, true, ""
		} else {
			// User attempting to join from 2 different devices at the same time.
			return state, false, "user already in match"
		}
	}

	// Check if match is full
	if len(worldState.presences)+worldState.joinsInProgress >= worldState.Capacity() {
		return state, false, "match is full"
	}

	worldState.joinsInProgress++
	return state, true, ""
}

func (m *WorldMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	currentPresences := make([]runtime.Presence, 0, len(worldState.presences))
	for _, presence := range worldState.presences {
		currentPresences = append(currentPresences, presence)
	}

	var simpleRooms *api.ListWorldRoomsMessage = nil

	for _, presence := range presences {
		worldState.emptyGameTick = 0
		worldState.presences[presence.GetSessionId()] = presence
		worldState.joinsInProgress--

		userID := presence.GetUserId()
		userName := presence.GetUsername()

		joinType := api.JoinType_JOIN_TYPE_NEW_USER_JOIN
		if _, ok := worldState.roomOwningUserIDs[userID]; ok {
			joinType = api.JoinType_JOIN_TYPE_USER_REJOIN
		}

		if len(currentPresences) > 0 {
			// Broadcast a message to all matches that a new user has joined
			joinMessage := &api.UserJoinLeaveMessage{
				UserId:   userID,
				UserName: userName,
				JoinType: joinType,
			}

			joinMessageBuf, err := m.marshaler.Marshal(joinMessage)
			if err != nil {
				logger.Error("error marshaling join message: %v", err)
				continue
			}

			dispatcher.BroadcastMessage(int64(api.OpCode_OPCODE_USER_JOIN), joinMessageBuf, currentPresences, nil, false)
		}

		// Check if the user is already owning a room
		if _, ok := worldState.roomOwningUserIDs[userID]; !ok {
			if simpleRooms == nil {
				totalRooms := len(worldState.Rooms)

				simpleRooms = &api.ListWorldRoomsMessage{
					WorldId:   worldState.ID,
					RoomTotal: int32(totalRooms),
					Rooms:     make([]*api.Room, 0, len(worldState.Rooms)),
				}

				for _, room := range worldState.Rooms {
					var ownerRoom *api.User = nil
					if room.OwnerID != "" {
						account, err := nk.AccountGetId(ctx, room.OwnerID)
						if err != nil {
							logger.Error("error getting account for room owner: %v", err)
							continue
						}

						ownerRoom = &api.User{
							UserId:     room.OwnerID,
							UserName:   account.User.DisplayName,
							UserAvatar: account.User.AvatarUrl,
						}
					}

					simpleRooms.Rooms = append(simpleRooms.Rooms, &api.Room{
						RoomX:     int32(room.RoomX),
						RoomY:     int32(room.RoomY),
						UserOwner: ownerRoom,
					})
				}
			}

			roomListMessageBuf, err := m.marshaler.Marshal(simpleRooms)
			if err != nil {
				logger.Error("error marshaling room update message: %v", err)
				continue
			}

			dispatcher.BroadcastMessage(int64(api.OpCode_OPCODE_ROOMS_LIST_AVAILABLE), roomListMessageBuf, []runtime.Presence{presence}, nil, true)
		}

		currentPresences = append(currentPresences, presence)
	}

	return worldState
}

func (m *WorldMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for i := 0; i < len(presences); i++ {
		userID := presences[i].GetUserId()
		userName := presences[i].GetUsername()

		// Broadcast a message to all match that a user has left
		leaveMessage := &api.UserJoinLeaveMessage{
			UserId:   userID,
			UserName: userName,
			JoinType: api.JoinType_JOIN_TYPE_USER_LEAVE,
		}

		messageBuf, err := m.marshaler.Marshal(leaveMessage)
		if err != nil {
			logger.Error("error marshaling leave message: %v", err)
			continue
		}

		worldState.Size--
		delete(worldState.presences, presences[i].GetSessionId())

		dispatcher.BroadcastMessage(int64(api.OpCode_OPCODE_USER_LEAVE), messageBuf, nil, nil, true)

	}

	return worldState
}

func (m *WorldMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	worldState.ticksUntilNextUpdate--

	if worldState.ticksUntilNextUpdate <= 0 {
		worldState.WorldTick++
		worldState.ticksUntilNextUpdate = worldState.TickRatio
	}

	// If there are no presences in the match, increment the empty tick counter
	if len(worldState.presences) == 0 {
		worldState.emptyGameTick++
	}

	// There's a game in progress. Check for input, update match state, and send messages to clients.
	// for _, message := range messages {
	// 	switch api.OpCode(message.GetOpCode()) {
	// 	case api.OpCode_OPCODE_ROOMS_LIST_AVAILABLE:
	// 		// List all available rooms
	// 		msg := &api.AvailableRoomsListParameters{}
	// 		err := m.unmarshaler.Unmarshal(message.GetData(), msg)
	// 		if err != nil {
	// 			_ = dispatcher.BroadcastMessage(int64(api.OpCode_OPCODE_USER_REJECT), nil, []runtime.Presence{message}, nil, true)
	// 			continue
	// 		}

	// 		isolatedRoomsCount := len(worldState.IsolatedRooms)
	// 		roomsListMessage := &api.AvailableRoomsMessage{
	// 			Size:           int32(isolatedRoomsCount),
	// 			AvailableRooms: nil,
	// 			Offset:         0,
	// 			Total:          0,
	// 		}
	// 		buf, err := m.marshaler.Marshal(roomsListMessage)

	// 		if err == nil {
	// 			_ = dispatcher.BroadcastMessage((int64(api.OpCode_OPCODE_ROOMS_LIST_AVAILABLE)), buf, []runtime.Presence{message}, nil, true)
	// 		}

	// case OPCODE_ROOMS_TAKE_OVER:
	// 	// Take a room
	// 	msg := &api.RoomTakeOverParameters{}

	// 	default:
	// 		logger.Warn("Unrecognized OpCode: %v", message.GetOpCode())
	// 		_ = dispatcher.BroadcastMessage(int64(api.OpCode_OPCODE_USER_REJECT), nil, []runtime.Presence{message}, nil, true)
	// 	}
	// }

	// If the match has been live for more than max living ticks, end the match by returning nil
	if worldState.WorldTick > worldState.MaxWorldTickToLive {
		return nil
	}

	return worldState
}

func (m *WorldMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	logger.Debug("match will terminate in %d seconds", graceSeconds)

	var matchId string

	// Find an existing match for the remaining connected presences to join
	limit := 1
	authoritative := true
	label := ""
	minSize := 2
	maxSize := 4
	query := "*"
	availableMatches, err := nk.MatchList(ctx, limit, authoritative, label, &minSize, &maxSize, query)
	if err != nil {
		logger.Error("error listing matches", err)
		return nil
	}

	if len(availableMatches) > 0 {
		matchId = availableMatches[0].MatchId
	} else {
		// No available matches, create a new match instead
		matchId, err = nk.MatchCreate(ctx, "match", nil)
		if err != nil {
			logger.Error("error creating match", err)
			return nil
		}
	}

	// Broadcast the new match id to all remaining connected presences
	data := map[string]string{
		matchId: matchId,
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		logger.Error("error marshaling new match message")
		return nil
	}

	dispatcher.BroadcastMessage(999, dataJson, nil, nil, true)

	return state
}

func (m *WorldMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
