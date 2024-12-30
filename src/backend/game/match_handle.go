package game

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type LobbyMatch struct {
}

type LobbyMatchState struct {
	presences  map[string]runtime.Presence
	emptyTicks int
}

func RegisterLobbyMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return &LobbyMatch{}, nil
}

func (m *LobbyMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &LobbyMatchState{
		emptyTicks: 0,
		presences:  map[string]runtime.Presence{},
	}
	tickRate := 1 // 1 tick per second = 1 MatchLoop func invocations per second
	label := ""
	return state, tickRate, label
}

func (m *LobbyMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	lobbyState, ok := state.(*LobbyMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for i := 0; i < len(presences); i++ {
		lobbyState.presences[presences[i].GetSessionId()] = presences[i]
	}

	return lobbyState
}

func (m *LobbyMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	// Allow all users to join the match
	return state, true, ""
}

func (m *LobbyMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	lobbyState, ok := state.(*LobbyMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	for i := 0; i < len(presences); i++ {
		delete(lobbyState.presences, presences[i].GetSessionId())
	}

	return lobbyState
}

func (m *LobbyMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	lobbyState, ok := state.(*LobbyMatchState)
	if !ok {
		logger.Error("state not a valid lobby state object")
		return nil
	}

	// If we have no presences in the match according to the match state, increment the empty ticks count
	if len(lobbyState.presences) == 0 {
		lobbyState.emptyTicks++
	}

	// If the match has been empty for more than 100 ticks, end the match by returning nil
	if lobbyState.emptyTicks > 100 {
		return nil
	}

	return lobbyState
}

func (m *LobbyMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
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

func (m *LobbyMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
