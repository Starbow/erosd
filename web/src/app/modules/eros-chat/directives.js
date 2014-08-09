
angular.module('erosApp.chat', [])
.directive('erosChat', function(){
	return {
		restrict: 'A',
		templateUrl: 'modules/eros-chat/chat.tpl.html'
	}
})
.directive('roomUsers', ['$rootScope', '$animate', function($rootScope, $animate){
	return{
		restrict: 'A',
		template: 	'<div class="room-users"><div ng-repeat="user in users" class="animate-fade">' +
					'	<div ng-bind="user.username" class="room-user"></div>' +
					'</div></div>',
		controller: ['$scope', function($scope){
			$scope.room = ''
			$scope.users;
		}],
		link: function($scope, $elem, $attrs, $controller){
			$scope.$watch("$parent.selectedRoom", function(room){
				if(room){
					$scope.users=room.room.users()
				}
			})
			
		}

	}
}])
.directive('username', function(){
	return {
		restrict: 'A',
		template: 'template',
		link: function($scope, $elem, $attrs, $controller){

		}
	}
})
.directive('chatMessage', function(){
	return {
		restrict: 'A',
		link: function($scope, $elem, $attrs, $controller){
			$elem.html($scope.message.message.replace(/@\w+/g, '<span username>'+'$&'+'</span>'))
		}
	}
})
;