'use strict';

/* Controllers */

var controllers = angular.module('erosApp.controllers', ['ngAudio']);

controllers.controller('ErosTestCtrl', ['$scope', '$http','connGrowl','$rootScope','ngAudio','notifier','timer', function($scope, $http, connGrowl, $rootScope, ngAudio, notifier, timer) {

	var server = window.location.host;

	$scope.message = '';
	$scope.activeUsers = 0;
	$scope.connected = false;
	$scope.latency = 0;
	$scope.rooms = {};
	$scope.privs = {};
	$scope.login = {};
	$scope.notifier =  notifier;

	$scope.matchmaking = {};

	$http({
		method: 'GET',
		url:'//starbowmod.com/user/api/info'
		// url:'http://127.0.0.1:12345/user/api/info'
	}).success(function(data, status, headers, config) {
		if (data.success) {
			// $scope.login.username = data.username;
			// $scope.login.password = data.token;
			$scope.connect(data.username,  data.token);
		} else {

			$scope.showLogin = true;
			$scope.message = 'Please log in to starbowmod.com to auto-fill your login details.';
			connGrowl.sendMsg('Please log in to starbowmod.com to auto-fill your login details.');

		}
    }).
    error(function(data, status, headers, config) {
    	$scope.showLogin = true;
    	$scope.message = 'Unable to autograb login info. ' + status;
    	connGrowl.sendMsg('Unable to autograb login info.');
    });

	var eros = new starbow.Eros({
		// The first parameter of every callback is the Eros object that initiated it.
		// We don't care, so we're not providing parameters except when we're interested
		// in other stuff.

		connected: function() {
			// This is pre-authentication connected. I guess it's pointless?
			$scope.$apply(function() {
				$scope.message = 'Connected. Authenticating...';
			});
			connGrowl.sendMsg('Connected. Authenticating...');

		},
		loggedIn: function() {
			// We're logged in. Fo real connected.
			$scope.$apply(function() {
				$scope.message = 'Authenticated! Wahoo.';
				$scope.connected = true;
			});
			connGrowl.sendMsg('Authenticated! Wahoo.',1);
		},
		loginFailed: function(eros, status) {
			// This shouldn't ever happen if we're pulling our auth direct from the API.
			$scope.$apply(function() {
				if (status === 2) {
					$scope.message = 'Already logged in from another location.';
				} else {
					$scope.message = 'Authentication failed. Stay shit.';
				}
				$scope.connected = false;
			});
			if (status === 2) {
				connGrowl.sendMsg('Already logged in from another location.', 0);
			} else {
				connGrowl.sendMsg('Authentication failed.',2);
			}
		},
		disconnected: function() {
			$scope.$apply(function() {
				$scope.connected = false;
			});
		},

		statsUpdate: function() {
			$scope.$apply(function() {
				$scope.stats = eros.stats;
			});
		},

		regionUpdate: function(eros, region) {
			// region is the name of the region (EU, NA, etc)
			$scope.$apply(function() {
				$scope.regions = eros.regions;
			});
		},

		latencyUpdate: function() {
			$scope.$apply(function() {
				$scope.latency = eros.latency;
			});
		},

		// Move to registerController in chat controller
		chat: { 
			joined: function(eros, room) {
				$scope.$apply(function() {
					if (!(room.key in $scope.rooms)) {
						$scope.rooms[room.key] = {
							room: room,
							messages: [],
							new_messages: [],
							visit: function(){
								this.messages = this.messages.concat(this.new_messages);
								this.new_messages = [];
							}
						};
					}
					$scope.rooms[room.key].active = true;

					if(typeof $scope.selectedRoom === 'undefined'){
						$scope.defaultRoom = $scope.rooms[room.key];
						$scope.selectedRoom = $scope.rooms[room.key];
						$rootScope.$emit("chat_room","selectedRoom");
					}
				});
			},
			left: function(eros, room) {
				$scope.$apply(function() {
					// $scope.rooms[room.key].active = false;
					// $scope.rooms[room.key].messages.push({
					// 	sender: eros.localUser,
					// 	message: 'left the channel.',
					// 	event: true,
					// 	date: new Date()
					// });
					if($scope.selectedRoom.room.key == room.key){
						delete $scope.rooms[room.key];
						if(typeof $scope.defaultRoom === 'undefined' || $scope.defaultRoom.room.key === room.key){
							$scope.defaultRoom = $scope.rooms[Object.keys($scope.rooms)[0]];
						}
						$scope.selectedRoom = $scope.defaultRoom;
						$rootScope.$emit("chat_room","selectedRoom");
					}else{
						delete $scope.rooms[room.key];
					}					

					window.event.cancelBubble = true;
				});
			},
			userJoined: function(eros, room, user) {
				$scope.$apply(function() {
					$scope.rooms[room.key].active = true;
					$scope.rooms[room.key].messages.push({
						sender: user,
						message: 'joined the channel.',
						event: true,
						date: new Date()
					});
					$rootScope.$emit("chat_room","joined");
				});
			},
			userLeft: function(eros, room, user) {
				$scope.$apply(function() {
					$scope.rooms[room.key].active = false;
					$scope.rooms[room.key].messages.push({
						sender: user,
						message: 'left the channel.',
						event: true,
						date: new Date()
					});
					$rootScope.$emit("chat_room","left");
				});
			},
			message: function(eros, room, user, content) {
				$scope.$apply(function() {

					var message = {
						sender: user,
						message: content,
						event: false,
						date: new Date()
					};

					if($scope.selectedRoom.room && room.key == $scope.selectedRoom.room.key){
						$scope.rooms[room.key].messages.push(message);
					}else{
						$scope.rooms[room.key].new_messages.push(message);
					}
					
				});
			},
			privjoined: function(eros, room){
				if (!(room.key in $scope.privs)) {
					$scope.privs[room.key] = {
						priv: room,
						messages: [],
						new_messages: [],
						visit: function(){
							$scope.newMessages = $scope.newMessages - this.new_messages;
							this.messages = this.messages.concat(this.new_messages);
							this.new_messages = [];
						}
					};
					// notifier.message()
				}
				$scope.privs[room.key].active = true;
				// $scope.selectedRoom = $scope.privs[room.key]
			},
			privleave: function(eros, priv){
				// $scope.$apply(function() {
				delete $scope.privs[priv.key];

				// window.event.cancelBubble = true

				if($scope.selectedRoom.priv == priv){
					$scope.selectedRoom = $scope.rooms[Object.keys($scope.rooms)[0]];
				}
				if($scope.selectedRoom.priv.key == priv.key){
					delete $scope.privs[priv.key];
					if(typeof $scope.defaultRoom === 'undefined' || $scope.defaultRoom.priv.key === priv.key){
						$scope.defaultRoom = $scope.rooms[Object.keys($scope.rooms)[0]];
					}
					$scope.selectedRoom = $scope.defaultRoom;
					$rootScope.$emit("chat_room","selectedRoom");
				}else{
					delete $scope.rooms[room.key];
				}	

				
				// })
			},
			privmessage: function(eros, room, user, message) {
				var message = {
					sender: user,
					message: message,
					event: false,
					date: new Date()
				};

				if($scope.selectedRoom.priv && room.key == $scope.selectedRoom.priv.key){
					var message_array = $scope.privs[room.key].messages;
					
				}else {
					var message_array = $scope.privs[room.key].new_messages;
					$scope.newMessages++;
				}

				if(!$scope.$$phase){
					$scope.$apply(function() {
						message_array.push(message);
					});
				}else{
					message_array.push(message);
				}

				if(document.hidden){
					notifier.message();
				}
			}
		}
	});

	// Horrible uglyness. Remove in production.
	window.eros = eros;
	// $scope.eros = eros;
	$rootScope.eros = eros;

	$scope.$on('$destroy', function(){
		// Disconnect when changing controller.
		// We absolutely don't want to do this in the real world.
		eros.disconnect();
	});

	$scope.newMessages = 0;

	$scope.$watch('newMessages', function(value){
		if(typeof value !== "undefined" && value > 0){
			notifier.title('['+value+']', 'chat', true);
		}else{
			notifier.title('','chat', false);
		}
	});


	$scope.connect = function(username, password) {
		eros.connect(username, password);
	};

}]);

