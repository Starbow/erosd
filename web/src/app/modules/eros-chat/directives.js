'use strict';

/* Chat directives */

angular.module('erosApp.chat')
.directive('erosChat', ['$rootScope','notifier', function($rootScope, notifier){
	return {
		restrict: 'A',
		templateUrl: 'modules/eros-chat/chat.tpl.html',
		controller: 'ChatCtrl'
	};
}])
.directive('roomUsers', ['$rootScope', '$animate', function($rootScope, $animate){
	return{
		restrict: 'AC',
		controller: ['$scope', function($scope){
			// $scope.room = ''
			$scope.users;
		}],
		link: function($scope, $elem, $attrs, $controller){
			$rootScope.$on('chat_room', function(){
				var room = $scope.$parent.selectedRoom.room;
				if(room){
					if(typeof(room.users) === "function"){
						$scope.roomusers=room.users(), 'stats.division';
					}else{
						$scope.roomusers=[];
					}
				}
			});
		}

	};
}])
.directive('username', function(){
	return {
		restrict: 'A',
		template: 'template',
		link: function($scope, $elem, $attrs, $controller){

		}
	};
})
.directive('chatMessage', function(){
	return {
		restrict: 'A',
		link: function($scope, $elem, $attrs, $controller){
			$elem.html($scope.message.message
				.replace(/@\w+/g, '<span username>'+'$&'+'</span>')
				.replace(/http\S*/g, function(match){
					var rep = '<a href="'+match+'" target="_blank">';
					rep += match.length > 40 ? match.slice(0,30)+'...'+match.slice(-6) : match;
					rep += '</a>'

					return rep;
				})
			);
		}
	};
})
.directive('erosUser', function(){
	return {
		restrict: 'A',
		link: function($scope, $elem, $attrs){
			$attrs.$observe('erosUser', function(value){
				if(value !== ""){
					$scope.user = eros.user(value);
				}
			})
		}
	};
})
.filter('toColor', function() {
	return function(input) {
		switch (input.length) {
		case 1:
		case 6:
		case 11:
			return "#FF0000"; // Red
			break;
		case 2:
		case 7:
		case 12:
			return "#4F64FF"; // Blue
			break;
		case 3:
		case 8:
		case 13:
			return "#23A136"; // Green
			break;
		case 4:
		case 9:
		case 14:
			return "#A75EAD"; // Purple
			break;
		default:
			return "#CC8E4B"; // Brown
			break;
		}
	};
})
.filter('divisionColor', function(){
	return function(division){
		switch(division) {
		case 'E':
			return "#9966cc";
			break
		case 'D':
			return "#AB3E3E";
			break
		case 'C':
			return "#ED9121";
			break
		case 'B':
			return "#0892d0";
			break
		case 'A':
			return "#009e60";
			break		
		default: // unranked
			return "#888";
			break;
		}
	}
})
