'use strict';
var kmControllers = angular.module('kmControllers', []);

kmControllers.controller('kmInput', function($scope,$routeParams, $location, $http){
    $scope.fields = [ 'Begin', 'Eerste', 'Laatste', 'Terug' ];

    $scope.getState = function(){
        $http.get('state/'+ $routeParams.id).success(function(data){
            $scope.form = data;
            var toFocus;
            if (data.Begin.Editable == true){
                toFocus = 'Begin';
            }else if(data.Eerste.Editable == true){
                toFocus = 'Eerste';
            }else if(data.Laatste.Editable == true){
                toFocus = 'Laatste';
            }else if(data.Terug.Editable == true){
                toFocus = 'Terug';
            }
            if(toFocus !== undefined){
                setTimeout(function(){ setFocus(document.getElementById(toFocus)) }, 100);
            }
        });
    };

    $scope.save = function(name, fieldValue){
        $http.post('/save', {name: name, value: fieldValue}).success(function(data){
            if(name == 'Terug'){
                $location.path('/overview');
            }else{
                eval('$scope.form.'+name+'.Editable=false')
                $scope.getState();
            }
        });
    };

    $scope.edit = function(name){
        eval('$scope.form.'+name+'.Editable=true')
    };

    $scope.valid = function(name){
        console.log($scope);
        return $scope.kmform['{{field}}'].$error.integer;
    }

    function setFocus(el){
        el.focus();
        var strl = el.value.length;
        el.setSelectionRange(strl,strl);
    }
    $scope.getState();
});

kmControllers.controller('kmOverviewController', function($scope,$routeParams, $location, $http){
    var suffix =  $routeParams.year + '/' + $routeParams.month;
    //TODO: combine in 1 request
    $http.get('overview/kilometers/' + suffix).success(function(data){
        $scope.days = data;
    });

    $scope.loadMe = function(){
        $http.get('overview/tijden/' + suffix).success(function(data){
            $scope.times = data;
        });
    };

    $scope.delete = function(index){
        $http.get('delete/' + $scope.days[index].Id ).success(function(data){
            $scope.days.splice(index, 1) // delete ellemnt from array (delete undefines element)
        });
    };
    $scope.edit = function(index){
        $location.path('/input/' + $scope.days[index].Id);
    };
    $scope.go = function(path){
        if( path == 'next' ){
            $location.path($scope.next.link);
        } else{
            $location.path($scope.prev.link);
        }
    };

    // don't set next when next is in the future
    var n = new Date();
    if (!($routeParams.month == (n.getMonth()+1) && $routeParams.year == n.getFullYear())) {
        n.setMonth($routeParams.month -1 );
        n.setFullYear($routeParams.year);
        n.setMonth(n.getMonth()+1);
        $scope.next = { date: n, link: 'overview/' + n.getFullYear() + '/' + (n.getMonth()+1) };
    }

    var p = new Date();
    p.setMonth($routeParams.month -1);
    p.setFullYear($routeParams.year);
    p.setMonth(p.getMonth()-1);
    $scope.prev = { date: p, link: 'overview/' + p.getFullYear() + '/' + (p.getMonth()+1) };
});

var INTEGER_REGEXP = /^\-?\d*$/;
kmApp.directive('integer', function() {
    return {
        require: 'ngModel',
        link: function(scope, elm, attrs, ctrl) {
            ctrl.$parsers.unshift(function(viewValue) {
                if (INTEGER_REGEXP.test(viewValue)) {
                    // it is valid
                    if(attrs.id == 'Begin'){
                        ctrl.$setValidity('integer', true);
                        return viewValue;
                    }
                    if(attrs.id == 'Eerste'){
                        if(viewValue >= scope.form.Begin.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id == 'Laatste'){
                        if(viewValue >= scope.form.Eerste.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id == 'Terug'){
                        if(viewValue >= scope.form.Laatste.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }

                }
                // it is invalid, return undefined (no model update)
                ctrl.$setValidity('integer', false);
                return undefined;
            });
        }
    };
});

kmApp.directive('ngEnter', function () {
    return function (scope, element, attrs) {
        element.bind("keydown keypress", function (event) {
            if(event.which === 13) {
                scope.$apply(function (){
                    scope.$eval(attrs.ngEnter);
                });

                event.preventDefault();
            }
        });
    };
});
