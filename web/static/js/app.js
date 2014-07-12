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
	$routeProvider.when('/', {templateUrl: '/static/partials/test.html?_='+version, controller: 'ErosTestCtrl'});
	
	$routeProvider.otherwise({redirectTo: '/'});

	$locationProvider.html5Mode(true);
}]);

erosApp.run(['$window', '$rootScope', function($window, $rootScope) {
	return;
}]);
