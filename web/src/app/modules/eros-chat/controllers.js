angular.module('erosApp.chat', ['ngAudio'])

.controller('ChatCtrl', ['$scope', '$rootScope',  function($scope, $rootScope){
	$scope.selectRoom = function(room){
		if(typeof room == "object"){
			$scope.$parent.selectedRoom = room;
			$rootScope.$emit("chat_room","selectedRoom");

			$('#chat-input > input').focus()
		}
	};

	$scope.sendChatMessage = function(target, message) {
		if(typeof ($scope.selectedRoom.priv) == 'object'){
			eros.chat.sendToPriv($scope.selectedRoom.priv.name, message);
		}else{
			eros.chat.sendToRoom(target, message);
		}
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
}])