'use strict';
var km = angular.module('km', []);

km.controller('KmCtrl', function($scope, $http){
    $http.get('state').success(function(data){
        $scope.form = data
        var toFocus;
        if (data.Begin.Editable == true){
            toFocus = 'Begin';
        }else if(data.Arnhem.Editable == true){
            toFocus = 'Arnhem';
        }else if(data.Laatste.Editable == true){
            toFocus = 'Laatste';
        }else if(data.Terugkomst.Editable == true){
            toFocus = 'Terugkomst';
        }
        setTimeout(function(){ setFocus(document.getElementById(toFocus)) }, 100);
    });


    $scope.save = function(name, fieldValue){
        $http.post('/save', {name: name, value: fieldValue});
        eval('$scope.form.'+name+'.Editable=false')
    };

    $scope.edit = function(name){
        eval('$scope.form.'+name+'.Editable=true')
    };

    function setFocus(el){
        el.focus();
        var strl = el.value.length;
        el.setSelectionRange(strl,strl);
    }
});

var INTEGER_REGEXP = /^\-?\d*$/;
km.directive('integer', function() {
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
                    if(attrs.id == 'Arnhem'){
                        if(viewValue > scope.form.Begin.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id == 'Laatste'){
                        if(viewValue > scope.form.Arnhem.Value){
                            ctrl.$setValidity('integer', true);
                            return viewValue;
                        }
                    }
                    if(attrs.id == 'Terugkomst'){
                        if(viewValue > scope.form.Arnhem.Value){
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

km.directive('ngEnter', function () {
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
