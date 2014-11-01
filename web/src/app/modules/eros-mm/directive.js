'use strict';

angular.module('erosApp.mm', [])

.controller('MmCtrl', ['$scope', function($scope){
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

	$scope.regions = {
		NA: false,
		EU: true
	}

	$scope.queue = function(){

		var regions = [];
		_.each($scope.regions, function(value, key){
			if(value){
				switch(key) {
					case 'NA': regions.push(protobufs.Region.NA); break
					case 'EU': regions.push(protobufs.Region.EU); break
				}

			}
		})
		if(regions.length < 1){
			console.warn("No region selected.")
			return
		}

		var result = $scope.eros.matchmaking.queue(regions, $scope.search_radius);

		if(typeof result == undefined){

		}
	}

	$scope.dequeue = function(){

		$scope.eros.matchmaking.dequeue();
	}
}])

.directive('erosMm',function(){
	return {
		templateUrl: 'modules/eros-mm/mm.tpl.html',
		replace: true,
		scope: true,
		controller: 'MmCtrl'
	}
});
