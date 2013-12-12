var kmApp = angular.module('kmApp', ['ngRoute', 'kmControllers' , 'ui.bootstrap']);

kmApp.config(['$routeProvider', function($routeProvider, $locationProvider){
    $routeProvider.
    when('/input/:id', {
        templateUrl: 'partials/input.html',
        controller: 'kmInput'
    }).
    when('/overview', {
        templateUrl:'partials/overview.html',
        controller: 'kmOverviewController'
    }).
    otherwise({
        redirectTo: '/overview'
    });
}]);
