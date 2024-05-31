// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the wall display.

var websocket;
let transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
let currentScreen = "blank";
let redSide;
let blueSide;
let currentMatch;

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "30px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "185px";
const scoreOut = "250px";
const scoreFieldsOut = "25px";
const overlayTopOffset = 110;
const timeoutDetailsIn = $("#timeoutDetails").css("width");
const timeoutDetailsOut = "570px";

// Game-specific constants and variables.
const amplifyProgressStartOffset = $("#leftAmplified svg circle").css("stroke-dashoffset");
const amplifyFadeTimeMs = 300;
const amplifyDwellTimeMs = 500;
let redAmplified = false;
let blueAmplified = false;

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function(targetScreen) {
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
const handleMatchLoad = function(data) {
  currentMatch = data.Match;
  $(`#${redSide}Team1`).text(currentMatch.Red1);
  $(`#${redSide}Team1`).attr("data-yellow-card", data.Teams["R1"]?.YellowCard);
  $(`#${redSide}Team2`).text(currentMatch.Red2);
  $(`#${redSide}Team2`).attr("data-yellow-card", data.Teams["R2"]?.YellowCard);
  $(`#${redSide}Team3`).text(currentMatch.Red3);
  $(`#${redSide}Team3`).attr("data-yellow-card", data.Teams["R3"]?.YellowCard);
  $(`#${redSide}Team1Avatar`).attr("src", getAvatarUrl(currentMatch.Red1));
  $(`#${redSide}Team2Avatar`).attr("src", getAvatarUrl(currentMatch.Red2));
  $(`#${redSide}Team3Avatar`).attr("src", getAvatarUrl(currentMatch.Red3));
  $(`#${blueSide}Team1`).text(currentMatch.Blue1);
  $(`#${blueSide}Team1`).attr("data-yellow-card", data.Teams["B1"]?.YellowCard);
  $(`#${blueSide}Team2`).text(currentMatch.Blue2);
  $(`#${blueSide}Team2`).attr("data-yellow-card", data.Teams["B2"]?.YellowCard);
  $(`#${blueSide}Team3`).text(currentMatch.Blue3);
  $(`#${blueSide}Team3`).attr("data-yellow-card", data.Teams["B3"]?.YellowCard);
  $(`#${blueSide}Team1Avatar`).attr("src", getAvatarUrl(currentMatch.Blue1));
  $(`#${blueSide}Team2Avatar`).attr("src", getAvatarUrl(currentMatch.Blue2));
  $(`#${blueSide}Team3Avatar`).attr("src", getAvatarUrl(currentMatch.Blue3));

  // Show alliance numbers if this is a playoff match.
  if (currentMatch.Type === matchTypePlayoff) {
    $("#" + redSide + "PlayoffAlliance").text(currentMatch.PlayoffRedAlliance);
    $("#" + blueSide + "PlayoffAlliance").text(currentMatch.PlayoffBlueAlliance);
    $(".playoff-alliance").show();

    // Show the series status if this playoff round isn't just a single match.
    if (data.Matchup.NumWinsToAdvance > 1) {
      $("#" + redSide + "PlayoffAllianceWins").text(data.Matchup.RedAllianceWins);
      $("#" + blueSide + "PlayoffAllianceWins").text(data.Matchup.BlueAllianceWins);
      $("#playoffSeriesStatus").css("display", "flex");
    } else {
      $("#playoffSeriesStatus").hide();
    }
  } else {
    $("#" + redSide + "PlayoffAlliance").text("");
    $("#" + blueSide + "PlayoffAlliance").text("");
    $(".playoff-alliance").hide();
    $("#playoffSeriesStatus").hide();
  }

  let matchName = data.Match.LongName;
  if (data.Match.NameDetail !== "") {
    matchName += " &ndash; " + data.Match.NameDetail;
  }
  $("#matchName").html(matchName);
  $("#timeoutNextMatchName").html(matchName);
  $("#timeoutBreakDescription").text(data.BreakDescription);
};

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchTime").text(getCountdownString(countdownSec));
  });
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function(data) {
  $("#" + redSide + "ScoreNumber").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.StagePoints);
  $("#" + blueSide + "ScoreNumber").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.StagePoints);

  $(`#${redSide}NoteNumerator`).text(data.Red.ScoreSummary.NumNotes);
  $(`#${redSide}NoteDenominator`).text(data.Red.ScoreSummary.NumNotesGoal);
  $(`#${blueSide}NoteNumerator`).text(data.Blue.ScoreSummary.NumNotes);
  $(`#${blueSide}NoteDenominator`).text(data.Blue.ScoreSummary.NumNotesGoal);
  if (currentMatch.Type === matchTypePlayoff) {
    $(`#${redSide}NoteDenominator`).hide();
    $(`#${blueSide}NoteDenominator`).hide();
    $(".note-splitter").hide();
  } else {
    $(`#${redSide}NoteDenominator`).show();
    $(`#${blueSide}NoteDenominator`).show();
    $(".note-splitter").show();
  }

  const redLightsDiv = $(`#${redSide}Lights`);
  const redAmplifiedDiv = $(`#${redSide}Amplified`);
  if (data.Red.AmplifiedTimeRemainingSec > 0 && !redAmplified) {
    redAmplified = true;
    redLightsDiv.transition({queue: false, opacity: 0}, amplifyFadeTimeMs, "linear", function() {
      redLightsDiv.hide();
      redAmplifiedDiv.show();
      redAmplifiedDiv.transition({queue: false, opacity: 1}, amplifyFadeTimeMs, "linear");
      $(`#${redSide}Amplified svg circle`).transition(
        {queue: false, strokeDashoffset: 158}, data.Red.AmplifiedTimeRemainingSec * 1000 - amplifyFadeTimeMs, "linear"
      );
    });
  } else if (data.Red.AmplifiedTimeRemainingSec === 0 && redAmplified) {
    redAmplified = false;
    setTimeout(function() {
      redAmplifiedDiv.transition({queue: false, opacity: 0}, amplifyFadeTimeMs, "linear", function () {
        $(`#${redSide}Amplified svg circle`).css("stroke-dashoffset", amplifyProgressStartOffset);
        redAmplifiedDiv.hide();
        redLightsDiv.show();
        redLightsDiv.transition({queue: false, opacity: 1}, amplifyFadeTimeMs, "linear");
      });
    }, amplifyDwellTimeMs);
  }

  const blueLightsDiv = $(`#${blueSide}Lights`);
  const blueAmplifiedDiv = $(`#${blueSide}Amplified`);
  if (data.Blue.AmplifiedTimeRemainingSec > 0 && !blueAmplified) {
    blueAmplified = true;
    blueLightsDiv.transition({queue: false, opacity: 0}, amplifyFadeTimeMs, "linear", function() {
      blueLightsDiv.hide();
      blueAmplifiedDiv.show();
      blueAmplifiedDiv.transition({queue: false, opacity: 1}, amplifyFadeTimeMs, "linear");
      $(`#${blueSide}Amplified svg circle`).transition(
        {queue: false, strokeDashoffset: -158}, data.Blue.AmplifiedTimeRemainingSec * 1000 - amplifyFadeTimeMs, "linear"
      );
    });
  } else if (data.Blue.AmplifiedTimeRemainingSec === 0 && blueAmplified) {
    blueAmplified = false;
    setTimeout(function() {
      blueAmplifiedDiv.transition({queue: false, opacity: 0}, amplifyFadeTimeMs, "linear", function () {
        $(`#${blueSide}Amplified svg circle`).css("stroke-dashoffset", "-" + amplifyProgressStartOffset);
        blueAmplifiedDiv.hide();
        blueLightsDiv.show();
        blueLightsDiv.transition({queue: false, opacity: 1}, amplifyFadeTimeMs, "linear");
      });
    }, amplifyDwellTimeMs);
  }

  $(`#${redSide}Lights .amp-low`).attr("data-lit", data.Red.Score.AmpSpeaker.BankedAmpNotes >= 1);
  $(`#${redSide}Lights .amp-high`).attr("data-lit", data.Red.Score.AmpSpeaker.BankedAmpNotes >= 2);
  $(`#${redSide}Lights .amp-coop`).attr("data-lit", data.Red.Score.AmpSpeaker.CoopActivated);
  $(`#${redSide}Amplified svg text`).text(data.Red.AmplifiedTimeRemainingSec);
  $(`#${blueSide}Lights .amp-low`).attr("data-lit", data.Blue.Score.AmpSpeaker.BankedAmpNotes >= 1);
  $(`#${blueSide}Lights .amp-high`).attr("data-lit", data.Blue.Score.AmpSpeaker.BankedAmpNotes >= 2);
  $(`#${blueSide}Lights .amp-coop`).attr("data-lit", data.Blue.Score.AmpSpeaker.CoopActivated);
  $(`#${blueSide}Amplified svg text`).text(data.Blue.AmplifiedTimeRemainingSec);
};

const transitionBlankToIntro = function(callback) {
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

const transitionBlankToMatch = function(callback) {
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
    $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
  });
};

const transitionBlankToTimeout = function(callback) {
  $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
  $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
    $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
  });
};

const transitionIntroToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      callback();
    });
  });
};

const transitionIntroToMatch = function(callback) {
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
    $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
  });
};

const transitionIntroToTimeout = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
        $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

const transitionMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".teams").hide();
      $(".score-fields").hide();
      callback();
    });
  });
};

const transitionMatchToIntro = function(callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
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

const transitionTimeoutToBlank = function(callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", callback);
  });
};

const transitionTimeoutToIntro = function(callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
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

const getAvatarUrl = function(teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

$(function() {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");
  const reversed = urlParams.get("reversed");
  if (reversed === "true") {
    redSide = "right";
    blueSide = "left";
  } else {
    redSide = "left";
    blueSide = "right";
  }
  $(".reversible-left").attr("data-reversed", reversed);
  $(".reversible-right").attr("data-reversed", reversed);

  // Adjust position and size of display contents.
  const overlayCentering = $("#overlayCentering");
  overlayCentering.css("top", parseInt(urlParams.get("topSpacingPx")) + overlayTopOffset + "px");
  overlayCentering.css("transform", `scale(${urlParams.get("zoomFactor")})`);

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
