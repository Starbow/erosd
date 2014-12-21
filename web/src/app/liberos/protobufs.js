var protobufs = dcodeIO.ProtoBuf.newBuilder().import({
    "package": "protobufs",
    "messages": [
        {
            "name": "Handshake",
            "fields": [
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "username",
                    "id": 1
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "auth_key",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "Division",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "id",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "name",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "promotion_threshold",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "demotion_threshold",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "icon_url",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "small_icon_url",
                    "id": 6
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "HandshakeResponse",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "HandshakeStatus",
                    "name": "status",
                    "id": 1
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "UserStats",
                    "name": "user",
                    "id": 2
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "int64",
                    "name": "id",
                    "id": 3
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Character",
                    "name": "character",
                    "id": 4
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Division",
                    "name": "division",
                    "id": 5
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Region",
                    "name": "active_region",
                    "id": 6
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "MapPool",
                    "name": "map_pool",
                    "id": 7
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "int64",
                    "name": "max_vetoes",
                    "id": 8
                }
            ],
            "enums": [
                {
                    "name": "HandshakeStatus",
                    "values": [
                        {
                            "name": "FAIL",
                            "id": 0
                        },
                        {
                            "name": "SUCCESS",
                            "id": 1
                        },
                        {
                            "name": "ALREADY_LOGGED_IN",
                            "id": 2
                        }
                    ],
                    "options": {}
                }
            ],
            "messages": [],
            "options": {}
        },
        {
            "name": "UserRegionStats",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "points",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "wins",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "losses",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "forfeits",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "walkovers",
                    "id": 6
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "mmr",
                    "id": 7
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "placements_remaining",
                    "id": 8
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "division",
                    "id": 9
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "division_rank",
                    "id": 10
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "UserStats",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "username",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "search_radius",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "points",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "wins",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "losses",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "forfeits",
                    "id": 6
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "walkovers",
                    "id": 7
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "UserRegionStats",
                    "name": "region",
                    "id": 8
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Map",
                    "name": "vetoes",
                    "id": 9
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "id",
                    "id": 10
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "mmr",
                    "id": 11
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "placements_remaining",
                    "id": 12
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "division",
                    "id": 13
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "division_rank",
                    "id": 14
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "MapPool",
            "fields": [
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Map",
                    "name": "map",
                    "id": 1
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "Map",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "battle_net_name",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int32",
                    "name": "battle_net_id",
                    "id": 3
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "description",
                    "id": 4
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "info_url",
                    "id": 5
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "preview_url",
                    "id": 6
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "SimulationResult",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "UserStats",
                    "name": "opponent",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "victory",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "match_quality",
                    "id": 3
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "MatchmakingQueue",
            "fields": [
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "radius",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "MatchmakingResult",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "timespan",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "double",
                    "name": "quality",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "UserStats",
                    "name": "opponent",
                    "id": 3
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "int64",
                    "name": "opponent_latency",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "channel",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "chat_room",
                    "id": 6
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "Map",
                    "name": "map",
                    "id": 7
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "long_unlock_time",
                    "id": 8
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "long_response_time",
                    "id": 9
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatRoomInfo",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "key",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "name",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "passworded",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "joinable",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "fixed",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "users",
                    "id": 6
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "UserStats",
                    "name": "participant",
                    "id": 7
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "forced",
                    "id": 8
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatRoomIndex",
            "fields": [
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "ChatRoomInfo",
                    "name": "room",
                    "id": 1
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatMessage",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "sender",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "target",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "message",
                    "id": 3
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatRoomMessage",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "ChatRoomInfo",
                    "name": "room",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "UserStats",
                    "name": "sender",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "message",
                    "id": 3
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatPrivateMessage",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "UserStats",
                    "name": "sender",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "message",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatRoomUser",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "ChatRoomInfo",
                    "name": "room",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "UserStats",
                    "name": "user",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ChatRoomRequest",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "room",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "password",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "MatchmakingStats",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "searching_users",
                    "id": 2
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "ServerStats",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "active_users",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "searching_users",
                    "id": 2
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "MatchmakingStats",
                    "name": "region",
                    "id": 3
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "Character",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int32",
                    "name": "subregion",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int32",
                    "name": "profile_id",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "character_name",
                    "id": 4
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "int32",
                    "name": "character_code",
                    "id": 5
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "profile_link",
                    "id": 6
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "string",
                    "name": "ingame_profile_link",
                    "id": 7
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "bool",
                    "name": "verified",
                    "id": 8
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "int32",
                    "name": "verification_portrait",
                    "id": 9
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "OAuthRequest",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                }
            ],
            "enums": [],
            "messages": [],
            "options": {},
            "oneofs": {}
        },
        {
            "name": "OAuthUrl",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "url",
                    "id": 1
                }
            ],
            "enums": [],
            "messages": [],
            "options": {},
            "oneofs": {}
        },
        {
            "name": "MatchParticipant",
            "fields": [
                {
                    "rule": "optional",
                    "options": {},
                    "type": "UserStats",
                    "name": "user",
                    "id": 1
                },
                {
                    "rule": "optional",
                    "options": {},
                    "type": "Character",
                    "name": "character",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "points_before",
                    "id": 3
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "points_after",
                    "id": 4
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int64",
                    "name": "points_difference",
                    "id": 5
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "bool",
                    "name": "victory",
                    "id": 6
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "race",
                    "id": 7
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "MatchResult",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "Region",
                    "name": "region",
                    "id": 1
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "Map",
                    "name": "map",
                    "id": 2
                },
                {
                    "rule": "repeated",
                    "options": {},
                    "type": "MatchParticipant",
                    "name": "participant",
                    "id": 3
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        },
        {
            "name": "BroadcastAlert",
            "fields": [
                {
                    "rule": "required",
                    "options": {},
                    "type": "string",
                    "name": "message",
                    "id": 2
                },
                {
                    "rule": "required",
                    "options": {},
                    "type": "int32",
                    "name": "predefined",
                    "id": 1
                }
            ],
            "enums": [],
            "messages": [],
            "options": {}
        }
    ],
    "enums": [
        {
            "name": "Region",
            "values": [
                {
                    "name": "NA",
                    "id": 1
                },
                {
                    "name": "EU",
                    "id": 2
                },
                {
                    "name": "KR",
                    "id": 3
                },
                {
                    "name": "CN",
                    "id": 5
                },
                {
                    "name": "SEA",
                    "id": 6
                }
            ],
            "options": {}
        }
    ],
    "imports": [],
    "options": {},
    "services": []
}).build("protobufs");
