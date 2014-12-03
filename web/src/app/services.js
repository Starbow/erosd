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

.factory('notifier', ['ngAudio', function(ngAudio){
	var _audioNotif = ngAudio.load("/static/sounds/Transmission.wav");
	var _matchNotif = ngAudio.load("/static/sounds/notification.wav");
	var _volume = 1;
	var _baseTitle = document.title
	var _title_messages = {}

	return {
		// Volume: optional
		message: function(){
			_audioNotif.play()
		},

		matched: function(){
			_matchNotif.play()
		},

		// proxy
		setMuting: function(value){
			_audioNotif.setMuting(value)
		},

		setVolume: function(value){
			if(typeof(value) === 'number' && value > 0 && value <= 1){
				_audioNotif.setMuting(false)
				_volume = value
			}
		},

		title: function(pre, container, blinking){
			if(typeof pre !== 'undefined'){
				if(typeof this.blinkInterval !== 'undefined'){
					clearInterval(this.blinkInterval)
				}

				if(typeof pre !== 'undefined' && pre.length > 0){
					_title_messages[container] = pre
				}else{
					delete _title_messages[container]
				}

				var message = '';
				for (var prop in _title_messages) {
			    	if(_title_messages.hasOwnProperty(prop)){
			        	var message = message+ _title_messages[prop]+" ";
			      	}
			   	}
				
				var blink = function(){
					if(document.title == _baseTitle){
						document.title = message + _baseTitle;
					}else{
						document.title = _baseTitle;
					}
				}

				if(blinking){
					this.blinkInterval = setInterval(blink, 1000)
				}else{
					document.title = pre + " "+_baseTitle;
				}
			}
		}
		
	};
}])

.factory('timer', function(){
	var _timerInterval = {}

	var _add_sec = function(scope, timer){
		if (scope[timer].s == 59){
			scope[timer].s = 0
			scope[timer].m = scope[timer].m+1
		}else{
			scope[timer].s = scope[timer].s+1
			if(scope[timer].s==0 && scope[timer].m == 0){
				this.stop(timer)
			}
		}
	}

	var _sub_sec = function(scope, timer){
		if (scope[timer].s == 0){
			scope[timer].s = 59
			scope[timer].m = scope[timer].m-1
		}else{
			scope[timer].s = scope[timer].s-1
		}
	}

	return {
		start: function(scope, timer){
			_timerInterval[timer] = setInterval(function(){_add_sec(scope, timer)},1000);
		},

		stop: function(timer){
			if(typeof _timerInterval[timer] !== 'undefined'){
				clearInterval(_timerInterval[timer])
			}
		},

		restart: function(scope, timer){
			this.stop(timer)
			scope[timer] = {s: 0, m:0}

			this.start(scope, timer)
		},

		timedown: function(scope, timer, seconds){
			this.stop(timer)

			scope[timer] = {s: seconds%60, m: Math.floor(seconds/60)}
			_timerInterval[timer] = setInterval(function(){_sub_sec(scope, timer)},1000);
		}
	}
})

.factory('browserid', ['$http', function($http){

	// Deferred for post-watch-callback actions.
    var requestDeferred = null;
    var logoutDeferred = null;

    // Track if we've avoided the first auto-login.
    var avoidedAutoLogin = false;

    // Public API
    var eros_browserid = {
        /**
         * Retrieve an assertion and use it to log the user into your site.
         * @param {object} requestArgs Options to pass to navigator.id.request.
         * @return {jQuery.Deferred} Deferred that resolves once the user has
         *                           been logged in.
         */
        login: function login(requestArgs) {
            return eros_browserid.getAssertion(requestArgs).then(function(assertion) {
                return eros_browserid.verifyAssertion(assertion);
            });
        },
                /**
         * Log the user out of your site.
         * @return {jQuery.Deferred} Deferred that resolves once the user has
         *                           been logged out.
         */
        logout: function logout() {
            return eros_browserid.getInfo().then(function(info) {
                logoutDeferred = $.Deferred();
                navigator.id.logout();

                return logoutDeferred.then(function() {
                    return $.ajax(info.logoutUrl, {
                        type: 'POST',
                        headers: {'X-CSRFToken': info.csrfToken},
                    });
                });
            });
        },

        /**
         * Retrieve an assertion via BrowserID.
         * @param {object} requestArgs Options to pass to navigator.id.request.
         * @return {jQuery.Deferred} Deferred that resolves with the assertion
         *                           once it is retrieved.
         */
        getAssertion: function getAssertion(requestArgs) {
            return eros_browserid.getInfo().then(function(info) {
                requestArgs = $.extend({}, info.requestArgs, requestArgs);

                requestDeferred = $.Deferred();
                navigator.id.request(requestArgs);
                return requestDeferred;
            });
        },

        /**
         * Verify that the given assertion is valid, and log the user in.
         * @param {string} Assertion to verify.
         * @return {jQuery.Deferred} Deferred that resolves with the login view
         *                           response once login is complete.
         */
        verifyAssertion: function verifyAssertion(assertion) {
            return eros_browserid.getInfo().then(function(info) {
                return $.ajax(info.loginUrl, {
                    type: 'POST',
                    data: {assertion: assertion},
                    headers: {'X-CSRFToken': info.csrfToken},
                });
            });
        },

        // Cache for the AJAX request created by eros_browserid.getInfo().
        // Stored on the public API so tests can reset it.
        _infoXHR: null,

        /**
         * Fetch the info for the Persona popup and login requests.
         * @return {jqXHR} jQuery XmlHttpResponse that returns the data.
         */
        getInfo: function getInfo() {
            if (eros_browserid._infoXHR === null) {
                // eros_browserid._infoXHR = $http({
                // 	method: 'GET',
                // 	url: 'http://127.0.0.1:8000/browserid/info/'
                // })
				eros_browserid._infoXHR = $http.jsonp('http://starbowmod.com/browserid/info/')
            }

            return eros_browserid._infoXHR;
        },

        /**
         * Check for the querystring parameter used to signal a failed login.
         * @param {Location} Location object containing URL to check. Defaults
         *                   to window.location, used for testing.
         * @return {Boolean} True if the parameter was found and login failed,
         *                   False otherwise.
         */
        didLoginFail: function didLoginFail(location) {
            location = location || window_location;
            return location.search.indexOf('bid_login_failed=1') !== -1;
        },

        /**
         * Register callbacks with navigator.id.watch that make the API work.
         * This must be called before calling any other API methods.
         * @return {jqXHR} Deferred that resolves after the handlers have been
         *                 have been registered.
         */
        registerWatchHandlers: function registerWatchHandlers() {
            return eros_browserid.getInfo().success(function(info) {
                navigator.id.watch({
                    loggedInUser: info.userEmail,
                    onlogin: function(assertion) {
                        // Avoid auto-login on failure.
                        if (!avoidedAutoLogin && eros_browserid.didLoginFail()) {
                            navigator.id.logout();
                            avoidedAutoLogin = true;
                            return;
                        }

                        if (requestDeferred) {
                            requestDeferred.resolve(assertion);
                        }
                    },
                    onlogout: function() {
                        if (logoutDeferred) {
                            logoutDeferred.resolve();
                        }
                    }
                });
            }).error(function(info){
            	console.error("Error getting login info.")
            });
        }
    };

    return eros_browserid;
}])