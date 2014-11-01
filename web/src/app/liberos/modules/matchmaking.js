(function (global) {
    "use strict";

    var Match = function(eros){
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
    		long_response_time

    	var match = this;
    }

    var MatchmakingModule = function(eros, sendRequest, options){
    	var status;
        var matchmaking = this;

    	function processServerMessage(command, payload) {
    		if (command == "MMQ" || command == "MMD"){
    			this.status = command == "MMQ" ? eros.enums.MatchmakingState.Queued : eros.enums.MatchmakingState.Idle;
    			match = undefined;

    			options.update_status(this.status)

    			return true;
    		}else if(command == "MMR"){
                console.log("Matched!")
    			matchmaking.match = protobufs.MatchmakingResult.decode64(payload)

                options.update_status(eros.enums.MatchmakingState.Matched)
                options.update_match(matchmaking.match)

    			return true;
    		}else if(command == "MMF"){
    			queued = false;
    			matched = false;
    			match = undefined;

    		}else{
    			return false;
    		}
    	}

    	this.commandHandlers = {
    		"MMQ": processServerMessage, // Queued
    		"MMD": processServerMessage, // Dequed
    		"MMR": processServerMessage, // Matched
    		"MMF": processServerMessage, // Forfeited
    	}

    	this.queue = function(regions, search_range){
            if(typeof regions != 'object'){
                console.error('[MatchmakingModule.queue] Region must be an object.')
                return
            }
            var request = new starbow.ErosRequests.MatchmakingQueueRequest(regions, search_range, function(success, command, request){
                if(success){
                    if(command == "MMQ"){
                        options.update_status(eros.enums.MatchmakingState.Queued)
                    }else if(command == "MMR"){
                        options.update_status(eros.enums.MatchmakingState.Matched)
                        options.update_match(match)

                        matchmaking.match = protobufs.MatchmakingResult.decode64(request.payload)
                    }
                }else{
                    // Need error handler
                    console.warn("Error "+command+": "+eros.locale.Error[command])
                }
            })
            console.log("Requesting queue.")
    		sendRequest(request)
    	}

        this.dequeue = function(){
            var request = new starbow.ErosRequests.MatchmakingDequeueRequest(function(success, command){
                if(success){
                    options.update_status(eros.enums.MatchmakingState.Idle)
                }else{
                    // Need error handler
                    console.warn("Error "+command+": "+eros.locale.Error[command])
                }
            })
            sendRequest(request)
        }
    }

    global.starbow.Eros.prototype.modules.matchmaking = MatchmakingModule;

})(this);