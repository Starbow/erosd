'use strict';

angular.module('erosApp.mm')

.controller('MmCtrl', ['$scope', 'timer','notifier',function($scope, timer, notifier){
	$scope.localUser = eros.localUser;

	$scope.matchtypes = {}
	$scope.matchtypes.one = true;
	$scope.matchtypes.two = false;
	$scope.matchtypes.bgh = false;
	$scope.matchtypes.ffa = false;

	// $scope.matchmaking = {}
	$scope.matchmaking.status = "IDLE"

	$scope.search_radius = $scope.eros.localUser.stats.search_radius;
	$scope.radius_options = [1,2,3,4,5]

	$scope.noshow_requested = false
	$scope.timeElapsed = {s: 0, m:0}
	$scope.noShowTimer = {s: 0, m:0}
	// $scope.$parent.noShowTimerResponse = {s: 0, m:0}
	$scope.timerInterval = []

	$scope.regions = {
		NA: false,
		EU: true
	}

	$scope.uploadreplay = false;

	eros.registerController('matchmaking',{
		update_status: function(value){
			$scope.$apply(function(){
				if(value == eros.enums.MatchmakingState.Queued){
					$scope.matchmaking.status = "QUEUED"
					$scope.hover = true
				} else if (value == eros.enums.MatchmakingState.Idle){
					$scope.matchmaking.status = "IDLE"

					timer.stop('timeElapsed')
					timer.stop('noShowTimer')
					$scope.noshow_requested = false
					notifier.title('','mm', false)
				} else if (value == eros.enums.MatchmakingState.Matched){
					$scope.hover = true
					$scope.matchmaking.status = "MATCHED"

					timer.stop('timeElapsed')
					notifier.title('[Matched]','mm', true)
					notifier.matched
					$scope.noshow_requested = false
					$scope.uploadreplay = false

					$('[eros-mm]').on('mouseover', function(){
						notifier.title('','mm', true)
						$('[eros-mm]').off('mouseover')
					})
				}
			})
			
		},
		update_match: function(match){
			$scope.$apply(function(){
				$scope.matchmaking.match = match;
				$scope.regions.NA = match.map.region == 1 
				$scope.regions.EU = match.map.region == 2
			})
		},

		update_longprocess: function(type){
			$scope.$apply(function(){
				if(type==eros.enums.LongProcess.NOSHOW){
					$scope.noshow_reponse = true
					timer.timedown($scope,'noShowResponseTimer',$scope.matchmaking.match.long_response_time.low)
				}else if(type==eros.enums.LongProcess.DRAW){
					$scope.draw_reponse = true
					timer.timedown($scope,'noShowResponseTimer',$scope.matchmaking.match.long_response_time.low)
				}else{
					$scope.noshow_reponse = false
					$scope.noshow_requested = false
					timer.stop('noShowResponseTimer')
				}
				
			})
		}
	})

	$scope.queue = function(){

		var regions = [];
		_.each($scope.regions, function(value, key){
			if(value){
				regions.push(protobufs.Region[key])
				// switch(key) {
				// 	case 'NA': regions.push(protobufs.Region.NA); break
				// 	case 'EU': regions.push(protobufs.Region.EU); break
				// }
			}
		})
		if(regions.length < 1){
			console.warn("No region selected.")
			return
		}

		var result = $scope.eros.matchmaking.queue(regions, $scope.search_radius);
		timer.restart($scope,'timeElapsed');

		if(typeof result == undefined){

		}
	}

	$scope.dequeue = function(){
		$scope.eros.matchmaking.dequeue();
	}

	$scope.upload_replay = function(){
		var file = document.getElementById('file').files[0]
		var reader = new FileReader();
		// reader.readAsBinaryString(file);
		reader.readAsDataURL(file);

		reader.onload = (function(read_file) {
			$scope.eros.matchmaking.upload_replay(read_file.target.result)
		})
	}

	$scope.toggle_region = function(region){
		$scope.regions[region] = !$scope.regions[region]
	}


	$scope.copyChat = function() {
  		window.prompt("Copy to clipboard: Ctrl+C, Enter", $scope.matchmaking.match.channel);
	}

	$scope.goToMap = function(){
		var map = $scope.matchmaking.match.map
		console.log('Opening map: starcraft://map/'+map.region+"/"+map.battle_net_id)
		window.open('starcraft://map/'+map.region+"/"+map.battle_net_id)
	}

	$scope.forfeit = function(){
		$scope.eros.matchmaking.request_forfeit()
	}

	$scope.reportNoShow=function(){
		eros.matchmaking.request_noshow(function(){
			$scope.noshow_requested = true
			timer.timedown($scope,'noShowTimer', $scope.matchmaking.match.long_response_time.low)
		})
	}

	$scope.respondNoShow= function(){
		eros.matchmaking.respond_noshow(function(){
			$scope.noshow_reponse = false;
			timer.stop('noShowResponseTimer')
		})
	}

	$scope.statusToString = function(value){
		if(typeof value === 'undefined'){
			return "IDLE"
		}
		if(value == eros.enums.MatchmakingState.Queued){
			return "QUEUED"
		} else if (value == eros.enums.MatchmakingState.Idle){
			return "IDLE"
		} else if (value == eros.enums.MatchmakingState.Matched){
			return "MATCHED"
		}
	}
}])

.directive('erosMm',function(){
	return {
		templateUrl: 'modules/eros-mm/mm.tpl.html',
		replace: true,
		scope: true,
		controller: 'MmCtrl',
		link: function($scope, $elem, $attrs, $controller){
			$scope.hover = $scope.statusToString(eros.matchmaking.status) != "IDLE"; // Full only when matchmaking IDLE

			$elem.mouseenter(function(){
				$scope.hover = true
			})
			$elem.mouseleave(function(){
				if($scope.matchmaking.status == "IDLE"){
					$scope.hover = false
				}
			})
		}
	}
})

.filter('timer', function(){
	return function(input){
		if (typeof input === 'undefined') { return ; }
		var minutes = input.m > 9 ? input.m : "0"+input.m;
		var seconds = input.s > 9 ? input.s : "0"+input.s;

		return minutes + ":" + seconds;
	}
})

.filter('region', function(){
	return function(input){
		var regions = ["North America", "Europe", "Korea"]

		return regions[input-1]
	}
})
