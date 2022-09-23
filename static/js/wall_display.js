// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
var currentScreen = "blank";
var redSide;
var blueSide;
var currentMatch;

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "23px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "135px";
const scoreOut = "255px";
const scoreFieldsOut = "40px";

// Handles a websocket message to change which screen is displayed.
var handleAudienceDisplayMode = function(targetScreen) {
  if (
      targetScreen !== "intro" &&
      targetScreen !== "match" &&
      targetScreen !== "timeout"
  ) {
    targetScreen = "blank";
  }

  transitionQueue.push(targetScreen);
  executeTransitionQueue();
};

// Sequentially executes all transitions in the queue. Returns without doing anything if another invocation is already
// in progress.
const executeTransitionQueue = function() {
  if (transitionInProgress) {
    // There is an existing invocation of this method which will execute all transitions in the queue.
    return;
  }

  if (transitionQueue.length > 0) {
    transitionInProgress = true;
    const targetScreen = transitionQueue.shift();
    const callback = function() {
      // When the current transition is complete, call this method again to invoke the next one in the queue.
      currentScreen = targetScreen;
      transitionInProgress = false;
      setTimeout(executeTransitionQueue, 100);  // A small delay is needed to avoid visual glitches.
    };

    if (targetScreen === currentScreen) {
      callback();
      return;
    }

    let transitions = transitionMap[currentScreen][targetScreen];
    if (transitions !== undefined) {
      transitions(callback);
    } else {
      // There is no direct transition defined; need to go to the blank screen first.
      transitionMap[currentScreen]["blank"](function() {
        transitionMap["blank"][targetScreen](callback);
      });
    }
  }
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  currentMatch = data.Match;
  $("#" + redSide + "Team1").text(currentMatch.Red1);
  $("#" + redSide + "Team2").text(currentMatch.Red2);
  $("#" + redSide + "Team3").text(currentMatch.Red3);
  $("#" + redSide + "Team1Avatar").attr("src", getAvatarUrl(currentMatch.Red1));
  $("#" + redSide + "Team2Avatar").attr("src", getAvatarUrl(currentMatch.Red2));
  $("#" + redSide + "Team3Avatar").attr("src", getAvatarUrl(currentMatch.Red3));
  $("#" + blueSide + "Team1").text(currentMatch.Blue1);
  $("#" + blueSide + "Team2").text(currentMatch.Blue2);
  $("#" + blueSide + "Team3").text(currentMatch.Blue3);
  $("#" + blueSide + "Team1Avatar").attr("src", getAvatarUrl(currentMatch.Blue1));
  $("#" + blueSide + "Team2Avatar").attr("src", getAvatarUrl(currentMatch.Blue2));
  $("#" + blueSide + "Team3Avatar").attr("src", getAvatarUrl(currentMatch.Blue3));

  // Show alliance numbers if this is an elimination match.
  if (currentMatch.Type === "elimination") {
    $("#" + redSide + "ElimAlliance").text(currentMatch.ElimRedAlliance);
    $("#" + blueSide + "ElimAlliance").text(currentMatch.ElimBlueAlliance);
    $(".elim-alliance").show();

    // Show the series status if this playoff round isn't just a single match.
    if (data.Matchup.NumWinsToAdvance > 1) {
      $("#" + redSide + "ElimAllianceWins").text(data.Matchup.RedAllianceWins);
      $("#" + blueSide + "ElimAllianceWins").text(data.Matchup.BlueAllianceWins);
      $("#elimSeriesStatus").css("display", "flex");
    } else {
      $("#elimSeriesStatus").hide();
    }
  } else {
    $("#" + redSide + "ElimAlliance").text("");
    $("#" + blueSide + "ElimAlliance").text("");
    $(".elim-alliance").hide();
    $("#elimSeriesStatus").hide();
  }

  if (data.Match.Type === "test") {
    $("#matchName").text(currentMatch.DisplayName);
  } else {
    $("#matchName").text(data.MatchType + " " + currentMatch.DisplayName);
  }
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#" + redSide + "ScoreNumber").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.HangarPoints);
  $("#" + blueSide + "ScoreNumber").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.HangarPoints);

  $("#" + redSide + "CargoNumerator").text(data.Red.ScoreSummary.CargoCount);
  $("#" + redSide + "CargoDenominator").text(data.Red.ScoreSummary.CargoGoal);
  $("#" + blueSide + "CargoNumerator").text(data.Blue.ScoreSummary.CargoCount);
  $("#" + blueSide + "CargoDenominator").text(data.Blue.ScoreSummary.CargoGoal);
  if (currentMatch.Type === "elimination") {
    $("#" + redSide + "CargoDenominator").hide();
    $("#" + blueSide + "CargoDenominator").hide();
    $(".cargo-splitter").hide();
  } else {
    $("#" + redSide + "CargoDenominator").show();
    $("#" + blueSide + "CargoDenominator").show();
    $(".cargo-splitter").show();
  }
};

var transitionBlankToIntro = function(callback) {
  //$("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function() {
    $(".teams").css("display", "flex");
    $(".avatars").css("display", "flex");
    $(".avatars").css("opacity", 1);
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function() {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
    });
  //});
};

var transitionBlankToMatch = function(callback) {
  //$("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function() {
    $(".teams").css("display", "flex");
    $(".score-fields").css("display", "flex");
    $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
    $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function() {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    });
  //});
};

var transitionBlankToTimeout = function(callback) {
  //$("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  //});
};

var transitionIntroToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      //$("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
      callback();
    });
  });
};

var transitionIntroToMatch = function(callback) {
  $(".avatars").transition({queue: false, opacity: 0}, 500, "ease", function() {
    $(".avatars").hide();
  });
  $(".score-fields").css("display", "flex");
  $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
  $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
  $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function() {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
  });
};

var transitionIntroToTimeout = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

var transitionMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".teams").hide();
      $(".score-fields").hide();
      //$("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
      callback();
    });
  });
};

var transitionMatchToIntro = function(callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function() {
      $(".score-fields").hide();
      $(".avatars").css("display", "flex");
      $(".avatars").transition({queue: false, opacity: 1}, 500, "ease", callback);
    });
  });
};

var transitionTimeoutToBlank = function(callback) {
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function() {
      //$("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
      callback();
    });
  });
};

var transitionTimeoutToIntro = function(callback) {
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function() {
      $(".avatars").css("display", "flex");
      $(".avatars").css("opacity", 1);
      $(".teams").css("display", "flex");
      $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
        $("#eventMatchInfo").show();
        $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      });
    });
  });
};

var getAvatarUrl = function(teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

$(function() {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");
  var reversed = urlParams.get("reversed");
  if (reversed === "true") {
    redSide = "right";
    blueSide = "left";
  } else {
    redSide = "left";
    blueSide = "right";
  }
  $(".reversible-left").attr("data-reversed", reversed);
  $(".reversible-right").attr("data-reversed", reversed);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/wall/websocket", {
    allianceSelection: function(event) { handleAllianceSelection(event.data); },
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    blank: {
      intro: transitionBlankToIntro,
      match: transitionBlankToMatch,
      timeout: transitionBlankToTimeout,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    match: {
      blank: transitionMatchToBlank,
      intro: transitionMatchToIntro,
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro,
    },
  }
});
