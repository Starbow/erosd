'use strict';

var module = angular.module('erosApp');

module.config(['$translateProvider', function ($translateProvider) {
    $translateProvider.translations('en', starbow.locale["en"]);
	$translateProvider.translations('kr', starbow.locale["kr"]);
	$translateProvider.translations('es', starbow.locale["es"]);
	$translateProvider.translations('ru', starbow.locale["ru"]);

    $translateProvider.preferredLanguage('en');
    // console.log($translateProvider.preferredLanguage())
    
}]);

module.controller('LocaleController', ['$scope', '$locale', '$translate', function($scope, $locale, $translate){
	// alert($locale.id + ", "+$locale.id.split('-')[0])

	var lang;
	switch($locale.id.split('-')[0]){
		case 'kr': 
			lang = 'kr';
			break;
		case 'es':
			lang = 'es'
			break;
		default:
			lang = 'en'
	}

	$scope.setLanguage = lang;
	$translate.use(lang);
}]);

module.directive('changeLanguage', function($translate){
	return {
		restrict: 'AC',
		// controller: 'LocaleController',
		link: function($scope, $elem, $attrs, $controller){
			$elem.on('click', function(){
				var lang = $attrs['lang']
				$scope.setLanguage = lang
				if(lang.length > 0){
					$translate.use(lang);
				}
			})
		}
	}
});