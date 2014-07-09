'use strict';

var version = new Date().getTime();
// Declare app level module which depends on filters, and services
var erosApp = angular.module('erosApp', [
  'ngRoute',
  'erosApp.filters',
  'erosApp.services',
  'erosApp.directives',
  'erosApp.controllers'
]);

erosApp.config(['$routeProvider', '$locationProvider', function($routeProvider, $locationProvider) {
	$routeProvider.when('/view1', {templateUrl: '/static/partials/partial1.html?_='+version, controller: 'MyCtrl1'});
	$routeProvider.when('/view2', {templateUrl: '/static/partials/partial2.html?_='+version, controller: 'MyCtrl2'});
	$routeProvider.otherwise({redirectTo: '/view1'});

	$locationProvider.html5Mode(true).hashPrefix('!');
}]);

erosApp.run(['$window', '$rootScope', function($window, $rootScope) {
	var ws = new WebSocket('ws://'+window.location.host+'/ws');
	ws.binaryType = "arraybuffer";
	var builder = dcodeIO.ProtoBuf.loadProtoFile("/eros.proto"),
	protobufs = builder.build('protobufs');



 	// absolutely do not do this in production. Make an Eros service.

	var txBase = 0;

	function sendMessage(command, message) {
		var tx = ++txBase;
		var data = message.toArrayBuffer();
		
		var out = dcodeIO.ByteBuffer.concat([command + ' ' + tx + ' ' + data.byteLength + '\n', data]).toBuffer();	
		
		ws.send(out);
	}

	function processServerMessage(command, buffer) {
		if (command == "SSU") {
			var stats = protobufs.ServerStats.decode(buffer);
			console.log('There are ' + stats.active_users + ' users online.');
		} else if (command == "CHJ") {
			var stats = protobufs.ChatRoomUser.decode(buffer);
			console.log('Joined channel ' + stats.room.name);
		}
	}

	function processMessage(command, tx, buffer) {
		if (command == "HSH") {
			var command = protobufs.HandshakeResponse.decode(buffer);
			console.log('Connected. We are user ID ' + command.id);
		}
	}

	ws.onopen = function() {

		var handshake = new protobufs.Handshake("username", "authkey");
		sendMessage("HSH", handshake);
	};
	ws.onmessage = function(e) {
		if (typeof(e.data === 'ArrayBuffer')) {
			var buffer = dcodeIO.ByteBuffer.wrap(e.data);

			var header = '';
			while (true) {
				var x = buffer.readString(1); 

				if (x == '' || x == '\n') {
					break;
				}

				header += x;
			}

			if (header != '') {
				header = header.split(' ');
				console.log(header[0]);
				if (header.length == 2) {
					processServerMessage(header[0], buffer, Number(header[1]))
				} else {
					processMessage(header[0], Number(header[1]), buffer, Number(header[2]))
				}
			}

		}
	};
	ws.onclose = function() {
		alert("closed");
	};
}]);
