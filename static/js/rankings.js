// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

'use strict';

var app = angular
.module('cheesyArenaRankings', []);

app.controller('RankingsController', ['$scope', 'Rankings', '$interval', function($scope, Rankings, $interval){
  
  $scope.date = new Date();

  // Load Rankings
  init();
  function init(){
    Rankings.get().then(function(data){
      $scope.rankingsOld = data;
      $scope.rankingsNew = data;
      setTimeout(equalize, 0);
      setTimeout(detectListSize, 0);
    });
  }

  // Detect if List is Long Enough to Require Scrolling
  function detectListSize(){
    $('.new').hide();
    if($('#container table').height() > $('#container').height()){
      $('.new').show();
      setTimeout(scrollToBottom, 0);
    }
  }

  // Scroll to Bottom of List
  function scrollToBottom(){
    var offset = $('#container').offset().top;
    $('#container').animate({scrollTop: $('.old').first().offset().top - offset}, 0);
    // $scope.interval = $interval(loadNewData, 10);
    var time = 1000 * $scope.rankingsOld.length;
    $('#container').animate({scrollTop: $('.new').first().offset().top - offset}, time, 'linear', scrollComplete);
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
        setTimeout(equalize, 0);
      });
    }
  }

  // Balance Column Widths
  function equalize(){
    var width = $('#container table').width();
    var count = $('#container tr').first().children('td').length;
    var offset = ($(window).width() - width) / (count + 1);
    var widths = [];
    $('#container tr').first().children('td').each(function(){
      var width = $(this).width()+offset;
      $(this).width(width);
      widths.push(width);
    });
    $('#header').children('td').each(function(index){
      $(this).width(widths[index]);
    });
    $('#container tr.new').children('td').each(function(index){
      $(this).width(widths[index]);
    });
  }

}]);

app.factory('Rankings', ['$http', '$log', '$q', function($http, $log, $q){
  return {
    get: function(){
      var deferred = $q.defer();
      $http.get('/reports/json/rankings').
        success(function(data, status, headers, config) {
          // data[0]['TeamId'] = jQuery.now();
          deferred.resolve(data);
        }).
        error(function(data, status, headers, config) {
          $log.error(data);
        });
        return deferred.promise;
    }
  };
}]);
