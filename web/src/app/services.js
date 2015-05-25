'use strict';

/* Services */


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('erosApp.services', ['ngAudio'])

.value('version', '0.1')

.factory('connGrowl', ['$rootScope', function($rootScope){

	// 0 -> disconnected (initial)
	// 1 -> connected
	// 2 -> disconnected (further on, as an alert)
	var _status = 0;

	var sendMsg = function(msg, status){
		status = status || _status;
		broadcastMsg(msg, status);
	};
	var broadcastMsg = function(msg, status){
		$rootScope.$emit("connGrowl", msg, status);
	};

	return {
		sendMsg: sendMsg
	};
}])

.factory('notifier', ['ngAudio', function(ngAudio){
	var _audioNotif = ngAudio.load("/static/sounds/Transmission.wav");
	var _matchNotif = ngAudio.load("/static/sounds/notification.wav");
	var _volume = 1;
	var _baseTitle = document.title;
	var _title_messages = {};

	return {
		// Volume: optional
		message: function(){
			_audioNotif.play();
		},

		matched: function(){
			_matchNotif.play();
		},

		// proxy
		setMuting: function(value){
			_audioNotif.setMuting(value);
		},

		setVolume: function(value){
			if(typeof(value) === 'number' && value > 0 && value <= 1){
				_audioNotif.setMuting(false);
				_volume = value;
			}
		},

		title: function(pre, container, blinking){
			if(typeof pre !== 'undefined'){
				if(typeof this.blinkInterval !== 'undefined'){
					clearInterval(this.blinkInterval);
				}

				if(typeof pre !== 'undefined' && pre.length > 0){
					_title_messages[container] = pre;
				}else{
					delete _title_messages[container];
				}

				var message = '';
				for (var prop in _title_messages) {
			    	if(_title_messages.hasOwnProperty(prop)){
			        	message = message+ _title_messages[prop]+" ";
			      	}
			   	}
				
				var blink = function(){
					if(document.title == _baseTitle){
						document.title = message + _baseTitle;
					}else{
						document.title = _baseTitle;
					}
				};

				if(blinking){
					this.blinkInterval = setInterval(blink, 1000);
				}else{
					document.title = pre + " "+_baseTitle;
				}
			}
		}
	};
}])

.factory('timer', function(){
	var _timerInterval = {};

	var _add_sec = function(scope, timer){
		if (scope[timer].s == 59){
			scope[timer].s = 0;
			scope[timer].m = scope[timer].m+1;
		}else{
			scope[timer].s = scope[timer].s+1;
			if(scope[timer].s === 0 && scope[timer].m === 0){
				this.stop(timer);
			}
		}
	};

	var _sub_sec = function(scope, timer){
		if (scope[timer].s === 0){
			scope[timer].s = 59;
			scope[timer].m = scope[timer].m-1;
		}else{
			scope[timer].s = scope[timer].s-1;
		}
	};

	return {
		start: function(scope, timer){
			_timerInterval[timer] = setInterval(function(){_add_sec(scope, timer)},1000);
		},

		stop: function(timer){
			if(typeof _timerInterval[timer] !== 'undefined'){
				clearInterval(_timerInterval[timer]);
			}
		},

		restart: function(scope, timer){
			this.stop(timer);
			scope[timer] = {s: 0, m:0};

			this.start(scope, timer);
		},

		timedown: function(scope, timer, seconds){
			this.stop(timer);

			scope[timer] = {s: seconds%60, m: Math.floor(seconds/60)};
			_timerInterval[timer] = setInterval(function(){_sub_sec(scope, timer)},1000);
		}
	}
})
