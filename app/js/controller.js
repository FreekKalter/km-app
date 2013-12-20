/*jslint browser:true*/
/*global angular, kmApp*/
'use strict';

var kmControllers = angular.module('kmControllers', []);

kmControllers.controller('kmInput', function($scope,$routeParams, $location, $http){
    $scope.fields = [ 'Begin', 'Eerste', 'Laatste', 'Terug' ];

    $scope.getState = function(){
        $http.get('state/'+ $routeParams.id).success(function(data){
            $scope.form = data;
            var toFocus;
            if (data.Begin.Editable === true){
                toFocus = 'Begin';
            }else if(data.Eerste.Editable === true){
                toFocus = 'Eerste';
            }else if(data.Laatste.Editable === true){
                toFocus = 'Laatste';
            }else if(data.Terug.Editable === true){
                toFocus = 'Terug';
            }
            if(toFocus !== undefined){
                setTimeout(function(){ setFocus(document.getElementById(toFocus)); }, 100);
            }
        });
    };

    $scope.save = function(name, fieldValue){
        $http.post('/save', {name: name, value: fieldValue}).success(function(data){
            if(name === 'Terug'){
                $location.path('/overview');
            }else{
                $scope.form[name].Editable=false;
                $scope.getState();
            }
        });
    };

    $scope.edit = function(name){
        $scope.form[name].Editable=true;
    };

    $scope.valid = function(name){
        return $scope.kmform['{{field}}'].$error.integer;
    };

    function setFocus(el){
        el.focus();
        var strl = el.value.length;
        el.setSelectionRange(strl,strl);
    }
    $scope.getState();
});

kmControllers.controller('kmOverviewController', function($scope,$routeParams, $location, $http){

    // Get tabs accesable via bookmarkeble url: bit of a hacky solution!!
    //
    // Load data has 2 functions
    //
    // 1) on initial page load it get gets called 1 time, to fetch data from backend.
    // 2) When switching tabs loadData gets called 2 times:
    //      - first time desired tab state does noet equal whats is currently in url
    //         -> so change url to match desired state
    //      - second time desired tab is also in url -> now fetch data from backend and render it
    //          (like on an initial page load)
    //
    $scope.loadData = function(category){
        var path = [ 'overview', category, $routeParams.year, $routeParams.month].join('/');
        if(category === $routeParams.category){
            if(category === 'kilometers'){
                $http.get(path).success(function(data){
                    $scope.kilometers = data;
                });
            }else if(category === 'tijden'){
                $http.get(path).success(function(data){
                    $scope.times = data;
                });
            }
        }else{
            $location.path(path);
        }
    };

    // Activate tab based on url
    if($routeParams.category === 'kilometers'){
        $scope.kiloActive = true;
    }
    if($routeParams.category === 'tijden'){
        $scope.timesActive = true;
    }

    $scope.deleteRow = function(index){
        $http.get('delete/' + $scope.kilometers[index].Id ).success(function(data){
            $scope.kilometers.splice(index, 1); // delete ellemnt from array (delete undefines element)
        });
    };

    $scope.editRow = function(index){
        $location.path('/input/' + $scope.kilometers[index].Id);
    };

    $scope.go = function(path){
        if( path === 'next' ){
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
        $scope.next = { date: n, link: ['overview' , $routeParams.category, n.getFullYear(), (n.getMonth()+1)].join('/') };
    }

    var p = new Date();
    p.setMonth($routeParams.month -1);
    p.setFullYear($routeParams.year);
    p.setMonth(p.getMonth()-1);
    $scope.prev = { date: p, link: ['overview' , $routeParams.category, p.getFullYear(), (p.getMonth()+1)].join('/') };
});

var INTEGER_REGEXP = /^\-?\d*$/;
kmApp.directive('integer', function() {
    return {
        require: 'ngModel',
        link: function(scope, elm, attrs, ctrl) {
            ctrl.$parsers.unshift(function(viewValue) {
                if (INTEGER_REGEXP.test(viewValue)) {
                    // it is valid
                    if(attrs.id === 'Begin'){
                        ctrl.$setValidity('integer', true);
                        return viewValue;
                    }
                    if(attrs.id === 'Eerste'){
                        if(viewValue >= scope.form.Begin.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id === 'Laatste'){
                        if(viewValue >= scope.form.Eerste.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id === 'Terug'){
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
