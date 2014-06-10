// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

var handle;

function logoScreen(){
  // Initialize
  $('#container').html($('.template#logoScreen').html());
  $('.blinds.center:not(.blank)').css({rotateY: '-180deg'})

  // In Animation
  closeBlinds(function(){
    setTimeout(function(){
      $('.blinds.center.blank').transition({rotateY: '180deg'});
      $('.blinds.center:not(.blank)').transition({rotateY: '0deg'});
    }, 400);
  });

  // Close Function
  return function(callback){
    $('.blinds.center.blank').transition({rotateY: '360deg'});
    $('.blinds.center:not(.blank)').transition({rotateY: '180deg'}, function(){
      openBlinds(callback);
    });
  }
}

function closeBlinds(callback){
  $('.blinds.right').transition({right: 0});
  $('.blinds.left').transition({left: 0}, function(){
    $(this).addClass('full');
    callback();
  });
}

function openBlinds(callback){
  $('.blinds.right').show();
  $('.blinds.left').removeClass('full');
  $('.blinds.right').show().transition({right: '-50%'});
  $('.blinds.left').transition({left: '-50%'}, function(){
    callback();
  });
}
