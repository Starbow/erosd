'use strict';

/* Controllers */

var controllers = angular.module('erosApp.controllers', []);

controllers.controller('ErosTestCtrl', ['$scope', '$http','connGrowl','$rootScope', function($scope, $http, connGrowl, $rootScope) {

	var server = window.location.host;

	$scope.message = '';
	$scope.activeUsers = 0;
	$scope.connected = false;
	$scope.latency = 0;
	$scope.rooms = {};
	$scope.privs = {};
	$scope.login = {};

	$http({
		method: 'GET',
		// url:'http://starbowmod.com/user/api/info'
		url:'http://127.0.0.1:12345/user/api/info'
	}).success(function(data, status, headers, config) {
		if (data.success) {
			// $scope.connect(data.username,  data.token)
			$scope.login.username = data.username;
			$scope.login.password = data.token;
		} else {
			$scope.message = 'Please log in to starbowmod.com to auto-fill your login details.';
			connGrowl.sendMsg('Please log in to starbowmod.com to auto-fill your login details.')
		}
    }).
    error(function(data, status, headers, config) {
    	$scope.message = 'Unable to autograb login info. ' + status;
    	connGrowl.sendMsg('Unable to autograb login info.')
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
			connGrowl.sendMsg('Connected. Authenticating...')

		},
		loggedIn: function() {
			// We're logged in. Fo real connected.
			$scope.$apply(function() {
				$scope.message = 'Authenticated! Wahoo.';
				$scope.connected = true;
			});
			connGrowl.sendMsg('Authenticated! Wahoo.',1)
		},
		loginFailed: function(eros, status) {
			// This shouldn't ever happen if we're pulling our auth direct from the API.
			$scope.$apply(function() {
				if (status === 2) {
					$scope.message = 'Already logged in from another location.'
				} else {
					$scope.message = 'Authentication failed. Stay shit.'
				}
				$scope.connected = false;
			});
			if (status === 2) {
				connGrowl.sendMsg('Already logged in from another location.', 0)
			} else {
				connGrowl.sendMsg('Authentication failed.',2)
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

		chat: {
			joined: function(eros, room) {
				$scope.$apply(function() {
					if (!(room.key in $scope.rooms)) {
						$scope.rooms[room.key] = {
							room: room,
							messages: []
						}
					}
					$scope.rooms[room.key].active = true;
					// $scope.rooms[room.key].messages.push({
					// 	sender: eros.localUser,
					// 	message: 'joined the channel.',
					// 	event: true,
					// 	date: new Date()
					// });
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
					delete $scope.rooms[room.key]

					window.event.cancelBubble = true
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
					$rootScope.$emit("chat_room","joined")
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
					$rootScope.$emit("chat_room","left")
				});
			},
			message: function(eros, room, user, message) {
				$scope.$apply(function() {
					$scope.rooms[room.key].messages.push({
						sender: user,
						message: message,
						event: false,
						date: new Date()
					});
				});
			},
			privjoined: function(eros, room){
				if (!(room.key in $scope.privs)) {
					$scope.privs[room.key] = {
						priv: room,
						messages: []
					}
				}
				$scope.privs[room.key].active = true;
				$scope.selectedRoom = $scope.privs[room.key]
			},
			privleave: function(eros, priv){
				// $scope.$apply(function() {
					delete $scope.privs[priv.key]

				// window.event.cancelBubble = true

				if($scope.selectedRoom.priv == priv){
					$scope.selectedRoom = $scope.rooms[Object.keys($scope.rooms)[0]];
				}

				
				// })
			},
			privmessage: function(eros, room, user, message) {
				if(!$scope.$$phase){
						$scope.$apply(function() {
						$scope.privs[room.key].messages.push({
							sender: user,
							message: message,
							event: false,
							date: new Date()
						});
					});
				}else{
					$scope.privs[room.key].messages.push({
						sender: user,
						message: message,
						event: false,
						date: new Date()
					});
				}
				
			},
		},

	});

	// Horrible uglyness. Remove in production.
	window.eros = eros;
	$scope.eros = eros;

	$scope.$on('$destroy', function(){
		// Disconnect when changing controller.
		// We absolutely don't want to do this in the real world.
		eros.disconnect();
	});


	$scope.connect = function(username, password) {
		if (!username) {
			username = "ngtest";
		}

		if (!password) {
			password = "ngtest";
		}
		eros.connect(username, password);
	}

}]);

