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
	return;
}]);
