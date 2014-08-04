// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
var currentScreen = "blank";

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
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    $(".blinds.center-blank").css({rotateY: "0deg"});
    setTimeout(function() {
      $(".blinds.center-blank").transition({queue: false, rotateY: "180deg"}, 500, "ease");
      $("#blindsCenter").transition({queue: false, rotateY: "0deg"}, 500, "ease", callback);
    }, 200);
  });
};

var transitionLogoToBlank = function(callback) {
  $(".blinds.center-blank").transition({queue: false, rotateY: "360deg"}, 500, "ease");
  $("#blindsCenter").transition({queue: false, rotateY: "180deg"}, 500, "ease", function() {
    setTimeout(function() {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
    }, 200);
  });
};

var transitionLogoToScore = function(callback) {
  $("#blindsCenter").transition({queue: false, top: "-350px"}, 750, "ease", function () {
    $("#finalScore").show();
    $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionBlankToScore = function(callback) {
  transitionBlankToLogo(function() {
    setTimeout(function() { transitionLogoToScore(callback); }, 100);
  });
};

var transitionScoreToLogo = function(callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "linear", function() {
    $("#finalScore").hide();
    $("#blindsCenter").transition({queue: false, top: 0}, 750, "ease", callback);
  });
};

var transitionScoreToBlank = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToBlank(callback);
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
    playSound: function(event) { handlePlaySound(event.data); }
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    blank: {
      intro: transitionBlankToIntro,
      match: transitionBlankToInMatch,
      score: transitionBlankToScore,
      logo: transitionBlankToLogo
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
      logo: transitionScoreToLogo
    },
    logo: {
      blank: transitionLogoToBlank,
      score: transitionLogoToScore
    }
  }
});
