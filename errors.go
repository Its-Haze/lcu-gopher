package lcu

import "errors"

var (
	ErrSummonerNotFound         = errors.New("summoner not found")
	ErrSummonerNotInGame        = errors.New("summoner not in game")
	ErrSummonerNotInLobby       = errors.New("summoner not in lobby")
	ErrSummonerNotInChampSelect = errors.New("summoner not in champ select")
	ErrSummonerNotInQueue       = errors.New("summoner not in queue")
)
