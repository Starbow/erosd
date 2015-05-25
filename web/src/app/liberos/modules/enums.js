(function (global) {
    "use strict";

    var EnumsModule = function(){
    	this.Region = {
			NA: 1,
			EU: 2,
			KR: 3,
			CN: 5,
			SEA: 6,
		};

    	this.MatchmakingState = {
    		Idle: 0,
			Queued: 1,
			Matched: 2,
			InvalidRegion: 401,
			Aborted: 402,
    	}

    	this.Error = {
    		None: 0,

			DatabaseReadError: 101,
			DatabaseWriteError: 102,
			DiskReadError: 103,
			DiskWriteError: 104,
			AuthenticationError: 105,
			GenericError: 106,
			BadName: 107,
			NameInUse: 108,
			CannotPerformActionWhileMatchmaking: 109,

			BadCharacterInfo: 201,
			CharacterExists: 202,
			BattleNetCommunicationError: 203,
			VerificationFailed: 204,

			ReplayProcessingError: 301,
			MatchProcessingError: 302,
			DuplicateReplay: 303,
			ClientNotInvolvedInMatch: 304,
			GameTooShort: 305,
			BadFormat: 306,
			BadMap: 307,
			InvalidParticipants: 308,
			PlayerNotInDatabase: 309,
			NotAssignedOpponent: 310,
			BadSpeed: 311,
			CannotVetoUnrankedMap: 312,
			MaxVetoesReached: 313, // Put your motherf**kin' hands up and follow me
			GameNotPrearranged: 314,

			NoCharacterForRegion: 401,
			MatchmakingAborted: 402,
			LongProcessRequestFailed: 403,

			RoomNotJoinable: 501,
			BadPassword: 502,
			RoomAlreadyExists: 503,
			RoomReserved: 504,
			MaximumRoomLimitReached: 505,
			NotOnChannel: 506,
			UserNotFound: 507,
			BadMessage: 508,
			NameTooShort: 509,
			RateLimited: 511,
    	}

    	this.LongProcess = {
    		// Hex to base64 (01 and 02)
    		NOSHOW: "AQ==",
    		DRAW: "Ag=="
    	}

    	this.ErrorDescriptor = {
		    301: "Error processing replay",
		    302: "Error while processing match result",
		    303: "Duplicate Replay",
		    304: "The submitting client was not involved in the match.",
		    305: "Game too short.",
		    306: "Bad format. Required 1v1 with no observers.",
		    307: "Bad map. Require a map in the map pool.",
		    308: "All participants of the game must be registered.",
		    309: "Player not found in database.",
		    310: "You didn't play your matchmade opponent. You have been forfeited from that game.",
		    311: "The game was not played on Faster.",
		    312: "Cannot add veto. Map not in ranked pool.",
		    313: "Cannot add veto. Maximum number of vetoes used.",
		    314: "You are not in a game arranged by the Eros matchmaker."
    	}
    }

    global.starbow.Eros.prototype.modules.enums = EnumsModule;


})(this);