/*jslint browser: true*/
/*jshint globalstrict:true */
/*global angular, kmApp, console*/
'use strict';

var kmControllers = angular.module('kmControllers', []);

kmControllers.controller('kmInput', function($scope,$routeParams, $location, $http){
    $scope.fields = [ 'Begin', 'Eerste', 'Laatste', 'Terug' ];
    $('#datumprik').datepicker({
        format: "dd-mm-yyyy",
        weekStart: 1,
        calendarWeeks: true,
        autoclose: true,
        todayHighlight: true
    });

    function formatDate(d){
        d = d.split("");
        return d.slice(0,2).join("") + "-" + d.slice(2,4).join("") + "-" + d.slice(4,8).join("");
    }
    function getTimeStamp(t){
        var d = new Date();
        return padStr(d.getHours()) + ":" + padStr(d.getMinutes())
    }
    function padStr(i) {
       return (i < 10) ? "0" + i : i;
    }

    $scope.getState = function(date){
        $http.get('state/'+ date).success(function(data){
            $scope.form  = data;
            console.log($scope.form);
            //deepcopy, otherwise form and original will point to the same thing
            // if performance becomes an issue here, make a custom copy function
            $scope.original = JSON.parse(JSON.stringify(data));
            $scope.form.Date = formatDate(date);

            var fields = $scope.form.Fields;
            for(var i=0; i<fields.length; i++){
                if(fields[i].Km != 0) {
                    fields[i].Saved = true;
                }
            }
            if(fields[0].Km == 0){
                fields[0].Km = data.LastDayKm;
            }else{
                for(var i=1; i<fields.length; i++){
                    if(fields[i].Km == 0){
                        fields[i].Km = Math.floor(fields[i-1].Km/1000);
                        break;
                    }
                }
            }
            for(var i=0; i<fields.length; i++ ){
                if (fields[i].Time == ""){
                    fields[i].Time = getTimeStamp();
                    break;
                }
            }
            var toFocus;
            if(toFocus !== undefined){
                setTimeout(function(){ setFocus(document.getElementById(toFocus)); }, 100);
            }
        });
    };

    $scope.updateTime = function(index){
        $scope.form.Fields[index].Time = getTimeStamp();
    };

    function getDateString(){
        var date = $routeParams.date;
        if ( date == "today" ){
            var d = new Date();
            date = padStr(d.getDate())+padStr(d.getMonth()+1)+padStr(d.getFullYear());
        }
        return date;
    }

    $scope.$watch('form.Date', function(newValues, oldValues){
        if(typeof newValues != 'undefined' && newValues != ""){
            $scope.getState(newValues.replace(/-/g, ""));
        }
    });

    $scope.$watch('form.Fields', function(newValue,oldValue){
        for(var i=0; i< newValue.length; i++){
            var kmSaved = newValue[i].Km != 0 && newValue[i].Km == $scope.original.Fields[i].Km;
            var timeSaved = newValue[i].Time != 0 && newValue[i].Time == $scope.original.Fields[i].Time;
            if(kmSaved && timeSaved){
                $scope.form.Fields[i].Saved = true;
            }else{
                $scope.form.Fields[i].Saved = false;
            }
        }
    },true);

    $scope.save = function(name, fieldValue){
        var id = $scope.id || $routeParams.id;
        var toSave = [];
        var original = $scope.original.Fields;
        var form = $scope.form.Fields;
        for(var i=0; i<original.length; i++){
            if(original[i].Km != form[i].Km || original[i].Time != form[i].Time){
                toSave[toSave.length] = {Name: form[i].Name, Km: form[i].Km, Time: form[i].Time};
            }
        }
        console.log(toSave);
        if(toSave.length > 0){
            $http.post('/save/' + getDateString(), toSave).success(function(data){
                // als laatste element van toSave Terug is, klaar voor vandaag -> ga naar overview
                if(toSave[toSave.length-1].Name =="Terug"){
                    $location.path('/overview');
                }else{
                    $scope.getState(getDateString());
                }
            });
        }
    };

    $scope.goTo = function(address){
       $location.path(address);
    };
    $scope.valid = function(name){
        return $scope.kmform['{{field}}'].$error.integer;
    };

    function setFocus(el){
        el.focus();
        var strl = el.value.length;
        el.setSelectionRange(strl,strl);
    }
    $scope.getState(getDateString());
});

kmControllers.controller('kmOverviewController', function($scope,$routeParams, $location, $http, $filter){

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
    // Activate tab based on url
    if($routeParams.category === 'kilometers'){
        $scope.kiloActive = true;
    }
    if($routeParams.category === 'tijden'){
        $scope.timesActive = true;
    }

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

    $scope.deleteRow = function(date){
        $http.get('delete/' + $filter('date')(date, 'ddMMyyyy') ).success(function(data){
            $scope.loadData()
        });
    };

    $scope.edit = function(date){
        var dateId = $filter('date')(date, 'ddMMyyyy');
        $location.path('/input/' + dateId);
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
                        if(viewValue >= scope.form.Fields[0].Km){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id === 'Laatste'){
                        if(viewValue >= scope.form.Fields[1].Km){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id === 'Terug'){
                        if(viewValue >= scope.form.Fields[2].Km){
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
