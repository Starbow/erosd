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

	// $scope.noshow_requested = false;
	// $scope.draw_requested = false;
	$scope.timeElapsed = {s: 0, m:0};
	$scope.longProcessTimer = {s: 0, m:0};
	// $scope.$parent.longProcessTimerResponse = {s: 0, m:0}
	$scope.timerInterval = [];

	$scope.selected_regions = {
		NA: false,
		EU: false
	};

	$scope.sel_regions = []

	$scope.mapPool = eros.ladder.maps;
	$scope.vetoedMaps = eros.ladder.vetoes;

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
					timer.stop('longProcessTimer');
					$scope.noshow_requested = false;
					$scope.draw_requested = false;
					notifier.title('','mm', false);
				} else if (value == eros.enums.MatchmakingState.Matched){
					$scope.hover = true;
					$scope.matchmaking.status = "MATCHED";

					timer.stop('timeElapsed');
					notifier.title('[Matched]','mm', true);
					notifier.matched();
					$scope.noshow_requested = false;
					$scope.draw_requested = false;
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
				$scope.selected_regions.NA = match.map.region == 1; 
				$scope.selected_regions.EU = match.map.region == 2;

				if (eros.chat.rooms[match.match_room]){
					eros.chat.rooms()[match.match_room].name = "MATCH " + eros.chat.rooms()[match.match_room].name;
				}
			});
		},

		update_longprocess: function(type){
			$scope.$apply(function(){
				if(type==eros.enums.LongProcess.NOSHOW){
					$scope.noshow_reponse = true;
					timer.timedown($scope,'longProcessResponseTimer',$scope.matchmaking.match.long_response_time.low);
				}else if(type==eros.enums.LongProcess.DRAW){
					$scope.draw_reponse = true;
					timer.timedown($scope,'longProcessResponseTimer',$scope.matchmaking.match.long_response_time.low);
				}else{
					$scope.noshow_reponse = false;
					$scope.noshow_requested = false;
					$scope.draw_reponse = false;
					$scope.draw_requested = false;
					timer.stop('longProcessResponseTimer');
				}
			});
		},

		update_characters: function(){
			$scope.$apply(function(){
				$scope.localUser.characters = eros.localUser.characters
				$scope.localUser.char_per_region = $scope.toRegions(eros.localUser.characters)
			})
		},

		update_maps: function(){
			$scope.$apply(function(){
				$scope.mapPool = eros.ladder.maps;
				$scope.vetoedMaps = eros.ladder.vetoes;
			});
		}
	});

	$scope.queue = function(){

		var regions = [];
		_.each($scope.selected_regions, function(value, key){
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

		$scope.upload_response = null;

		
		reader.onload = (function(read_file) {
			$scope.uploading_file = true;
			$scope.eros.matchmaking.upload_replay(read_file.target.result, function(result){
				$scope.uploading_file = false;
				$scope.upload_response = result
			});

		});
				
	};

	$scope.toggle_region = function(region){
		$scope.selected_regions[region] = !$scope.selected_regions[region];
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
			timer.timedown($scope,'longProcessTimer', $scope.matchmaking.match.long_response_time.low);
		});
	};

	$scope.respondNoShow= function(){
		eros.matchmaking.respond_noshow(function(){
			$scope.noshow_reponse = false;
			timer.stop('longProcessResponseTimer');
		});
	};

	$scope.requestDraw = function(){
		eros.matchmaking.request_draw(function(){
			$scope.draw_requested = true;
			timer.timedown($scope,'longProcessTimer', $scope.matchmaking.match.long_response_time.low);
		});
	}

	$scope.respondDraw = function(accept){
		eros.matchmaking.respond_draw(accept, function(){
			$scope.draw_reponse = false;
			timer.stop('longProcessResponseTimer');
		});
	}

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

	$scope.vetoModal = function(){
		$('#vetoesDialog').modal('show');
	}

	$scope.toggleVeto = function(map) {
		// map.vetoed = true;
		eros.matchmaking.toggleVeto(map)
	}
}])