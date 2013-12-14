var kmApp = angular.module('kmApp', ['ngRoute', 'kmControllers' , 'ui.bootstrap']);

kmApp.config(['$routeProvider', function($routeProvider, $locationProvider){
    var d = new Date();
    $routeProvider.
    when('/input/:id', {
        templateUrl: 'partials/input.html',
        controller: 'kmInput'
    }).
    when('/overview/:year/:month', {
        templateUrl:'partials/overview.html',
        controller: 'kmOverviewController'
    }).
    when('/overview', {
        redirectTo: '/overview/' + d.getFullYear() + '/' + (d.getMonth() +1)
    }).
    otherwise({
        redirectTo: '/overview/'
    });
}]);
