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
let messageText = "";
let hasMessage = false;

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "20px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "185px";
const scoreOut = "250px";
const scoreFieldsOut = "25px";
const overlayTopOffset = 110;
const timeoutDetailsIn = $("#timeoutDetails").css("width");
const timeoutDetailsOut = "570px";

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function (targetScreen) {
  if (targetScreen === "logoLuma") {
    targetScreen = "logo";
  }
  if (
    targetScreen !== "intro" &&
    targetScreen !== "match" &&
    targetScreen !== "timeout" &&
    targetScreen !== "logo"
  ) {
    targetScreen = "blank";
  }

  transitionQueue.push(targetScreen);
  executeTransitionQueue();
};

// Sequentially executes all transitions in the queue. Returns without doing anything if another invocation is already
// in progress.
const executeTransitionQueue = function () {
  if (transitionInProgress) {
    // There is an existing invocation of this method which will execute all transitions in the queue.
    return;
  }

  if (transitionQueue.length > 0) {
    transitionInProgress = true;
    const targetScreen = transitionQueue.shift();
    const callback = function () {
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
      transitionMap[currentScreen]["blank"](function () {
        transitionMap["blank"][targetScreen](callback);
      });
    }
  }
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
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

// Store current match time data and score data for hub indicator updates
let currentMatchTimeData = null;
let currentScoreData = null;

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  currentMatchTimeData = data;
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchTime").text(getCountdownString(countdownSec));
  });

  // Update hub indicators when time changes (if we have score data)
  if (currentScoreData) {
    updateHubIndicators(currentScoreData);
  }
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data) {
  currentScoreData = data; // Store for hub indicator updates

  $(`#${redSide}ScoreNumber`).text(data.Red.ScoreSummary.Score);
  $(`#${blueSide}ScoreNumber`).text(data.Blue.ScoreSummary.Score);

  // Update FUEL counts (just the number, white color)
  $(`#${redSide}Fuel`).text(data.Red.ScoreSummary.TotalFuel);
  $(`#${blueSide}Fuel`).text(data.Blue.ScoreSummary.TotalFuel);

  // Update hub activation indicators
  updateHubIndicators(data);
};

// Updates the hub activation indicators based on current match state and time
const updateHubIndicators = function(scoreData) {
  if (!currentMatchTimeData) {
    return;
  }

  const matchTimeSec = currentMatchTimeData.MatchTimeSec;
  const matchState = currentMatchTimeData.MatchState;

  const redAutoPoints = scoreData.Red.ScoreSummary.AutoPoints;
  const blueAutoPoints = scoreData.Blue.ScoreSummary.AutoPoints;
  let redWonAuto = redAutoPoints > blueAutoPoints;
  let blueWonAuto = blueAutoPoints > redAutoPoints;

  // Handle tie case - use random tie-breaker from backend
  if (!redWonAuto && !blueWonAuto) {
    if (scoreData.AutoTieWinner === "red") {
      redWonAuto = true;
    } else {
      blueWonAuto = true;
    }
  }

  const redHubActive = isRedHubActive(matchTimeSec, matchState, redWonAuto);
  const blueHubActive = isBlueHubActive(matchTimeSec, matchState, blueWonAuto);
  const shouldFlash = shouldHubFlash(matchTimeSec, matchState);

  const redIndicator = $(`#${redSide}HubIndicator`);
  if (redHubActive) {
    redIndicator.addClass("active");
    if (shouldFlash) {
      redIndicator.addClass("flashing");
    } else {
      redIndicator.removeClass("flashing");
    }
  } else {
    redIndicator.removeClass("active flashing");
  }

  const blueIndicator = $(`#${blueSide}HubIndicator`);
  if (blueHubActive) {
    blueIndicator.addClass("active");
    if (shouldFlash) {
      blueIndicator.addClass("flashing");
    } else {
      blueIndicator.removeClass("flashing");
    }
  } else {
    blueIndicator.removeClass("active flashing");
  }
};

const isRedHubActive = function(matchTimeSec, matchState, redWonAuto) {
  const teleopStartSec = 23; // warmup(0) + auto(20) + pause(3)
  const transitionDurationSec = 10; // First 10 seconds of teleop when both hubs are active
  const teleopEndSec = 163; // teleopStartSec + teleop(140)
  const endGameDurationSec = 30;

  // During auto and pause, both hubs are active
  // matchState: 3 = AUTO_PERIOD, 4 = PAUSE_PERIOD (see match_timing.js)
  if (matchState === 3 || matchState === 4) {
    return true;
  }

  // During transition period (first 10 seconds of teleop), both hubs are active
  if (matchState === 5 && matchTimeSec >= teleopStartSec && matchTimeSec < teleopStartSec + transitionDurationSec) {
    return true;
  }

  if (matchTimeSec >= teleopEndSec - endGameDurationSec && matchTimeSec < teleopEndSec) {
    return true;
  }

  // After the match ends, hubs are not active
  if (matchTimeSec >= teleopEndSec) {
    return false;
  }

  if (matchTimeSec < teleopStartSec) {
    return false;
  }

  // Calculate which alternating shift we're in (after transition period)
  const postTransitionSec = matchTimeSec - (teleopStartSec + transitionDurationSec);
  if (postTransitionSec < 0) {
    return false;
  }
  const shift = Math.floor(postTransitionSec / 25);

  if (redWonAuto) {
    // Red won auto, so Red is INACTIVE first, then alternates
    return shift % 2 === 1;
  } else {
    // Blue won auto, so Red is ACTIVE first, then alternates
    return shift % 2 === 0;
  }
};

const isBlueHubActive = function(matchTimeSec, matchState, blueWonAuto) {
  const teleopStartSec = 23; // warmup(0) + auto(20) + pause(3)
  const transitionDurationSec = 10; // First 10 seconds of teleop when both hubs are active
  const teleopEndSec = 163; // teleopStartSec + teleop(140)
  const endGameDurationSec = 30;

  // During auto and pause, both hubs are active
  // matchState: 3 = AUTO_PERIOD, 4 = PAUSE_PERIOD (see match_timing.js)
  if (matchState === 3 || matchState === 4) {
    return true;
  }

  // During transition period (first 10 seconds of teleop), both hubs are active
  if (matchState === 5 && matchTimeSec >= teleopStartSec && matchTimeSec < teleopStartSec + transitionDurationSec) {
    return true;
  }

  if (matchTimeSec >= teleopEndSec - endGameDurationSec && matchTimeSec < teleopEndSec) {
    return true;
  }

  // After the match ends, hubs are not active
  if (matchTimeSec >= teleopEndSec) {
    return false;
  }

  if (matchTimeSec < teleopStartSec) {
    return false;
  }

  // Calculate which alternating shift we're in (after transition period)
  const postTransitionSec = matchTimeSec - (teleopStartSec + transitionDurationSec);
  if (postTransitionSec < 0) {
    return false;
  }
  const shift = Math.floor(postTransitionSec / 25);

  if (blueWonAuto) {
    // Blue won auto, so Blue is INACTIVE first, then alternates
    return shift % 2 === 1;
  } else {
    // Red won auto, so Blue is ACTIVE first, then alternates
    return shift % 2 === 0;
  }
};

const shouldHubFlash = function(matchTimeSec, matchState) {
  const teleopStartSec = 23; // warmup(0) + auto(20) + pause(3)
  const transitionDurationSec = 10;
  const transitionEndSec = teleopStartSec + transitionDurationSec;
  const teleopEndSec = 163; // teleopStartSec + teleop(140)
  const shiftDurationSec = 25;
  const flashThresholdSec = 3;

  if (matchTimeSec >= teleopEndSec - flashThresholdSec && matchTimeSec < teleopEndSec) {
    return true;
  }

  // Flash during last 3 seconds of transition period
  if (matchTimeSec >= transitionEndSec - flashThresholdSec && matchTimeSec < transitionEndSec) {
    return true;
  }

  // Flash during last 3 seconds of each shift (during teleop, not in END GAME)
  // matchState: 5 = TELEOP_PERIOD (see match_timing.js)
  if (matchState === 5 && matchTimeSec >= transitionEndSec && matchTimeSec < teleopEndSec - 30) {
    const postTransitionSec = matchTimeSec - transitionEndSec;
    const timeInShift = postTransitionSec % shiftDurationSec;
    return timeInShift >= shiftDurationSec - flashThresholdSec;
  }

  return false;
};

const transitionBlankToIntro = function (callback) {
  hideMessage(function () {
    $(".teams").css("display", "flex");
    $(".avatars").css("display", "flex");
    $(".avatars").css("opacity", 1);
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
    });
  });
};

const transitionBlankToLogo = function (callback) {
  showMessage(callback);
}

const transitionBlankToMatch = function (callback) {
  hideMessage(function () {
    $(".teams").css("display", "flex");
    $(".score-fields").css("display", "flex");
    $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
    $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function () {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
    });
  });
};

const transitionBlankToTimeout = function (callback) {
  hideMessage(function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
      $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  });
};

const transitionIntroToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      showMessage(callback);
    });
  });
};

const transitionIntroToMatch = function (callback) {
  $(".avatars").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $(".avatars").hide();
  });
  $(".score-fields").css("display", "flex");
  $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
  $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
  $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function () {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    $(".score-aux").transition({queue: false, opacity: 1}, 750, "ease");
  });
};

const transitionIntroToTimeout = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
        $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

const transitionLogoToBlank = function (callback) {
  showMessage(callback);
}

const transitionMatchToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".teams").hide();
      $(".score-fields").hide();
      showMessage(callback);
    });
  });
};

const transitionMatchToIntro = function (callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-aux").transition({queue: false, opacity: 0}, 750, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
      $(".score-fields").hide();
      $(".avatars").css("display", "flex");
      $(".avatars").transition({queue: false, opacity: 1}, 500, "ease", callback);
    });
  });
};

const transitionTimeoutToBlank = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      showMessage(callback);
    });
  });
};

const transitionTimeoutToIntro = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
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

const showMessage = function (callback) {
  if (!hasMessage) {
    if (callback) {
      callback();
    }
    return;
  }
  $("#message").show();
  $("#message").transition({queue: false, opacity: 1}, 750, "ease", callback);
};

const hideMessage = function (callback) {
  if (!hasMessage) {
    if (callback) {
      callback();
    }
    return;
  }
  $("#message").transition({queue: false, opacity: 0}, 750, "ease", function () {
    $("#message").hide();
    if (callback) {
      callback();
    }
  });
};

const getAvatarUrl = function (teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

$(function () {
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

  messageText = urlParams.get("message") || "";
  hasMessage = messageText !== "";
  const messageDiv = $("#message");
  messageDiv.text(messageText);
  messageDiv.toggle(hasMessage);
  if (hasMessage) {
    showMessage();
  }

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/wall/websocket", {
    allianceSelection: function (event) {
      handleAllianceSelection(event.data);
    },
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    matchTiming: function (event) {
      handleMatchTiming(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    blank: {
      intro: transitionBlankToIntro,
      logo: transitionBlankToLogo,
      match: transitionBlankToMatch,
      timeout: transitionBlankToTimeout,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    logo: {
      blank: transitionLogoToBlank,
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
