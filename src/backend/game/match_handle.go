package game

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type WorldMatch struct {
}

type WorldMatchState struct {
	Presences         map[string]runtime.Presence
	RoomOwningUserIDs map[string]string
	EmptyTick         int64
	TerritoryWorld
}

func RegisterWorldMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &WorldMatch{}, nil
}

func (m *WorldMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &WorldMatchState{
		EmptyTick:         0,
		Presences:         map[string]runtime.Presence{},
		RoomOwningUserIDs: map[string]string{},
		TerritoryWorld: *NewTerritoryWorld(
			5,
			5,
			600,
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

	for _, presence := range presences {
		worldState.Presences[presence.GetSessionId()] = presence
	}

	return worldState
}

func (m *WorldMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	// Allow all users to join the match
	return state, true, ""
}

func (m *WorldMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for i := 0; i < len(presences); i++ {
		delete(worldState.Presences, presences[i].GetSessionId())
	}

	return worldState
}

func (m *WorldMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	worldState, ok := state.(*WorldMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	// If there are no presences in the match, increment the empty tick counter
	if len(worldState.Presences) == 0 {
		worldState.EmptyTick++
	}

	// If the match has been live for more than max living ticks, end the match by returning nil
	if tick > worldState.MaxTickToLive {
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
