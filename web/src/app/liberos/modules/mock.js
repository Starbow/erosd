(function (global) {
    "use strict";

    var MockModule = function(){

    	this.generateUsers = function(count){
    		count = count > 0 ? count : 20;

    		var rooms = eros.chat.rooms()
    		var room = rooms[Object.keys(rooms)[0]]
            var name_possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    		for(var i = 0; i < count; i++){
                var placement_matches =0,
                    x, 
                    points = 0,
                    region = 1, 
                    wins = 0,
                    loses = 0, 
                    id = 0, 
                    mmr = 1500,
                    division = 0,
                    rank = 0,
                    name = '',
                    name_length = 0;

                if (Math.random() > 0.9){ // Placement
                    placement_matches = Math.ceil(Math.random()*5);
                    wins = Math.floor(Math.random()*(5-placement_matches));
                    loses = Math.ceil(Math.random()*(5-placement_matches-wins));
                    mmr = 1400 + Math.random()*200;

                    division = 0;
                }else{
                    x = Math.random()
                    points = Math.round(x*Math.exp(-7*x)*10*20000);

                    wins = Math.round((-Math.random()^2 + Math.random())*Math.random()*600);
                    loses = Math.round((-Math.random()^2 + Math.random())*Math.random()*600);
                    
                    x = Math.random()
                    var mmr = Math.round(x*Math.exp(-7*x)*19*3000+Math.random()*500);

                    division = getDivision(points);
                    rank = Math.round(Math.random()*30);
                }

                id = 100000+i;
                region = Math.ceil(Math.random()*2);

                // Name
                name_length = Math.random()*15;
                for (var n = 0; n < name_length; n++){
                    name += name_possible.charAt(Math.floor(Math.random() * name_possible.length));
                }

                var region_stats = new protobufs.UserRegionStats(region, points, wins, loses, 2, 11, mmr, placement_matches, division, rank);
    			var user = new protobufs.UserStats(name, 3, points, wins, loses, 2, 11, region_stats, [], id, mmr, placement_matches, division, rank);
    			
    			var chatRoomInfo = new protobufs.ChatRoomInfo(room.key, room.name, room.passworded, room.joinable, room.fixed, Object.keys(room.users()).length, user, room.forced);
    			var chatUser = new protobufs.ChatRoomUser(chatRoomInfo, user);
    			eros.processServerMessage('CHJ', chatUser.toBase64())
    		}
    		
    	}

        var getDivision = function(points){
            if(points < 400){
                return 1
            } else if(points < 3000){
                return 2
            } else if(points < 6000){
                return 3
            } else if(points < 9000){
                return 4
            } else {
                return 5
            }
        }
    }

    global.starbow.Eros.prototype.modules.mock = MockModule;
})(this);