(function (global) {
    "use strict";

    var Match = function(eros) {
    	// var opponent,
    	// 	match_time,
    	// 	battle_net_channel,
    	// 	map,
    	// 	region

    	// From protobufs
    	var timespan,
    		quality,
    		opponent,
    		opponent_latency,
    		channel,
    		chat_room,
    		map,
    		long_unlock_time,
    		long_response_time;

    	var match = this;
    };

    var MatchmakingModule = function(eros, sendRequest, options){
    	var status;
        var matchmaking = this;

    	function processServerMessage(command, payload) {
    		if (command == "MMQ" || command == "MMD"){
    			matchmaking.status = command == "MMQ" ? eros.enums.MatchmakingState.Queued : eros.enums.MatchmakingState.Idle;
    			match = undefined;

                matchmaking.controller.update_status(matchmaking.status);

    			return true;
    		}else if(command == "MMR"){
                matchmaking.status = eros.enums.MatchmakingState.Matched;

    			matchmaking.match = protobufs.MatchmakingResult.decode64(payload);
                matchmaking.match.opponent = eros.user(matchmaking.match.opponent.username);

                matchmaking.controller.update_status(eros.enums.MatchmakingState.Matched);
                matchmaking.controller.update_match(matchmaking.match);

    			return true;
    		}else if(command == "MMF"){
                matchmaking.status = eros.enums.MatchmakingState.Idle;
    			queued = false;
    			matched = false;
    			match = undefined;
                return true;
            }else if(command == "REP"){
                matchmaking.status = eros.enums.MatchmakingState.Idle;
                matchmaking.match = null;

                matchmaking.controller.update_status(eros.enums.MatchmakingState.Idle);
                matchmaking.controller.update_match(null);
                return true;

    		}else if(command=="MMI"){
                matchmaking.status = eros.enums.MatchmakingState.Idle;
                matchmaking.match = null;

                matchmaking.controller.update_status(eros.enums.MatchmakingState.Idle);
                return true;
            }else if(command=="LPF" || command=="LPD"){
                matchmaking.controller.update_longprocess(command=="LPF" ? eros.enums.LongProcess.NOSHOW : eros.enums.LongProcess.DRAW);
                return true;
            }else if(command=="LPR"){
                matchmaking.controller.update_longprocess();
                return true;
            }else if(command=="BNN"){
                var character = protobufs.Character.decode64(payload)
                eros.localUser.addCharacter(character)
                matchmaking.controller.update_characters()
                return true;
            }else if(command=="VET"){
                var vetoed_maps = protobufs.MapPool.decode64(payload).map;
                eros.vetoMaps(vetoed_maps);
                matchmaking.controller.update_maps();
                return true;
            }else{
    			return false;
    		}
    	}

    	this.commandHandlers = {
    		"MMQ": processServerMessage, // Queued
    		"MMD": processServerMessage, // Dequed
    		"MMR": processServerMessage, // Matched
    		"MMF": processServerMessage, // Forfeited
            "REP": processServerMessage, // Match accepted
            "MMI": processServerMessage, // Matchmaking Idle
            "RLP": processServerMessage, // LongProcess Request
            "LPR": processServerMessage, // LongProcess Response
            "LPF": processServerMessage, // LongProcess Forfeit
            "LPD": processServerMessage, // LongProcess Draw
            "BNN": processServerMessage, // 
            "VET": processServerMessage // Map Vetos
    	};

    	this.queue = function(regions, search_range){
            if(typeof regions != 'object'){
                console.error('[MatchmakingModule.queue] Region must be an object.');
                return;
            }

            var request = new starbow.ErosRequests.MatchmakingQueueRequest(regions, search_range, function(success, command, match){
                if(success){
                    if(command == "MMQ"){
                        matchmaking.controller.update_status(eros.enums.MatchmakingState.Queued);
                    }else if(command == "MMR"){
                        matchmaking.match = match;
                        matchmaking.match.opponent = eros.user(match.opponent.username);

                        matchmaking.controller.update_status(eros.enums.MatchmakingState.Matched);
                        matchmaking.controller.update_match(matchmaking.match);
                    }
                }else{
                    // Need error handler
                    console.warn("Error "+command+": "+eros.locale.Error[command]);
                }
            });
    		sendRequest(request);
    	};

        this.dequeue = function(){
            var request = new starbow.ErosRequests.MatchmakingDequeueRequest(function(success, command){
                if(success){
                    matchmaking.controller.update_status(eros.enums.MatchmakingState.Idle);
                }else{
                    // Need error handler
                    console.warn("Error "+command+": "+eros.locale.Error[command]);
                }
            });
            sendRequest(request);
        };

        this.request_forfeit = function(){
            var request =  new starbow.ErosRequests.MatchmakingForfeitRequest(function(success,command){
                if(success){
                    console.info("Forfeit success.");
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.upload_replay = function(file) {
            var request = new starbow.ErosRequests.MatchmakingUploadReplayRequest(file, function(success, request){
                if(success){
                    if(request.command == "REP"){
                        console.log("[MM] Upload success.");

                        matchmaking.controller.update_status(eros.enums.MatchmakingState.Idle);
                        matchmaking.match = null;
                        // options.update_match(null)
                    }
                }else{
                    // Need error handler
                    // console.warn("Error "+command+": "+eros.locale.Error[command])
                    console.warn("Error "+command);
                }
            });
            console.log("[MM] Uploading replay.");
            sendRequest(request);
        };

        this.request_noshow = function(callback){
            console.log("Requesting no show.");
            var request =  new starbow.ErosRequests.MatchmakingLongProcessRequest(eros.enums.LongProcess.NOSHOW, function(success,command){
                if(success){
                    if(typeof callback == 'function'){
                        callback();
                    }
                    
                    console.info("No-show request success.");
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.request_draw = function(callback){
            console.log("Requesting long process");
            var request =  new starbow.ErosRequests.MatchmakingLongProcessRequest(eros.enums.LongProcess.DRAW, function(success,command){
                if(success){
                    if(typeof callback == 'function'){
                        callback();
                    }
                    console.info("Draw request success.");
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.respond_noshow = function(callback){
            var request =  new starbow.ErosRequests.MatchmakingLongProcessResponse(0, function(success,command){
                if(success){
                    if(typeof callback == 'function'){
                        callback();
                    }
                    
                    console.info("Respond no-show success.");
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.respond_draw = function(accept, callback){
            var request =  new starbow.ErosRequests.MatchmakingLongProcessResponse(accept, function(success,command){
                if(success){
                    if(typeof callback == 'function'){
                        callback();
                    }
                    
                    console.info("Respond draw success.");
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.request_verification = function(region,callback){
            var request =  new starbow.ErosRequests.OAuthVerificationRequest(region, function(success,command, payload){
                if(success){
                    if(typeof callback == 'function'){
                        callback();
                    }
                    var response = protobufs.OAuthUrl.decode64(payload);
                    console.debug("Request url: "+response.url);
                    window.open(response.url)
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        };

        this.request_remove_character = function(character, callback){
            var regions = ["NA", "EU", "KR"]

            var proto_char = {
                region: protobufs.Region[regions[character.region-1]],
                subregion: character.subregion,
                profile_id: character.profile_id,
                character_name: character.character_name
            }

            var request = new starbow.ErosRequests.RemoveCharacterRequest(proto_char, function(success, command, payload){
                if(success){
                    eros.localUser.removeCharacter(character)
                    matchmaking.controller.update_characters()

                    if(typeof callback == 'function'){
                        callback();
                    }
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            });
            sendRequest(request);
        }

        this.toggleVeto = function(map){
            var request = new starbow.ErosRequests.ToggleVetoRequest(map, function(success, command, payload){
                if(success){
                    var vetoed_maps = protobufs.MapPool.decode64(payload).map;
                    eros.vetoMaps(vetoed_maps);
                    matchmaking.controller.update_maps();
                }else{
                    // Need error handler
                    console.warn("Error "+command);
                }
            })
            sendRequest(request);
        }
    };

    global.starbow.Eros.prototype.modules.matchmaking = MatchmakingModule;

})(this);