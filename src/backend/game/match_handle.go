package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

type WorldMatch struct {
}

type WorldMatchState struct {
	presences            map[string]runtime.Presence
	roomOwningUserIDs    map[string]*Room
	emptyGameTick        int
	ticksUntilNextUpdate int
	// Number of users currently in the process of connecting to the match.
	joinsInProgress int
	TerritoryWorld
}

func RegisterWorldMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &WorldMatch{}, nil
}

func (m *WorldMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &WorldMatchState{
		emptyGameTick:        0,
		presences:            map[string]runtime.Presence{},
		roomOwningUserIDs:    map[string]*Room{},
		ticksUntilNextUpdate: 0,
		TerritoryWorld: *NewTerritoryWorld(
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

func (m *WorldMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	currentTime := time.Now().UTC()

	for _, presence := range presences {
		worldState.emptyGameTick = 0
		worldState.joinsInProgress--

		userID := presence.GetUserId()
		userName := presence.GetUsername()
		joinType := JoinMessage_JOINTYPE_NEWUSER
		if _, ok := worldState.roomOwningUserIDs[userID]; ok {
			joinType = JoinMessage_JOINTYPE_REJOIN
		}

		// Broadcast a message to all matchs that a new user has joined
		joinMessage := JoinLeaveMessage{
			UserId:       userID,
			UserName:     userName,
			JoinType:     joinType,
			JoinDatetime: currentTime,
		}

		joinMessageJson, err := json.Marshal(joinMessage)
		if err != nil {
			logger.Error("error marshaling join message: %v", err)
			continue
		}

		worldState.presences[presence.GetSessionId()] = presence

		dispatcher.BroadcastMessage(int64(OpCode_OPCODE_JOIN), joinMessageJson, nil, nil, true)
	}

	return worldState
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
		leaveMessage := JoinLeaveMessage{
			UserId:       userID,
			UserName:     userName,
			JoinType:     JoinMessage_JOINTYPE_LEAVE,
			JoinDatetime: time.Now().UTC(),
		}

		messageJson, err := json.Marshal(leaveMessage)
		if err != nil {
			logger.Error("error marshaling leave message: %v", err)
			continue
		}

		worldState.Size--
		delete(worldState.presences, presences[i].GetSessionId())

		dispatcher.BroadcastMessage(int64(OpCode_OPCODE_LEAVE), messageJson, nil, nil, true)
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
