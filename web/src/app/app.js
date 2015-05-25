'use strict';

var version = new Date().getTime();
// Declare app level module which depends on filters, and services
var erosApp = angular.module('erosApp', [
  'ngRoute',
  'ngAnimate',
  'luegg.directives',
  'erosApp.filters',
  'erosApp.services',
  'erosApp.directives',
  'erosApp.controllers',
  'erosApp.chat',
  'erosApp.mm',
  'pascalprecht.translate',
  'templates.app',
]);

erosApp.config(['$routeProvider', '$locationProvider', '$httpProvider', function($routeProvider, $locationProvider, $httpProvider) {
	$routeProvider.when('/', {templateUrl: '/static/test.html?_='+version, controller: 'ErosTestCtrl'});
	
	$routeProvider.otherwise({redirectTo: '/'});

	$locationProvider.html5Mode(true);

	$httpProvider.defaults.withCredentials = true;
}]);

erosApp.run(['$window', '$rootScope', function($window, $rootScope) {
	return;
}]);