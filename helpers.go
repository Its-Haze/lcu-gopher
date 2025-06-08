package lcu

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetCurrentSummoner retrieves information about the currently logged-in summoner
func (c *Client) GetCurrentSummoner() (*Summoner, error) {
	resp, err := c.Get("/lol-summoner/v1/current-summoner")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get summoner info: status %d", resp.StatusCode)
	}

	var summoner Summoner
	if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
		return nil, fmt.Errorf("failed to decode summoner: %w", err)
	}

	return &summoner, nil
}

// GetSummonerByName retrieves summoner information by name
func (c *Client) GetSummonerByName(name string) (*Summoner, error) {
	resp, err := c.Get("/lol-summoner/v1/summoners?name=" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrSummonerNotFound, resp.StatusCode)
	}

	var summoner Summoner
	if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
		return nil, fmt.Errorf("failed to decode summoner: %w", err)
	}

	return &summoner, nil
}

// GetChampSelectSession retrieves the current champion select session
func (c *Client) GetChampSelectSession() (*ChampSelectSession, error) {
	resp, err := c.Get("/lol-champ-select/v1/session")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not in champion select")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get champion select session: status %d", resp.StatusCode)
	}

	var session ChampSelectSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode champion select session: %w", err)
	}

	return &session, nil
}

// GetFriendsList retrieves the friends list
func (c *Client) GetFriendsList() ([]Friend, error) {
	resp, err := c.Get("/lol-chat/v1/friends")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get friends list: status %d", resp.StatusCode)
	}

	var friends []Friend
	if err := json.NewDecoder(resp.Body).Decode(&friends); err != nil {
		return nil, fmt.Errorf("failed to decode friends list: %w", err)
	}

	return friends, nil
}

// GetLobby retrieves the current lobby information
func (c *Client) GetLobby() (*Lobby, error) {
	resp, err := c.Get("/lol-lobby/v2/lobby")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not in a lobby")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get lobby: status %d", resp.StatusCode)
	}

	var lobby Lobby
	if err := json.NewDecoder(resp.Body).Decode(&lobby); err != nil {
		return nil, fmt.Errorf("failed to decode lobby: %w", err)
	}

	return &lobby, nil
}

// GetMatchmakingSearchState retrieves the current matchmaking search state
func (c *Client) GetMatchmakingSearchState() (*MatchmakingSearchState, error) {
	resp, err := c.Get("/lol-lobby/v2/lobby/matchmaking/search-state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get matchmaking search state: status %d", resp.StatusCode)
	}

	var state MatchmakingSearchState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode matchmaking search state: %w", err)
	}

	return &state, nil
}

// Common position constants
const (
	PositionTop     = "top"
	PositionJungle  = "jungle"
	PositionMiddle  = "middle"
	PositionBottom  = "bottom"
	PositionUtility = "utility"
	PositionFill    = "fill"
)

// RankedStats represents a summoner's ranked statistics
type RankedStats struct {
	QueueMap map[string]RankedQueueStats `json:"queueMap"`
}

// RankedQueueStats represents stats for a specific queue
type RankedQueueStats struct {
	LeaguePoints int    `json:"leaguePoints"`
	Rank         string `json:"rank"`
	Tier         string `json:"tier"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
}

// GetRankedStats retrieves the ranked statistics for the current summoner
func (c *Client) GetRankedStats() (*RankedStats, error) {
	resp, err := c.Get("/lol-ranked/v1/current-ranked-stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get ranked stats: status %d", resp.StatusCode)
	}

	var stats RankedStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode ranked stats: %w", err)
	}

	return &stats, nil
}

// GetGameSession returns the current game session information
func (c *Client) GetGameSession() (*GameSession, error) {
	resp, err := c.Get("/lol-gameflow/v1/session")
	if err != nil {
		return nil, fmt.Errorf("failed to get game session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var session GameSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode game session: %w", err)
	}

	return &session, nil
}

// SubscribeToGamePhase subscribes to game phase changes
func (c *Client) SubscribeToGamePhase(handler func(phase GamePhase)) error {
	return c.Subscribe("/lol-gameflow/v1/session", func(event *Event) {
		if event.EventType == string(EventTypeUpdate) {
			if data, ok := event.Data.(map[string]interface{}); ok {
				if phase, ok := data["phase"].(string); ok {
					handler(GamePhase(phase))
				}
			}
		}
	}, EventTypeUpdate)
}
