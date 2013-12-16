var kmApp = angular.module('kmApp', ['ngRoute', 'kmControllers' , 'ui.bootstrap']);

kmApp.config(['$routeProvider', function($routeProvider, $locationProvider){
    var d = new Date();
    $routeProvider.
    when('/input/:id', {
        templateUrl: 'partials/input.html',
        controller: 'kmInput'
    }).
    when('/overview/:category/:year/:month', {
        templateUrl:'partials/overview.html',
        controller: 'kmOverviewController'
    }).
    when('/overview', {
        redirectTo: '/overview/kilometers/' + d.getFullYear() + '/' + (d.getMonth() +1)
    }).
    otherwise({
        redirectTo: '/overview/'
    });
}]);
