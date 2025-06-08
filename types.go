package lcu

// EventType represents the type of LCU event
type EventType string

const (
	EventTypeCreate EventType = "Create"
	EventTypeUpdate EventType = "Update"
	EventTypeDelete EventType = "Delete"
)

// Common queue IDs for matchmaking
const (
	QueueCustom       = 0    // Custom Game
	QueueNormalBlind  = 430  // Normal Blind Pick (incorrect - should be 430 for draft)
	QueueNormalDraft  = 400  // Normal Draft Pick
	QueueRankedSolo   = 420  // Ranked Solo/Duo
	QueueRankedFlex   = 440  // Ranked Flex
	QueueARAM         = 450  // ARAM
	QueueClash        = 700  // Clash
	QueueTutorial     = 2000 // Tutorial
	QueueBeginner     = 830  // Intro Bot
	QueueIntermediate = 840  // Beginner Bot
	QueueAdvanced     = 850  // Intermediate Bot
)

// GamePhase represents the current game phase
type GamePhase string

const (
	GamePhaseNone            GamePhase = "None"
	GamePhaseLobby           GamePhase = "Lobby"
	GamePhaseMatchmaking     GamePhase = "Matchmaking"
	GamePhaseChampSelect     GamePhase = "ChampSelect"
	GamePhaseInProgress      GamePhase = "InProgress"
	GamePhaseWaitingForStats GamePhase = "WaitingForStats"
	GamePhasePreEndOfGame    GamePhase = "PreEndOfGame"
	GamePhaseEndOfGame       GamePhase = "EndOfGame"
)

// Summoner represents a League of Legends summoner
type Summoner struct {
	AccountID                   int64  `json:"accountId"`
	DisplayName                 string `json:"displayName"`
	GameName                    string `json:"gameName"`
	InternalName                string `json:"internalName"`
	NameChangeFlag              bool   `json:"nameChangeFlag"`
	PercentCompleteForNextLevel int    `json:"percentCompleteForNextLevel"`
	Privacy                     string `json:"privacy"`
	ProfileIconID               int    `json:"profileIconId"`
	Puuid                       string `json:"puuid"`
	RerollPoints                struct {
		CurrentPoints    int `json:"currentPoints"`
		MaxRolls         int `json:"maxRolls"`
		NumberOfRolls    int `json:"numberOfRolls"`
		PointsCostToRoll int `json:"pointsCostToRoll"`
		PointsToReroll   int `json:"pointsToReroll"`
	} `json:"rerollPoints"`
	SummonerID       int64  `json:"summonerId"`
	SummonerLevel    int    `json:"summonerLevel"`
	TagLine          string `json:"tagLine"`
	Unnamed          bool   `json:"unnamed"`
	XpSinceLastLevel int    `json:"xpSinceLastLevel"`
	XpUntilNextLevel int    `json:"xpUntilNextLevel"`
}

// GameSession represents current game session info
type GameSession struct {
	GameClient struct {
		ObserverServerIP   string `json:"observerServerIP"`
		ObserverServerPort int    `json:"observerServerPort"`
		Running            bool   `json:"running"`
		ServerIP           string `json:"serverIP"`
		ServerPort         int    `json:"serverPort"`
		Visible            bool   `json:"visible"`
	} `json:"gameClient"`
	GameDodge struct {
		DodgeID int    `json:"dodgeId"`
		Phase   string `json:"phase"`
		State   string `json:"state"`
	} `json:"gameDodge"`
	Map struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"map"`
	Mode struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"mode"`
	Phase string `json:"phase"`
}

// ChampSelectSession represents the current champion select session
type ChampSelectSession struct {
	Actions             [][]ChampSelectAction `json:"actions"`
	AllowBattleBoost    bool                  `json:"allowBattleBoost"`
	AllowDuplicatePicks bool                  `json:"allowDuplicatePicks"`
	AllowLockedEvents   bool                  `json:"allowLockedEvents"`
	AllowRerolling      bool                  `json:"allowRerolling"`
	AllowSkinSelection  bool                  `json:"allowSkinSelection"`
	Bans                struct {
		MyTeamBans    []int `json:"myTeamBans"`
		TheirTeamBans []int `json:"theirTeamBans"`
		NumBans       int   `json:"numBans"`
	} `json:"bans"`
	BenchChampionIds   []int `json:"benchChampionIds"`
	BenchEnabled       bool  `json:"benchEnabled"`
	BoostableSkinCount int   `json:"boostableSkinCount"`
	ChatDetails        struct {
		ChatRoomName     string `json:"chatRoomName"`
		ChatRoomPassword string `json:"chatRoomPassword"`
	} `json:"chatDetails"`
	Counter              int `json:"counter"`
	EntitledFeatureState struct {
		AdditionalRerolls int   `json:"additionalRerolls"`
		UnlockedSkinIds   []int `json:"unlockedSkinIds"`
	} `json:"entitledFeatureState"`
	GameId               int64               `json:"gameId"`
	HasSimultaneousBans  bool                `json:"hasSimultaneousBans"`
	HasSimultaneousPicks bool                `json:"hasSimultaneousPicks"`
	IsCustomGame         bool                `json:"isCustomGame"`
	IsSpectating         bool                `json:"isSpectating"`
	LocalPlayerCellId    int                 `json:"localPlayerCellId"`
	LockedEventIndex     int                 `json:"lockedEventIndex"`
	MyTeam               []ChampSelectPlayer `json:"myTeam"`
	RerollsRemaining     int                 `json:"rerollsRemaining"`
	SkipChampionSelect   bool                `json:"skipChampionSelect"`
	TheirTeam            []ChampSelectPlayer `json:"theirTeam"`
	Timer                struct {
		AdjustedTimeLeftInPhase int64  `json:"adjustedTimeLeftInPhase"`
		InternalNowInEpochMs    int64  `json:"internalNowInEpochMs"`
		IsInfinite              bool   `json:"isInfinite"`
		Phase                   string `json:"phase"`
		TotalTimeInPhase        int64  `json:"totalTimeInPhase"`
	} `json:"timer"`
	TradingEnabled bool `json:"tradingEnabled"`
}

// ChampSelectAction represents an action in champion select
type ChampSelectAction struct {
	ActorCellId  int    `json:"actorCellId"`
	ChampionId   int    `json:"championId"`
	Completed    bool   `json:"completed"`
	Id           int    `json:"id"`
	IsAllyAction bool   `json:"isAllyAction"`
	IsInProgress bool   `json:"isInProgress"`
	PickTurn     int    `json:"pickTurn"`
	Type         string `json:"type"`
}

// ChampSelectPlayer represents a player in champion select
type ChampSelectPlayer struct {
	AssignedPosition     string `json:"assignedPosition"`
	CellId               int    `json:"cellId"`
	ChampionId           int    `json:"championId"`
	ChampionPickIntent   int    `json:"championPickIntent"`
	EntitledFeatureType  string `json:"entitledFeatureType"`
	NameVisibilityType   string `json:"nameVisibilityType"`
	ObfuscatedPuuid      string `json:"obfuscatedPuuid"`
	ObfuscatedSummonerId int64  `json:"obfuscatedSummonerId"`
	PickTurn             int    `json:"pickTurn"`
	Puuid                string `json:"puuid"`
	SelectedSkinId       int    `json:"selectedSkinId"`
	Spell1Id             int    `json:"spell1Id"`
	Spell2Id             int    `json:"spell2Id"`
	SummonerId           int64  `json:"summonerId"`
	Team                 int    `json:"team"`
	WardSkinId           int    `json:"wardSkinId"`
}

// Friend represents a friend in the friends list
type Friend struct {
	Availability            string      `json:"availability"`
	DisplayGroupId          int         `json:"displayGroupId"`
	DisplayGroupName        string      `json:"displayGroupName"`
	GameName                string      `json:"gameName"`
	GameTag                 string      `json:"gameTag"`
	GroupId                 int         `json:"groupId"`
	GroupName               string      `json:"groupName"`
	Icon                    int         `json:"icon"`
	Id                      string      `json:"id"`
	IsP2PConversationMuted  bool        `json:"isP2PConversationMuted"`
	LastSeenOnlineTimestamp interface{} `json:"lastSeenOnlineTimestamp"`
	Lol                     struct {
		BannerIdSelected         string `json:"bannerIdSelected"`
		ChallengeCrystalLevel    string `json:"challengeCrystalLevel"`
		ChallengePoints          string `json:"challengePoints"`
		ChallengeTokensSelected  string `json:"challengeTokensSelected"`
		ChampionId               string `json:"championId"`
		CompanionId              string `json:"companionId"`
		DamageSkinId             string `json:"damageSkinId"`
		GameId                   string `json:"gameId"`
		GameMode                 string `json:"gameMode"`
		GameQueueType            string `json:"gameQueueType"`
		GameStatus               string `json:"gameStatus"`
		IconOverride             string `json:"iconOverride"`
		IsObservable             string `json:"isObservable"`
		LegendaryMasteryScore    string `json:"legendaryMasteryScore"`
		Level                    string `json:"level"`
		MapId                    string `json:"mapId"`
		MapSkinId                string `json:"mapSkinId"`
		PlayerTitleSelected      string `json:"playerTitleSelected"`
		ProfileIcon              string `json:"profileIcon"`
		Pty                      string `json:"pty"`
		Puuid                    string `json:"puuid"`
		QueueId                  string `json:"queueId"`
		RankedLeagueDivision     string `json:"rankedLeagueDivision"`
		RankedLeagueQueue        string `json:"rankedLeagueQueue"`
		RankedLeagueTier         string `json:"rankedLeagueTier"`
		RankedLosses             string `json:"rankedLosses"`
		RankedPrevSeasonDivision string `json:"rankedPrevSeasonDivision"`
		RankedPrevSeasonTier     string `json:"rankedPrevSeasonTier"`
		RankedSplitRewardLevel   string `json:"rankedSplitRewardLevel"`
		RankedWins               string `json:"rankedWins"`
		Regalia                  string `json:"regalia"`
		SkinVariant              string `json:"skinVariant"`
		Skinname                 string `json:"skinname"`
		TimeStamp                string `json:"timeStamp"`
	} `json:"lol"`
	Name          string `json:"name"`
	Note          string `json:"note"`
	Patchline     string `json:"patchline"`
	Pid           string `json:"pid"`
	PlatformId    string `json:"platformId"`
	Product       string `json:"product"`
	ProductName   string `json:"productName"`
	Puuid         string `json:"puuid"`
	StatusMessage string `json:"statusMessage"`
	Summary       string `json:"summary"`
	SummonerId    int64  `json:"summonerId"`
	Time          int64  `json:"time"`
}

// RunePage represents a rune page configuration
type RunePage struct {
	AutoModifiedSelections []interface{} `json:"autoModifiedSelections"`
	Current                bool          `json:"current"`
	Id                     int           `json:"id"`
	IsActive               bool          `json:"isActive"`
	IsDeletable            bool          `json:"isDeletable"`
	IsEditable             bool          `json:"isEditable"`
	IsValid                bool          `json:"isValid"`
	LastModified           int64         `json:"lastModified"`
	Name                   string        `json:"name"`
	Order                  int           `json:"order"`
	PrimaryStyleId         int           `json:"primaryStyleId"`
	SelectedPerkIds        []int         `json:"selectedPerkIds"`
	SubStyleId             int           `json:"subStyleId"`
}

// Lobby represents a lobby/party
type Lobby struct {
	CanStartActivity bool   `json:"canStartActivity"`
	ChatRoomId       string `json:"chatRoomId"`
	ChatRoomKey      string `json:"chatRoomKey"`
	GameConfig       struct {
		AllowablePremadeSizes        []int         `json:"allowablePremadeSizes"`
		CustomLobbyName              string        `json:"customLobbyName"`
		CustomMutatorName            string        `json:"customMutatorName"`
		CustomRewardsDisabledReasons []string      `json:"customRewardsDisabledReasons"`
		CustomSpectatorPolicy        string        `json:"customSpectatorPolicy"`
		CustomSpectators             []interface{} `json:"customSpectators"`
		CustomTeam100                []interface{} `json:"customTeam100"`
		CustomTeam200                []interface{} `json:"customTeam200"`
		GameMode                     string        `json:"gameMode"`
		IsCustom                     bool          `json:"isCustom"`
		IsLobbyFull                  bool          `json:"isLobbyFull"`
		IsTeamBuilderManaged         bool          `json:"isTeamBuilderManaged"`
		MapId                        int           `json:"mapId"`
		MaxHumanPlayers              int           `json:"maxHumanPlayers"`
		MaxLobbySize                 int           `json:"maxLobbySize"`
		MaxTeamSize                  int           `json:"maxTeamSize"`
		PickType                     string        `json:"pickType"`
		PremadeSizeAllowed           bool          `json:"premadeSizeAllowed"`
		QueueId                      int           `json:"queueId"`
		ShowPositionSelector         bool          `json:"showPositionSelector"`
	} `json:"gameConfig"`
	Invitations  []interface{} `json:"invitations"`
	LocalMember  LobbyMember   `json:"localMember"`
	LobbyId      string        `json:"lobbyId"`
	Members      []LobbyMember `json:"members"`
	PartyId      string        `json:"partyId"`
	PartyType    string        `json:"partyType"`
	Restrictions []interface{} `json:"restrictions"`
	Warnings     []interface{} `json:"warnings"`
}

// LobbyMember represents a member in a lobby
type LobbyMember struct {
	AllowedChangeActivity         bool   `json:"allowedChangeActivity"`
	AllowedInviteOthers           bool   `json:"allowedInviteOthers"`
	AllowedKickOthers             bool   `json:"allowedKickOthers"`
	AllowedStartActivity          bool   `json:"allowedStartActivity"`
	AllowedToggleInvite           bool   `json:"allowedToggleInvite"`
	AutoFillEligible              bool   `json:"autoFillEligible"`
	AutoFillProtectedForPromos    bool   `json:"autoFillProtectedForPromos"`
	AutoFillProtectedForSoloing   bool   `json:"autoFillProtectedForSoloing"`
	AutoFillProtectedForStreaking bool   `json:"autoFillProtectedForStreaking"`
	BotChampionId                 int    `json:"botChampionId"`
	BotDifficulty                 string `json:"botDifficulty"`
	BotId                         string `json:"botId"`
	FirstPositionPreference       string `json:"firstPositionPreference"`
	IsBot                         bool   `json:"isBot"`
	IsLeader                      bool   `json:"isLeader"`
	IsSpectator                   bool   `json:"isSpectator"`
	Puuid                         string `json:"puuid"`
	Ready                         bool   `json:"ready"`
	SecondPositionPreference      string `json:"secondPositionPreference"`
	ShowGhostedBanner             bool   `json:"showGhostedBanner"`
	SummonerIconId                int    `json:"summonerIconId"`
	SummonerId                    int64  `json:"summonerId"`
	SummonerInternalName          string `json:"summonerInternalName"`
	SummonerLevel                 int    `json:"summonerLevel"`
	SummonerName                  string `json:"summonerName"`
	TeamId                        int    `json:"teamId"`
}

// MatchmakingSearchState represents the current matchmaking state
type MatchmakingSearchState struct {
	DodgeData struct {
		DodgerId int64  `json:"dodgerId"`
		State    string `json:"state"`
	} `json:"dodgeData"`
	Errors             []interface{} `json:"errors"`
	EstimatedQueueTime float64       `json:"estimatedQueueTime"`
	IsCurrentlyInQueue bool          `json:"isCurrentlyInQueue"`
	LobbyId            string        `json:"lobbyId"`
	LowPriorityData    struct {
		BustedLeaverAccessToken string  `json:"bustedLeaverAccessToken"`
		PenalizedSummonerIds    []int64 `json:"penalizedSummonerIds"`
		PenaltyTime             float64 `json:"penaltyTime"`
		PenaltyTimeRemaining    float64 `json:"penaltyTimeRemaining"`
		Reason                  string  `json:"reason"`
	} `json:"lowPriorityData"`
	ReadyCheck struct {
		DeclinerIds    []int64 `json:"declinerIds"`
		DodgeWarning   string  `json:"dodgeWarning"`
		PlayerResponse string  `json:"playerResponse"`
		State          string  `json:"state"`
		SuppressUx     bool    `json:"suppressUx"`
		Timer          float64 `json:"timer"`
	} `json:"readyCheck"`
	SearchState string  `json:"searchState"`
	TimeInQueue float64 `json:"timeInQueue"`
}
