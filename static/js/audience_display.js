// Copyright 2014 Team 254. All Rights Reserved.
// Authors: pat@patfairbank.com (Patrick Fairbank)
//          nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
var currentScreen = "blank";
var allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());

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

var handleSetMatch = function(data) {
  $("#redTeam1").text(data.Match.Red1)
  $("#redTeam2").text(data.Match.Red2)
  $("#redTeam3").text(data.Match.Red3)
  $("#blueTeam1").text(data.Match.Blue1)
  $("#blueTeam2").text(data.Match.Blue2)
  $("#blueTeam3").text(data.Match.Blue3)
  $("#matchName").text(data.MatchName + " " + data.Match.DisplayName);
};

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

var handleRealtimeScore = function(data) {
  $("#redScoreNumber").text(data.RedScore);
  $("#redAssist1").attr("data-on", data.RedCycle.Assists >= 1);
  $("#redAssist2").attr("data-on", data.RedCycle.Assists >= 2);
  $("#redAssist3").attr("data-on", data.RedCycle.Assists >= 3);
  $("#redTruss").attr("data-on", data.RedCycle.Truss);
  $("#redCatch").attr("data-on", data.RedCycle.Catch);
  $("#blueScoreNumber").text(data.BlueScore);
  $("#blueAssist1").attr("data-on", data.BlueCycle.Assists >= 1);
  $("#blueAssist2").attr("data-on", data.BlueCycle.Assists >= 2);
  $("#blueAssist3").attr("data-on", data.BlueCycle.Assists >= 3);
  $("#blueTruss").attr("data-on", data.BlueCycle.Truss);
  $("#blueCatch").attr("data-on", data.BlueCycle.Catch);
};

var handleSetFinalScore = function(data) {
  $("#redFinalScore").text(data.RedScore.Score);
  $("#redFinalTeam1").text(data.Match.Red1);
  $("#redFinalTeam2").text(data.Match.Red2);
  $("#redFinalTeam3").text(data.Match.Red3);
  $("#redFinalAuto").text(data.RedScore.AutoPoints);
  $("#redFinalTeleop").text(data.RedScore.TeleopPoints);
  $("#redFinalFoul").text(data.RedScore.FoulPoints);
  $("#blueFinalScore").text(data.BlueScore.Score);
  $("#blueFinalTeam1").text(data.Match.Blue1);
  $("#blueFinalTeam2").text(data.Match.Blue2);
  $("#blueFinalTeam3").text(data.Match.Blue3);
  $("#blueFinalAuto").text(data.BlueScore.AutoPoints);
  $("#blueFinalTeleop").text(data.BlueScore.TeleopPoints);
  $("#blueFinalFoul").text(data.BlueScore.FoulPoints);
  $("#finalMatchName").text(data.MatchName + " " + data.Match.DisplayName);
};

var handlePlaySound = function(sound) {
  $("audio").each(function(k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#" + sound)[0].play();
};

var handleAllianceSelection = function(alliances) {
  if (alliances) {
    $.each(alliances, function(k, v) {
      v.Index = k + 1;
    });
    $("#allianceSelection").html(allianceSelectionTemplate(alliances));
  }
};

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
    $(".teams").transition({queue: false, width: "75px"}, 100, "linear", function() {
      $(".score").transition({queue: false, width: "120px"}, 500, "ease", function() {
        $("#eventMatchInfo").show();
        var height = -$("#eventMatchInfo").height();
        $("#eventMatchInfo").transition({queue: false, bottom: height + "px"}, 500, "ease", callback);
      });
    });
  });
};

var transitionIntroToInMatch = function(callback) {
  $("#logo").transition({queue: false, top: "25px"}, 500, "ease");
  $(".score").transition({queue: false, width: "230px"}, 500, "ease", function() {
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
    $(".teams").transition({queue: false, width: "75px"}, 100, "linear", function() {
      $("#logo").transition({queue: false, top: "25px"}, 500, "ease");
      $(".score").transition({queue: false, width: "230px"}, 500, "ease", function() {
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
    $("#logo").transition({queue: false, top: "45px"}, 500, "ease");
    $(".score").transition({queue: false, width: "120px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "75px"}, 500, "ease", callback);
  });
};

var transitionInMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, bottom: "0px"}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $("#logo").transition({queue: false, top: "45px"}, 500, "ease");
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
}

var transitionBlankToAllianceSelection = function(callback) {
  $("#allianceSelectionCentering").show();
  if (callback) {
    callback();
  }
};

var transitionAllianceSelectionToBlank = function(callback) {
  $("#allianceSelectionCentering").hide();
  if (callback) {
    callback();
  }
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

// Load and Prioritize Sponsor Data
var sponsors;
var sponsorIndex = [];
var initializeSponsorDisplay = function() {
  $.getJSON("/api/sponsors", function(data) {
    sponsors = data;

    // Invert Priorities
    $.each(sponsors, function(index){
      var priorityCount = 10 / sponsors[index]['Priority'];
      for(i=0; i<priorityCount; i++){
        sponsorIndex.push(index);
      }
    });

    // Load Tiles
    loadNextSponsor(true);
    if(sponsors.length > 1)
      loadNextSponsor();
    $('.carousel#sponsor').on('slid.bs.carousel', function(){
      loadNextSponsor();
      $('#sponsorContainer').children()[0].remove();
    });
  });
}
var lastSponsor;
var loadNextSponsor = function(active) {
  // Don't load same sponsor twice in a row
  var currentSponsor = sponsorIndex[Math.round(Math.random()*sponsorIndex.length)];
  while(currentSponsor == lastSponsor){
    currentSponsor = sponsorIndex[Math.round(Math.random()*sponsorIndex.length)];
  }
  lastSponsor = currentSponsor;
  currentSponsor = sponsors[currentSponsor];

  if(active == true)
    active = 'active'
  else
    active = '';
  
  if(currentSponsor['Image'].length)
    $('#sponsorContainer').append('<div class="item '+active+'"><img src="/static/img/sponsors/'+currentSponsor['Image']+'" /><h1>'+currentSponsor['Subtitle']+'</h1></div>');
  else
    $('#sponsorContainer').append('<div class="item '+active+'"><h2>'+currentSponsor['Line1']+'<br />'+currentSponsor['Line2']+'</h2><h1>'+currentSponsor['Subtitle']+'</h1></div>');
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
