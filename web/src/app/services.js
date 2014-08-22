'use strict';

/* Services */


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('erosApp.services', [])
  .value('version', '0.1')


.factory('connGrowl', ['$rootScope', function($rootScope){

	// 0 -> disconnected (initial)
	// 1 -> connected
	// 2 -> disconnected (further on, as an alert)
	var _status = 0

	var sendMsg = function(msg, status){
		status = status || _status;
		broadcastMsg(msg, status);
	}
	var broadcastMsg = function(msg, status){
		$rootScope.$emit("connGrowl", msg, status);
	}

	return {
		sendMsg: sendMsg
	}
}])