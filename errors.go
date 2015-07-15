package main

// type ErosError interface {
//  Error() string
//  ErrorCode() int
// }

type ErosError interface {
    Error() string
    Code() int
}

type ErosErrorImpl struct {
	desc string
	code int
}

func (err ErosErrorImpl) Error() string {
	return err.desc;
}

func (err ErosErrorImpl) Code() int {
	return err.code;
}

func Error(desc string, code int) ErosError {
	return ErosErrorImpl {
		desc: desc,
		code: code,
	}
}

// Errors
var (
    // 1xx - INTERNAL SERVER ERRORS
    ErrDatabaseRead                     = Error("Database read error", 101)
    ErrDatabaseWrite                    = Error("Database write error", 102)
	ErrDbInsert                         = Error("An error occured while writing to the database.", 102)
    ErrDiskRead                         = Error("Disk read error", 103)
    ErrDiskWrite                        = Error("Disk write error", 104)
    ErrAuthentication                   = Error("Authentication error", 105)
    ErrGeneric                          = Error("Generic error.", 106)
    ErrBadName                          = Error("Bad name.", 107)
    ErrNameInUse                        = Error("Name in use.", 108)
    ErrCannotWhileMatched               = Error("Cannot do that while matched via matchmaking.", 109)

    // 2xx - BATTLE.NET ERRORS
    ErrBadCharacterInfo                 = Error("Bad character info", 201)
    ErrCharacterAlreadyExists           = Error("Character already exists", 202)
    ErrCommunicatingWithBattleNet       = Error("Error while communicating with Battle.net", 203)
    ErrVerificationFailed               = Error("Verification failed.", 204)

    // 3xx - LADDER ERRORS
    ErrLadderErrorProcesingReplay       = Error("Error processing replay.", 301)
    ErrLadderErrorProcessingMatchResult = Error("Error while processing match result", 302)
    ErrLadderDuplicateReplay            = Error("The provided has been processed previously.", 303)
    ErrLadderClientNotInvolved          = Error("None of the client's registered characters were found in the replay participant list.", 304)
    ErrLadderGameTooShort               = Error("The provided game was too short.", 305)
    ErrLadderInvalidFormat              = Error("Matches must be a 1v1 with no observers.", 306)
    ErrLadderInvalidMap                 = Error("Matches must be on a valid map in the map pool.", 307)
    ErrLadderWrongMap                   = Error("The provided game was not on the correct map.", 307)
    ErrLadderInvalidMatchParticipents   = Error("All participents of a game must be registered.", 308)
    ErrLadderPlayerNotFound             = Error("The player was not found in the database.", 309)
    ErrLadderWrongOpponent              = Error("The provided game was not against your matchmade opponent. You have been forfeited.", 310)
    ErrLadderWrongSpeed                 = Error("The provided game was not on the Faster speed setting.", 311)
    ErrLadderVetoNotInRankedPool        = Error("Cannot add veto. Map not in ranked pool.", 312)
    ErrLadderAllVetoesUsed              = Error("Cannot add veto. Maximum number of vetoes used.", 313)
    ErrLadderGameNotPrearranged         = Error("The provided game was not arranged by the Eros matchmaker.", 314)

    // 4xx - MATCHMAKING ERRORS
    ErrNoCharacterInRegion              = Error("Can't queue on this region without a character on this region.", 401)
    ErrMatchmakingRequestCancelled      = Error("The matchmaking request was cancelled.", 402)
    ErrLongProcessUnavailable           = Error("Long process unavailable.", 403)

    // 5xx - CHAT ERRORS
    ErrChatRoomNotJoinable              = Error("Chat room not joinable.", 501)
    ErrChatBadPassword                  = Error("Bad password.", 502)
    ErrChatRoomAlreadyExists            = Error("The chat room name specified already exists.", 503)
    ErrChatRoomReserved                 = Error("The chat room name is reserved.", 504)
    ErrChatMaxChannelLimit              = Error("Can't join. Max channel limit reached.", 505)
    ErrChatNotOnChannel                 = Error("Can't send message. Not on channel.", 506)
    ErrChatUserOffline                  = Error("Can't send message. User offline.", 507)
    ErrChatMissingFields                = Error("Can't send message. Missing fields.", 508)
    ErrChatRoomNameTooShort             = Error("The chat room name is too short.", 509)
    ErrChatRateLimit                    = Error("Can't send message. Rate limit.", 510)
    ErrChatMessageTooLong               = Error("Can't send message. Message too long.", 511)
	ErrChatRoomNotFound                 = Error("Chat room not found.", 512)
)
