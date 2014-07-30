'use strict';

/* Directives */


angular.module('erosApp.directives', []).
  directive('appVersion', ['version', function(version) {
    return function(scope, elm, attrs) {
      elm.text(version);
    };
  }])
	
	.directive('connGrowl', ['$rootScope', '$animate', function($rootScope, $animate){
		return {
			restrict: 'A',
			template: 	'<div class="conn-growl" ng-class="status"><div ng-repeat="message in messages" class="animate-fade">' +
						'	<div ng-bind="message" class="message"></div>' +
						'</div></div>',
			controller: ['$scope', '$timeout', function($scope, $timeout){
				$scope.messages = [];
				$scope.status = "";

				var addMessage = function(message){
					$scope.messages.push(message)

					$timeout(function () {
						$scope.deleteMessage(message);
					}, 2000);
				}

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
							$scope.status = "";
							break;
						case 1:
							$scope.status = "on";
							break;
						case 2:
							$scope.status = "off";
							break;
					}
				});
			}]
		}
	}])
