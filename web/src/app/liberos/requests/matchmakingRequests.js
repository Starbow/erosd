(function (global) {
    "use strict";

    var MatchmakingQueueRequest = function(regions, search_range){ // MMQ

    	queue_options = new protobufs.MatchmakingQueue(regions, search_range);
    	var request = new starbow.ErosRequests.Request("MMQ", queue_options, function(){
    		
    	})
    }

    global["starbow"]["ErosRequests"]["MatchmakingQueueRequest"] = MatchmakingQueueRequest;
})(this);