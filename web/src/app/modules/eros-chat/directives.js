'use strict';

/* Chat directives */

angular.module('erosApp.chat', ['ngAudio'])
.directive('erosChat', ['$rootScope','notifier', function($rootScope, notifier){
	return {
		restrict: 'A',
		templateUrl: 'modules/eros-chat/chat.tpl.html',
		controller: ['$scope',  function($scope){

			$scope.selectRoom = function(room){
				if(typeof room == "object"){
					$scope.$parent.selectedRoom = room;
					$rootScope.$emit("chat_room","selectedRoom");

					$('#chat-input > input').focus()
				}
			};

			$scope.sendChatMessage = function(target, message) {
				// if(message[0] === "@"){
				// 	target = message.split(" ",1)[0].split("@")[1];
				// 	eros.chat.sendToPriv(target, message.split(" ").slice(1).join(" "));
				// 	$scope.selectRoom($scope.privs[target.toLowerCase()]);
				// }else 
				if(typeof ($scope.selectedRoom.priv) == 'object'){
					eros.chat.sendToPriv($scope.selectedRoom.priv.name, message);
				}else{
					eros.chat.sendToRoom(target, message);
				}
				// $scope.chatMessage = "real"
			};

			// Deprecated
			$scope.addUserMsg = function(user){
				var newmessage;
				if(typeof $scope.chatMessage == "undefined" || $scope.chatMessage.length === 0){
					newmessage = "@" + user + " ";
				}else{
					newmessage = $scope.chatMessage + " @" + user + " ";
				}
				$scope.$parent.chatMessage = newmessage;

				document.getElementById("chat-input").childNodes[0].focus();
			};

			$scope.openPrivUser = function(username){
				if(typeof username === 'string' && username != eros.localUser.username){
					if(typeof $scope.privs[username.toLowerCase()] === 'undefined'){
						eros.chat.joinPriv(username);
					}

					$scope.selectRoom($scope.privs[username.toLowerCase()]);
					$('#chat-input > input').focus();
				}
			};

			$scope.updateChatInput = function(message){
				// Replace user names
				$scope.chatMessage;
			};

			$scope.seeChat = function(){
				if(!document.hidden){
					$rootScope.$emit("favicon_alert", false);
				}
			}
		}]
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
