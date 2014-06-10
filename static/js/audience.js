// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre) 

var handle;

function logoScreen(){
  // Initialize
  $('#logoScreen .blinds.center:not(.blank)').css({rotateY: '-180deg'})

  // In Animation
  closeBlinds('logoScreen', function(){
    setTimeout(function(){
      $('#logoScreen .blinds.center.blank').transition({rotateY: '180deg'});
      $('#logoScreen .blinds.center:not(.blank)').transition({rotateY: '0deg'});
    }, 400);
  });

  // Close Function
  return function(callback){
    $('#logoScreen .blinds.center.blank').transition({rotateY: '360deg'});
    $('#logoScreen .blinds.center:not(.blank)').transition({rotateY: '180deg'}, function(){
      openBlinds('logoScreen', callback);
    });
  }
}

function closeBlinds(id, callback){
  $('#'+id+' .blinds.right').transition({right: 0});
  $('#'+id+' .blinds.left').transition({left: 0}, function(){
    $(this).addClass('full');
    callback();
  });
}

function openBlinds(id, callback){
  $('#'+id+' .blinds.right').show();
  $('#'+id+' .blinds.left').removeClass('full');
  $('#'+id+' .blinds.right').show().transition({right: '-50%'});
  $('#'+id+' .blinds.left').transition({left: '-50%'}, function(){
    callback();
  });
}
