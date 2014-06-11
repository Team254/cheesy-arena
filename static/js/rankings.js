// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

initialize();

var template;
function initialize(){
  getData(populateView);
  template = Handlebars.compile($('#row-template').html());
}

var rankings;
function getData(callback){
  $.getJSON('/reports/json/rankings', function(data){
    data = {teams: data};
    rankings = data;
    if(typeof(callback) == "function")
      callback();
    var date = new Date();
    console.log("New Data Acquired\n"+date);
  });
}

function populateView(){
  $('#container table').html(template(rankings));
  equalize();
  setTimeout(scroll, PAUSE_TIME);
}

// Balance Column Widths
var widths = [];
function equalize(){
  $('#container #new tr').first().children('td').each(function(){
    var width = $(this).width();
    widths.push(width);
  });
  $('#header').children('td').each(function(index){
    $(this).width(widths[index]);
  });
}

var SCROLL_SPEED = 1000;  // Smaller is Faster
function scroll(){
  $('#container').scrollTop(0);

  var offset = $('table#new').offset().top - $('#container').offset().top;
  var scrollTime = SCROLL_SPEED * $('table#old tr').length;
  $('#container').animate({scrollTop: offset}, scrollTime, 'linear', reset);

  $('#container table#new').html(template(rankings));
  equalize();

  interval = setInterval(pollForUpdate, POLL_INTERVAL);
}

var PAUSE_TIME = 5000;
function reset(){
  $('#container table#old').html($('#container table#new').html());
  setTimeout(scroll, PAUSE_TIME);
}

var POLL_INTERVAL = 1000;
function pollForUpdate(){
  if($('#container').offset().top * $('#container').height() > $('#container table#old tr').last().prev().offset().top){
    getData();
    clearInterval(interval);
  }
}