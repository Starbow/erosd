(function(global) {
    "use strict";


    var Request = function(command, payload, payloadHandlers, errorHandlers) {
    	this.status = 0;
    	this.command = command;
    	this.complete = false;
    	this.result = undefined;

    	if (typeof(payload) === "undefined") {
    		this.payload = "";
    	} else {
    		this.payload = payload;
    	}

    	if (typeof(errorHandlers) === "undefined") {
    		errorHandlers = function(code, payload) {
    			this.status = code;
    			this.complete = true;
    		}
    	}

    	this.processPayload = function(command, payload) {
    		var handlers = undefined;
    		if (isNaN(command)) {
    			handlers = payloadHandlers;
    		} else {
    			handlers = errorHandlers;
    		}

    		if (typeof(handlers) === 'function') {
    			return handlers(command, payload);
    		} else if ((typeof(handlers) === 'object') && (command in handlers)) {
    			return handlers[command](command, payload);
    		} else {
    			console.log('No handler found for command ' + command);
    			return false;
    		}
    	};
    };

    if (!global["starbow"]) {
        global["starbow"] = {};
    }

    if (!global["starbow"]["ErosRequests"]) {
        global["starbow"]["ErosRequests"] = {};
    }

    global["starbow"]["ErosRequests"]["Request"] = Request;
})(this);