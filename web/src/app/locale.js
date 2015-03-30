'use strict';

var module = angular.module('erosApp');

module.config(['$translateProvider', function ($translateProvider) {

    $translateProvider.translations('en', starbow.locale["en"]);
	$translateProvider.translations('kr', starbow.locale["kr"]);

    $translateProvider.preferredLanguage('en');
}]);

module.directive('changeLanguage', function($translate){
	return {
		restrict: 'AC',
		link: function($scope, $elem, $attrs, $controller){
			$elem.on('click', function(){
				var lang = $attrs['lang']
				if(lang.length > 0){
					$translate.use(lang);
				}
			})
		}
	}
});