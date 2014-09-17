'use strict';

angular.module('erosApp.mm', [])

.controller('MmCtrl', ['$scope', function($scope){
	$scope.localUser = eros.localUser;
}])

.directive('erosMm',function(){
	return {
		templateUrl: 'modules/eros-mm/mm.tpl.html',
		replace: true,
		scope: true,
		controller: 'MmCtrl'
	}
});
