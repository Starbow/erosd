'use strict';

angular.module('erosApp.mm', [])

.controller('MmCtrl', ['$scope', 'timer','notifier',function($scope, timer, notifier){
	$scope.localUser = eros.localUser;

	$scope.matchtypes = {};
	$scope.matchtypes.one = true;
	$scope.matchtypes.two = false;
	$scope.matchtypes.bgh = false;
	$scope.matchtypes.ffa = false;

	// $scope.matchmaking = {}
	$scope.matchmaking.status = "IDLE";

	$scope.search_radius = $scope.eros.localUser.stats.search_radius;
	$scope.radius_options = [1,2,3,4,5];

	$scope.noshow_requested = false;
	$scope.timeElapsed = {s: 0, m:0};
	$scope.noShowTimer = {s: 0, m:0};
	// $scope.$parent.noShowTimerResponse = {s: 0, m:0}
	$scope.timerInterval = [];

	$scope.regions = {
		NA: false,
		EU: true
	};

	$scope.uploadreplay = false;

	eros.registerController('matchmaking',{
		update_status: function(value){
			$scope.$apply(function(){
				if(value == eros.enums.MatchmakingState.Queued){
					$scope.matchmaking.status = "QUEUED";
					$scope.hover = true;
				} else if (value == eros.enums.MatchmakingState.Idle){
					$scope.matchmaking.status = "IDLE";

					timer.stop('timeElapsed');
					timer.stop('noShowTimer');
					$scope.noshow_requested = false;
					notifier.title('','mm', false);
				} else if (value == eros.enums.MatchmakingState.Matched){
					$scope.hover = true;
					$scope.matchmaking.status = "MATCHED";

					timer.stop('timeElapsed');
					notifier.title('[Matched]','mm', true);
					notifier.matched();
					$scope.noshow_requested = false;
					$scope.uploadreplay = false;

					$('[eros-mm]').on('mouseover', function(){
						notifier.title('','mm', true);
						$('[eros-mm]').off('mouseover');
					});
				}
			});	
		},
		update_match: function(match){
			$scope.$apply(function(){
				$scope.matchmaking.match = match;
				$scope.regions.NA = match.map.region == 1; 
				$scope.regions.EU = match.map.region == 2;

				if (eros.chat.rooms[match.match_room]){
					eros.chat.rooms()[match.match_room].name = "MATCH " + eros.chat.rooms()[match.match_room].name;
				}
			});
		},

		update_longprocess: function(type){
			$scope.$apply(function(){
				if(type==eros.enums.LongProcess.NOSHOW){
					$scope.noshow_reponse = true;
					timer.timedown($scope,'noShowResponseTimer',$scope.matchmaking.match.long_response_time.low);
				}else if(type==eros.enums.LongProcess.DRAW){
					$scope.draw_reponse = true;
					timer.timedown($scope,'noShowResponseTimer',$scope.matchmaking.match.long_response_time.low);
				}else{
					$scope.noshow_reponse = false;
					$scope.noshow_requested = false;
					timer.stop('noShowResponseTimer');
				}
			});
		},

		update_characters: function(){
			$scope.$apply(function(){
				$scope.localUser.characters = eros.localUser.characters
				$scope.localUser.char_per_region = $scope.toRegions(eros.localUser.characters)
			})
		}
	});

	$scope.queue = function(){

		var regions = [];
		_.each($scope.regions, function(value, key){
			if(value){
				regions.push(protobufs.Region[key]);
				// switch(key) {
				// 	case 'NA': regions.push(protobufs.Region.NA); break
				// 	case 'EU': regions.push(protobufs.Region.EU); break
				// }
			}
		});
		if(regions.length < 1){
			console.warn("No region selected.");
			return;
		}

		var result = $scope.eros.matchmaking.queue(regions, $scope.search_radius);
		timer.restart($scope,'timeElapsed');
	};

	$scope.dequeue = function(){
		$scope.eros.matchmaking.dequeue();
	};

	$scope.upload_replay = function(){
		var file = document.getElementById('file').files[0];
		var reader = new FileReader();
		// reader.readAsBinaryString(file);
		reader.readAsDataURL(file);

		reader.onload = (function(read_file) {
			$scope.eros.matchmaking.upload_replay(read_file.target.result);
		});
	};

	$scope.toggle_region = function(region){
		$scope.regions[region] = !$scope.regions[region];
	};

	$scope.updateSearchRadius = function(){
		if(typeof $scope.search_radius !== 'number'){
			$scope.search_radius = $scope.eros.localUser.stats.search_radius;
			return false
		}

		if($scope.search_radius > 5){
			$scope.search_radius = 5
		}else if($scope.search_radius < 1){
			$scope.search_radius = 1
		}else{
			// Update stats (TODO)
			$scope.eros.localUser.stats.search_radius = $scope.search_radius;
		}
	}


	$scope.copyChat = function() {
  		window.prompt("Copy to clipboard: Ctrl+C, Enter", $scope.matchmaking.match.channel);
	};

	$scope.goToMap = function(){
		var map = $scope.matchmaking.match.map;
		console.debug('Opening map: starcraft://map/'+map.region+"/"+map.battle_net_id);
		window.open('starcraft://map/'+map.region+"/"+map.battle_net_id);
	};

	$scope.forfeit = function(){
		$scope.eros.matchmaking.request_forfeit();
	};

	$scope.reportNoShow=function(){
		eros.matchmaking.request_noshow(function(){
			$scope.noshow_requested = true;
			timer.timedown($scope,'noShowTimer', $scope.matchmaking.match.long_response_time.low);
		});
	};

	$scope.respondNoShow= function(){
		eros.matchmaking.respond_noshow(function(){
			$scope.noshow_reponse = false;
			timer.stop('noShowResponseTimer');
		});
	};

	$scope.statusToString = function(value){
		if(typeof value === 'undefined'){
			return "IDLE";
		}
		if(value == eros.enums.MatchmakingState.Queued){
			return "QUEUED";
		} else if (value == eros.enums.MatchmakingState.Idle){
			return "IDLE";
		} else if (value == eros.enums.MatchmakingState.Matched){
			return "MATCHED";
		}
	};

	$scope.toRegions = function(input){
		if(typeof input === 'object'){
			var regions = []
			for(var i = 0; i < input.length; i++){
				if(typeof regions[input[i].region] === "undefined"){
					regions[input[i].region] = []
				}
				regions[input[i].region].push(input[i])
			}

			console.log(regions)
			return regions;
		}

		return null;		
	}

	$scope.removeChar = function(character){
		$scope.char_to_remove = character;
		$('#confirmRemoveChar').modal('show')
		delete $scope.char_remove_status
		delete $scope.char_remove_completed
	}

	$scope.requestCharRemove = function(){
		$scope.char_remove_status = "Deleting..."
		$scope.char_remove_completed = false
		eros.matchmaking.request_remove_character($scope.char_to_remove,function(){
			$scope.char_remove_status = "Completed!"
			$scope.char_remove_completed = true
			setTimeout(1000, function(){$('#confirmRemoveChar').modal('show')})
		})
	}
}])