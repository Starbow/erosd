'use strict';

/* Directives */


angular.module('erosApp.directives', []).

directive('appVersion', ['version', function(version) {
	return function(scope, elm, attrs) {
	  elm.text(version);
	};
}])
	
// TODO: add unique
.directive('connGrowl', ['$rootScope', '$animate', function($rootScope, $animate){
	return {
		restrict: 'A',
		template: 	'<div class="conn-growl" ng-class="connStatus"><div ng-repeat="message in messages" class="animate-fade">' +
					'	<div ng-bind="message" class="message"></div>' +
					'</div></div>',
		controller: ['$scope', '$timeout', function($scope, $timeout){
			$scope.messages = [];
			$scope.connStatus = "";

			var addMessage = function(message){
				$scope.messages.push(message);

				$timeout(function () {
					$scope.deleteMessage(message);
				}, 2000);
			};

			$scope.deleteMessage = function (message) {
				var index = $scope.messages.indexOf(message);
				if (index > -1) {
					$scope.messages.splice(index, 1);
				}
			};

			$rootScope.$on("connGrowl", function (event, message, status) {
				addMessage(message);

				switch(status){
					case 0:
						$scope.connStatus = "";
						break;
					case 1:
						$scope.connStatus = "on";
						break;
					case 2:
						$scope.connStatus = "off";
						break;
				}
			});
		}]
	};
}]);

// .directive('browserid', ['browserid', function(eros_browserid){
// 	return {
// 		restrict: 'A',
// 		link: function($scope, $elem, $attr, $conn){
// 			eros_browserid.registerWatchHandlers().then(function() {
// 				$elem.on('click', function(e){
// 					e.preventDefault();
// 		            var $link = $(this);
// 		            eros_browserid.login().then(function(verifyResult) {
// 		            	console.log("Logged in, verifyResult:");
// 		            	console.log(verifyResult);
// 		                // window.location = $link.data('next') || verifyResult.redirect;
// 		            });
// 				});
// 			});
// 		}
// 	};
// }]);

