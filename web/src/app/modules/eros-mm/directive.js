'use strict';

angular.module('erosApp.mm')

.directive('erosMm',function(){
	return {
		templateUrl: 'modules/eros-mm/mm.tpl.html',
		replace: true,
		scope: true,
		controller: 'MmCtrl',
		link: function($scope, $elem, $attrs, $controller){
			$scope.hover = $scope.statusToString(eros.matchmaking.status) != "IDLE"; // Full only when matchmaking IDLE

			$elem.mouseenter(function(){
				$scope.hover = true;
			});
			$elem.mouseleave(function(){
				if($scope.matchmaking.status == "IDLE"){
					$scope.hover = false;
				}
			});
		}
	};
})

.directive('bnetProfile', function(){
	return {
		templateUrl: 'modules/eros-mm/bnetprofile.tpl.html',
		replace: true,
		controller: 'MmCtrl',
		link: function($scope, $elem, $attrs, $controller){
			$scope.RequestVerification = function(){
				eros.matchmaking.request_verification(1)
			}
		}
	};
})

.filter('timer', function(){
	return function(input){
		if (typeof input === 'undefined') { return ; }
		var minutes = input.m > 9 ? input.m : "0"+input.m;
		var seconds = input.s > 9 ? input.s : "0"+input.s;

		return minutes + ":" + seconds;
	};
})

.filter('region', function(){
	return function(input){
		var regions = ["North America", "Europe", "Korea"];

		return regions[input-1];
	};
});
