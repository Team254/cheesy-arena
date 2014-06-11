// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

var screens = {

  blank: {
    init: function(cb){callback(cb);},
    open: function(cb){callback(cb);},
    close: function(cb){callback(cb);}
  },

  logo: {
    init: function(cb){
      $('.blinds.center:not(.blank)').css({rotateY: '-180deg'});
      callback(cb);
    },
    open: function(cb){
      closeBlinds(function(){
        setTimeout(function(){
          $('.blinds.center.blank').transition({rotateY: '180deg'});
          $('.blinds.center:not(.blank)').transition({rotateY: '0deg'}, function(){
            callback(cb);
          });
        }, 400);
      });
    },
    close: function(cb){
      $('.blinds.center.blank').transition({rotateY: '360deg'});
      $('.blinds.center:not(.blank)').transition({rotateY: '180deg'}, function(){
        openBlinds(callback);
      });
    }
  }

};

var currentScreen = 'blank';
function openScreen(screen){

  // If Screen Exists
  if(typeof(screens[screen]) == 'object' && $('.template#'+screen).length > 0 && currentScreen != screen){

    // Initialize New Screen
    $('#topcontainer').append("<div class='container' id='"+screen+"'>"+$('.template#'+screen).html()+"</div>");
    screens[screen].init(function(){

      // Close Current Screen
      screens[currentScreen].close(function(){

        // Open New Screen
        currentScreen = screen;
        screens[screen].open();

      });
    });
  }
}

function callback(cb){
  if(typeof(cb) == 'function')
    cb();
}

function closeBlinds(cb){
  $('.blinds.right').transition({right: 0});
  $('.blinds.left').transition({left: 0}, function(){
    $(this).addClass('full');
    callback(cb);
  });
}

function openBlinds(cb){
  $('.blinds.right').show();
  $('.blinds.left').removeClass('full');
  $('.blinds.right').show().transition({right: '-50%'});
  $('.blinds.left').transition({left: '-50%'}, function(){
    callback(cb);
  });
}
