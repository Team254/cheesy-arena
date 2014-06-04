// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

'use strict';

var app = angular
.module('cheesyArenaRankings', []);

app.controller('RankingsController', ['$scope', 'Rankings', '$interval', function($scope, Rankings, $interval){
  
  // Load Rankings
  init();
  function init(){
    Rankings.get().then(function(data){
      $scope.rankingsOld = data;
      $scope.rankingsNew = data;
      setTimeout(detectListSize, 0);
    });
  }

  // Detect if List is Long Enough to Require Scrolling
  function detectListSize(){
    $('.new').hide();
    if($(document).height() > $(window).height()){
      $('.new').show();
      setTimeout(scrollToBottom, 0);
    }
  }

  // Scroll to Bottom of List
  function scrollToBottom(){
    $('body').animate({scrollTop: $('.old').first().offset().top}, 0);
    $scope.interval = $interval(loadNewData, 10);
    var time = 500 * $scope.rankingsOld.length;
    $('body').animate({scrollTop: $('.new').first().offset().top}, time, 'linear', scrollComplete);
  }

  // Go Back to Top
  function scrollComplete(){
    $scope.rankingsOld = $scope.rankingsNew;
    setTimeout(scrollToBottom, 0);
  }

  // Load New Data When 2 Elements Away from End of List
  function loadNewData(){
    var position = $(window).scrollTop()+$(window).height();
    var offset = $('.old:last').prev().prev().offset().top;
    if(position >= offset){
      Rankings.get().then(function(data){
        $scope.rankingsOld = $scope.rankingsNew;
        $scope.rankingsNew = data;
        $interval.cancel($scope.interval);
      });
    }
  }

}]);

app.factory('Rankings', ['$http', '$log', '$q', function($http, $log, $q){
  return {
    get: function(){
      var deferred = $q.defer();
      $http.get('/reports/json/rankings').
        success(function(data, status, headers, config) {
          data[0]['TeamId'] = jQuery.now();
          deferred.resolve(data);
        }).
        error(function(data, status, headers, config) {
          $log.error(data);
        });
        return deferred.promise;
    }
  };
}]);