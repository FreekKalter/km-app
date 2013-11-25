'use strict';
var km = angular.module('km', []);

km.controller('KmCtrl', function($scope, $http){
    $http.get('state').success(function(data){
        $scope.date = data.date;
        $scope.begin = data.begin;
        $scope.arnhem = data.arnhem;
        $scope.laatste = data.laatste;
        $scope.terugkomst = data.terugkomst;
    });

    $scope.beginEnabled = true;
    $scope.arnhemEnabled = true;
    $scope.laatsteEnabled = true;
    $scope.terugkomstEnabled = true;
    $scope.save = function(name, fieldValue){
        $http.post('/save', {name: name, value: fieldValue});
    };

});

