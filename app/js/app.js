var kmApp = angular.module('kmApp', ['ngRoute', 'kmControllers' ]);

kmApp.config(['$routeProvider', function($routeProvider){
    $routeProvider.
    when('/', {
        templateUrl: 'partials/input.html',
        controller: 'kmInput'
    }).
    when('/overview', {
        templateUrl:'partials/overview.html',
        controller: 'kmOverviewController'
    }).
    otherwise({
        redirectTo: '/'
    });
}]);
