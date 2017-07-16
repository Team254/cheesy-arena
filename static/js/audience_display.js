// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
var currentScreen = "blank";
var allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());

// Handles a websocket message to change which screen is displayed.
var handleSetAudienceDisplay = function(targetScreen) {
  if (targetScreen == currentScreen) {
    return;
  }

  transitions = transitionMap[currentScreen][targetScreen];
  if (transitions == null) {
    // There is no direct transition defined; need to go to the blank screen first.
    transitions = function() {
      transitionMap[currentScreen]["blank"](transitionMap["blank"][targetScreen]);
    };
  }
  transitions();

  currentScreen = targetScreen;
};

// Handles a websocket message to update the teams for the current match.
var handleSetMatch = function(data) {
  $("#redTeam1").text(data.Match.Red1)
  $("#redTeam2").text(data.Match.Red2)
  $("#redTeam3").text(data.Match.Red3)
  $("#blueTeam1").text(data.Match.Blue1)
  $("#blueTeam2").text(data.Match.Blue2)
  $("#blueTeam3").text(data.Match.Blue3)
  $("#matchName").text(data.MatchName + " " + data.Match.DisplayName);
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length == 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScoreNumber").text(data.RedScoreSummary.Score);
  $("#redPressurePoints").text(data.RedScoreSummary.PressurePoints);
  $("#redRotors").text(data.RedScoreSummary.Rotors);
  $("#redTakeoffs").text(data.RedScore.Takeoffs);
  $("#blueScoreNumber").text(data.BlueScoreSummary.Score);
  $("#bluePressurePoints").text(data.BlueScoreSummary.PressurePoints);
  $("#blueRotors").text(data.BlueScoreSummary.Rotors);
  $("#blueTakeoffs").text(data.BlueScore.Takeoffs);
};

// Handles a websocket message to populate the final score data.
var handleSetFinalScore = function(data) {
  $("#redFinalScore").text(data.RedScore.Score);
  $("#redFinalTeam1").text(data.Match.Red1);
  $("#redFinalTeam2").text(data.Match.Red2);
  $("#redFinalTeam3").text(data.Match.Red3);
  $("#redFinalAutoMobilityPoints").text(data.RedScore.AutoMobilityPoints);
  $("#redFinalPressurePoints").text(data.RedScore.PressurePoints);
  $("#redFinalRotorPoints").text(data.RedScore.RotorPoints);
  $("#redFinalTakeoffPoints").text(data.RedScore.TakeoffPoints);
  $("#redFinalFoulPoints").text(data.RedScore.FoulPoints);
  $("#redFinalPressureGoalReached").html(data.RedScore.PressureGoalReached ? "&#x2714;" : "&#x2718;");
  $("#redFinalPressureGoalReached").attr("data-checked", data.RedScore.PressureGoalReached);
  $("#redFinalRotorGoalReached").html(data.RedScore.RotorGoalReached ? "&#x2714;" : "&#x2718;");
  $("#redFinalRotorGoalReached").attr("data-checked", data.RedScore.RotorGoalReached);
  $("#blueFinalScore").text(data.BlueScore.Score);
  $("#blueFinalTeam1").text(data.Match.Blue1);
  $("#blueFinalTeam2").text(data.Match.Blue2);
  $("#blueFinalTeam3").text(data.Match.Blue3);
  $("#blueFinalAutoMobilityPoints").text(data.BlueScore.AutoMobilityPoints);
  $("#blueFinalPressurePoints").text(data.BlueScore.PressurePoints);
  $("#blueFinalRotorPoints").text(data.BlueScore.RotorPoints);
  $("#blueFinalTakeoffPoints").text(data.BlueScore.TakeoffPoints);
  $("#blueFinalFoulPoints").text(data.BlueScore.FoulPoints);
  $("#blueFinalPressureGoalReached").html(data.BlueScore.PressureGoalReached ? "&#x2714;" : "&#x2718;");
  $("#blueFinalPressureGoalReached").attr("data-checked", data.BlueScore.PressureGoalReached);
  $("#blueFinalRotorGoalReached").html(data.BlueScore.RotorGoalReached ? "&#x2714;" : "&#x2718;");
  $("#blueFinalRotorGoalReached").attr("data-checked", data.BlueScore.RotorGoalReached);
  $("#finalMatchName").text(data.MatchName + " " + data.Match.DisplayName);
};

// Handles a websocket message to play a sound to signal match start/stop/etc.
var handlePlaySound = function(sound) {
  $("audio").each(function(k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#" + sound)[0].play();
};

// Handles a websocket message to update the alliance selection screen.
var handleAllianceSelection = function(alliances) {
  if (alliances) {
    $.each(alliances, function(k, v) {
      v.Index = k + 1;
    });
    $("#allianceSelection").html(allianceSelectionTemplate(alliances));
  }
};

// Handles a websocket message to populate and/or show/hide a lower third.
var handleLowerThird = function(data) {
  if (data.BottomText == "") {
    $("#lowerThirdTop").hide();
    $("#lowerThirdBottom").hide();
    $("#lowerThirdSingle").text(data.TopText);
    $("#lowerThirdSingle").show();
  } else {
    $("#lowerThirdSingle").hide();
    $("#lowerThirdTop").text(data.TopText);
    $("#lowerThirdBottom").text(data.BottomText);
    $("#lowerThirdTop").show();
    $("#lowerThirdBottom").show();
  }
};

var transitionBlankToIntro = function(callback) {
  $("#centering").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $(".teams").transition({queue: false, width: "65px"}, 100, "linear", function() {
      $(".score").transition({queue: false, width: "120px"}, 500, "ease", function() {
        $("#eventMatchInfo").show();
        var height = -$("#eventMatchInfo").height();
        $("#eventMatchInfo").transition({queue: false, bottom: height + "px"}, 500, "ease", callback);
      });
    });
  });
};

var transitionIntroToInMatch = function(callback) {
  $("#logo").transition({queue: false, top: "10px"}, 500, "ease");
  $(".score").transition({queue: false, width: "250px"}, 500, "ease", function() {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
  });
};

var transitionIntroToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: "0px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "40px"}, 500, "ease", function() {
      $("#centering").transition({queue: false, bottom: "-340px"}, 1000, "ease", callback);
    });
  });
};

var transitionBlankToInMatch = function(callback) {
  $("#centering").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $(".teams").transition({queue: false, width: "65px"}, 100, "linear", function() {
      $("#logo").transition({queue: false, top: "10px"}, 500, "ease");
      $(".score").transition({queue: false, width: "250px"}, 500, "ease", function() {
        $("#eventMatchInfo").show();
        $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
        $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
        var height = -$("#eventMatchInfo").height();
        $("#eventMatchInfo").transition({queue: false, bottom: height + "px"}, 500, "ease", callback);
      });
    });
  });
}

var transitionInMatchToIntro = function(callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: "30px"}, 500, "ease");
    $(".score").transition({queue: false, width: "120px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "65px"}, 500, "ease", callback);
  });
};

var transitionInMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, bottom: "0px"}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $("#logo").transition({queue: false, top: "30px"}, 500, "ease");
    $(".score").transition({queue: false, width: "0px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "40px"}, 500, "ease", function() {
      $("#centering").transition({queue: false, bottom: "-340px"}, 1000, "ease", callback);
    });
  });
};

var transitionBlankToLogo = function(callback) {
  $(".blindsCenter.blank").css({rotateY: "0deg"});
  $(".blindsCenter.full").css({rotateY: "-180deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function() {
      $(".blindsCenter.blank").transition({queue: false, rotateY: "180deg"}, 500, "ease");
      $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 500, "ease", callback);
    }, 200);
  });
};

var transitionLogoToBlank = function(callback) {
  $(".blindsCenter.blank").transition({queue: false, rotateY: "360deg"}, 500, "ease");
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 500, "ease", function() {
    setTimeout(function() {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
    }, 200);
  });
};

var transitionLogoToScore = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: "-350px"}, 625, "ease");
  $("#finalScore").show();
  $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

var transitionBlankToScore = function(callback) {
  transitionBlankToLogo(function() {
    setTimeout(function() { transitionLogoToScore(callback); }, 50);
  });
};

var transitionScoreToLogo = function(callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "ease", function(){
    $("#finalScore").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0}, 625, "ease", callback);
};

var transitionScoreToBlank = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToBlank(callback);
  });
};

var transitionBlankToAllianceSelection = function(callback) {
  $('#allianceSelectionCentering').css("right","-60em").show();
  $('#allianceSelectionCentering').transition({queue: false, right: "3em"}, 500, "ease", callback);
};

var transitionAllianceSelectionToBlank = function(callback) {
  $('#allianceSelectionCentering').transition({queue: false, right: "-60em"}, 500, "ease", callback);
};

var transitionBlankToLowerThird = function(callback) {
  $("#lowerThird").show();
  $("#lowerThird").transition({queue: false, left: "150px"}, 750, "ease", callback);
};

var transitionLowerThirdToBlank = function(callback) {
  $("#lowerThird").transition({queue: false, left: "-1000px"}, 1000, "ease", function() {
    $("#lowerThird").hide();
    if (callback) {
      callback();
    }
  });
};

var transitionBlankToSponsor = function(callback) {
  $(".blindsCenter.blank").css({rotateY: "90deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function() {
      $("#sponsor").show();
      $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
    }, 200);
  });
};

var transitionSponsorToBlank = function(callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function() {
    setTimeout(function() {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
      $("#sponsor").hide();
    }, 200);
  });
};

var transitionLogoToSponsor = function(callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "90deg"}, 750, "ease", function () {
    $("#sponsor").show();
    $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionSponsorToLogo = function(callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function() {
    $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 750, "ease", callback);
    $("#sponsor").hide();
  });
};

var transitionScoreToSponsor = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToSponsor(callback);
  });
};

var transitionSponsorToScore = function(callback) {
  transitionSponsorToLogo(function() {
    transitionLogoToScore(callback);
  });
};

// Loads sponsor slide data and builds the slideshow HTML.
var initializeSponsorDisplay = function() {
  $.getJSON("/api/sponsor_slides", function(sponsors) {
    if (!sponsors) {
      return;
    }

    // Populate Tiles
    $.each(sponsors, function(index){
      var active = 'active';
      if(index)
        active = '';

      if(sponsors[index]['Image'].length)
        $('#sponsorContainer').append('<div class="item '+active+'" data-interval="'+sponsors[index]["DisplayTimeSec"]*1000+'"><img src="/static/img/sponsors/'+sponsors[index]['Image']+'" /><h1>'+sponsors[index]['Subtitle']+'</h1></div>');
      else
        $('#sponsorContainer').append('<div class="item '+active+'" data-interval="'+sponsors[index]["DisplayTimeSec"]*1000+'"><h2>'+sponsors[index]['Line1']+'<br />'+sponsors[index]['Line2']+'</h2><h1>'+sponsors[index]['Subtitle']+'</h1></div>');

    });

    // Start Carousel
    var t;
    var start = $('.carousel#sponsor').find('.active').attr('data-interval');
    t = setTimeout("$('.carousel#sponsor').carousel({interval: 1000});", start-1000);

    $('.carousel#sponsor').on('slid.bs.carousel', function () {   
         clearTimeout(t);  
         var duration = $(this).find('.active').attr('data-interval');

         $('.carousel#sponsor').carousel('pause');
         t = setTimeout("$('.carousel#sponsor').carousel();", duration-1000);
    })

    $('.carousel-control.right').on('click', function(){
        clearTimeout(t);   
    });

    $('.carousel-control.left').on('click', function(){
        clearTimeout(t);   
    });

  });
}

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/audience/websocket", {
    setAudienceDisplay: function(event) { handleSetAudienceDisplay(event.data); },
    setMatch: function(event) { handleSetMatch(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    setFinalScore: function(event) { handleSetFinalScore(event.data); },
    playSound: function(event) { handlePlaySound(event.data); },
    allianceSelection: function(event) { handleAllianceSelection(event.data); },
    lowerThird: function(event) { handleLowerThird(event.data); }
  });

  initializeSponsorDisplay();

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    blank: {
      intro: transitionBlankToIntro,
      match: transitionBlankToInMatch,
      score: transitionBlankToScore,
      logo: transitionBlankToLogo,
      sponsor: transitionBlankToSponsor,
      allianceSelection: transitionBlankToAllianceSelection,
      lowerThird: transitionBlankToLowerThird
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToInMatch
    },
    match: {
      blank: transitionInMatchToBlank,
      intro: transitionInMatchToIntro
    },
    score: {
      blank: transitionScoreToBlank,
      logo: transitionScoreToLogo,
      sponsor: transitionScoreToSponsor
    },
    logo: {
      blank: transitionLogoToBlank,
      score: transitionLogoToScore,
      sponsor: transitionLogoToSponsor
    },
    sponsor: {
      blank: transitionSponsorToBlank,
      logo: transitionSponsorToLogo,
      score: transitionSponsorToScore
    },
    allianceSelection: {
      blank: transitionAllianceSelectionToBlank
    },
    lowerThird: {
      blank: transitionLowerThirdToBlank
    }
  }
});
