// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
let transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
let currentScreen = "blank";
let redSide;
let blueSide;
let currentMatch;
let overlayCenteringHideParams;
let overlayCenteringShowParams;
const allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());
const sponsorImageTemplate = Handlebars.compile($("#sponsorImageTemplate").html());
const sponsorTextTemplate = Handlebars.compile($("#sponsorTextTemplate").html());

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const overlayCenteringTopUp = "-130px";
const overlayCenteringBottomHideParams = {queue: false, bottom: $("#overlayCentering").css("bottom")};
const overlayCenteringBottomShowParams = {queue: false, bottom: "0px"};
const overlayCenteringTopHideParams = {queue: false, top: overlayCenteringTopUp};
const overlayCenteringTopShowParams = {queue: false, top: "50px"};
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "20px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "185px";
const scoreOut = "370px";
const scoreFieldsOut = "150px";
const scoreLogoTop = "-530px";
const bracketLogoTop = "-780px";
const bracketLogoScale = 0.75;
const timeoutDetailsIn = $("#timeoutDetails").css("width");
const timeoutDetailsOut = "570px";

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function (targetScreen) {
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

    if (targetScreen === "sponsor") {
      initializeSponsorDisplay();
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
    $(`#${redSide}PlayoffAlliance`).text(currentMatch.PlayoffRedAlliance);
    $(`#${blueSide}PlayoffAlliance`).text(currentMatch.PlayoffBlueAlliance);
    $(".playoff-alliance").show();

    // Show the series status if this playoff round isn't just a single match.
    if (data.Matchup.NumWinsToAdvance > 1) {
      $(`#${redSide}PlayoffAllianceWins`).text(data.Matchup.RedAllianceWins);
      $(`#${blueSide}PlayoffAllianceWins`).text(data.Matchup.BlueAllianceWins);
      $("#playoffSeriesStatus").css("display", "flex");
    } else {
      $("#playoffSeriesStatus").hide();
    }
  } else {
    $(`#${redSide}PlayoffAlliance`).text("");
    $(`#${blueSide}PlayoffAlliance`).text("");
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

  // Determine who won auto
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

  // Calculate hub activation status
  const redHubActive = isRedHubActive(matchTimeSec, matchState, redWonAuto);
  const blueHubActive = isBlueHubActive(matchTimeSec, matchState, blueWonAuto);

  // Determine if we should flash (last 3 seconds of a shift or match)
  const shouldFlash = shouldHubFlash(matchTimeSec, matchState);

  // Update red hub indicator
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

  // Update blue hub indicator
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

// Helper function to determine if red hub is active (mirrors game/match_timing.go logic)
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

  // During END GAME (last 30 seconds), both hubs are active
  if (matchTimeSec >= teleopEndSec - endGameDurationSec && matchTimeSec < teleopEndSec) {
    return true;
  }

  // After the match ends, hubs are not active
  if (matchTimeSec >= teleopEndSec) {
    return false;
  }

  // During teleop alternating shifts (after transition period)
  if (matchTimeSec < teleopStartSec) {
    return false;
  }

  // Calculate which alternating shift we're in (after transition period)
  const postTransitionSec = matchTimeSec - (teleopStartSec + transitionDurationSec);
  if (postTransitionSec < 0) {
    return false;
  }
  const shift = Math.floor(postTransitionSec / 25); // 25 second shifts

  if (redWonAuto) {
    // Red won auto, so Red is INACTIVE first, then alternates
    // Red is INACTIVE on even shifts (0, 2, 4...), ACTIVE on odd shifts (1, 3, 5...)
    return shift % 2 === 1;
  } else {
    // Blue won auto, so Red is ACTIVE first, then alternates
    // Red is ACTIVE on even shifts (0, 2, 4...), INACTIVE on odd shifts (1, 3, 5...)
    return shift % 2 === 0;
  }
};

// Helper function to determine if blue hub is active (mirrors game/match_timing.go logic)
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

  // During END GAME (last 30 seconds), both hubs are active
  if (matchTimeSec >= teleopEndSec - endGameDurationSec && matchTimeSec < teleopEndSec) {
    return true;
  }

  // After the match ends, hubs are not active
  if (matchTimeSec >= teleopEndSec) {
    return false;
  }

  // During teleop alternating shifts (after transition period)
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
    // Blue is INACTIVE on even shifts (0, 2, 4...), ACTIVE on odd shifts (1, 3, 5...)
    return shift % 2 === 1;
  } else {
    // Red won auto, so Blue is ACTIVE first, then alternates
    // Blue is ACTIVE on even shifts (0, 2, 4...), INACTIVE on odd shifts (1, 3, 5...)
    return shift % 2 === 0;
  }
};

// Helper function to determine if hub indicators should flash
const shouldHubFlash = function(matchTimeSec, matchState) {
  const teleopStartSec = 23; // warmup(0) + auto(20) + pause(3)
  const transitionDurationSec = 10;
  const transitionEndSec = teleopStartSec + transitionDurationSec;
  const teleopEndSec = 163; // teleopStartSec + teleop(140)
  const shiftDurationSec = 25;
  const flashThresholdSec = 3;

  // Flash during last 3 seconds of match
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

// Handles a websocket message to populate the final score data.
const handleScorePosted = function (data) {
  $(`#${redSide}FinalScore`).text(data.RedScoreSummary.Score);
  $(`#${redSide}FinalAlliance`).text("Alliance " + data.Match.PlayoffRedAlliance);
  setTeamInfo(redSide, 1, data.Match.Red1, data.RedCards, data.RedRankings);
  setTeamInfo(redSide, 2, data.Match.Red2, data.RedCards, data.RedRankings);
  setTeamInfo(redSide, 3, data.Match.Red3, data.RedCards, data.RedRankings);
  if (data.RedOffFieldTeamIds.length > 0) {
    setTeamInfo(redSide, 4, data.RedOffFieldTeamIds[0], data.RedCards, data.RedRankings);
  } else {
    setTeamInfo(redSide, 4, 0, data.RedCards, data.RedRankings);
  }
  $(`#${redSide}FinalAutoFuelPoints`).text(data.RedScoreSummary.AutoFuelPoints);
  $(`#${redSide}FinalAutoClimbPoints`).text(data.RedScoreSummary.AutoClimbPoints);
  $(`#${redSide}FinalActiveFuelPoints`).text(data.RedScoreSummary.ActiveFuelPoints);
  $(`#${redSide}FinalTeleopClimbPoints`).text(data.RedScoreSummary.TeleopClimbPoints);
  $(`#${redSide}FinalFoulPoints`).text(data.RedScoreSummary.FoulPoints);
  $(`#${redSide}FinalEnergizedRankingPoint`).html(
    data.RedScoreSummary.EnergizedRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${redSide}FinalEnergizedRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.EnergizedRankingPoint
  );
  $(`#${redSide}FinalSuperchargedRankingPoint`).html(
    data.RedScoreSummary.SuperchargedRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${redSide}FinalSuperchargedRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.SuperchargedRankingPoint
  );
  $(`#${redSide}FinalTraversalRankingPoint`).html(
    data.RedScoreSummary.TraversalRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${redSide}FinalTraversalRankingPoint`).attr(
    "data-checked", data.RedScoreSummary.TraversalRankingPoint
  );
  $(`#${redSide}FinalRankingPoints`).html(data.RedRankingPoints);
  $(`#${redSide}FinalWins`).text(data.RedWins);
  const redFinalDestination = $(`#${redSide}FinalDestination`);
  redFinalDestination.html(data.RedDestination.replace("Advances to ", "Advances to<br>"));
  redFinalDestination.toggle(data.RedDestination !== "");
  redFinalDestination.attr("data-won", data.RedWon);

  $(`#${blueSide}FinalScore`).text(data.BlueScoreSummary.Score);
  $(`#${blueSide}FinalAlliance`).text("Alliance " + data.Match.PlayoffBlueAlliance);
  setTeamInfo(blueSide, 1, data.Match.Blue1, data.BlueCards, data.BlueRankings);
  setTeamInfo(blueSide, 2, data.Match.Blue2, data.BlueCards, data.BlueRankings);
  setTeamInfo(blueSide, 3, data.Match.Blue3, data.BlueCards, data.BlueRankings);
  if (data.BlueOffFieldTeamIds.length > 0) {
    setTeamInfo(blueSide, 4, data.BlueOffFieldTeamIds[0], data.BlueCards, data.BlueRankings);
  } else {
    setTeamInfo(blueSide, 4, 0, data.BlueCards, data.BlueRankings);
  }
  $(`#${blueSide}FinalAutoFuelPoints`).text(data.BlueScoreSummary.AutoFuelPoints);
  $(`#${blueSide}FinalAutoClimbPoints`).text(data.BlueScoreSummary.AutoClimbPoints);
  $(`#${blueSide}FinalActiveFuelPoints`).text(data.BlueScoreSummary.ActiveFuelPoints);
  $(`#${blueSide}FinalTeleopClimbPoints`).text(data.BlueScoreSummary.TeleopClimbPoints);
  $(`#${blueSide}FinalFoulPoints`).text(data.BlueScoreSummary.FoulPoints);
  $(`#${blueSide}FinalEnergizedRankingPoint`).html(
    data.BlueScoreSummary.EnergizedRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${blueSide}FinalEnergizedRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.EnergizedRankingPoint
  );
  $(`#${blueSide}FinalSuperchargedRankingPoint`).html(
    data.BlueScoreSummary.SuperchargedRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${blueSide}FinalSuperchargedRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.SuperchargedRankingPoint
  );
  $(`#${blueSide}FinalTraversalRankingPoint`).html(
    data.BlueScoreSummary.TraversalRankingPoint ? "&#x2714;" : "&#x2718;"
  );
  $(`#${blueSide}FinalTraversalRankingPoint`).attr(
    "data-checked", data.BlueScoreSummary.TraversalRankingPoint
  );
  $(`#${blueSide}FinalRankingPoints`).html(data.BlueRankingPoints);
  $(`#${blueSide}FinalWins`).text(data.BlueWins);
  const blueFinalDestination = $(`#${blueSide}FinalDestination`);
  blueFinalDestination.html(data.BlueDestination.replace("Advances to ", "Advances to<br>"));
  blueFinalDestination.toggle(data.BlueDestination !== "");
  blueFinalDestination.attr("data-won", data.BlueWon);

  let matchName = data.Match.LongName;
  if (data.Match.NameDetail !== "") {
    matchName += " &ndash; " + data.Match.NameDetail;
  }
  $("#finalMatchName").html(matchName);

  // Reload the bracket to reflect any changes.
  $("#bracketSvg").attr("src", "/api/bracket/svg?activeMatch=saved&v=" + new Date().getTime());

  if (data.Match.Type === matchTypePlayoff) {
    // Hide bonus ranking points and show playoff-only fields.
    $(".playoff-hidden-field").hide();
    $(".playoff-only-field").show();
  } else {
    $(".playoff-hidden-field").show();
    $(".playoff-only-field").hide();
  }
  $(".coopertition-hidden-field").toggle(data.CoopertitionEnabled);
};

// Handles a websocket message to play a sound to signal match start/stop/etc.
const handlePlaySound = function (sound) {
  $("audio").each(function (k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#sound-" + sound)[0].play();
};

// Handles a websocket message to update the alliance selection screen.
const handleAllianceSelection = function (data) {
  const alliances = data.Alliances;
  const rankedTeams = data.RankedTeams;
  if (alliances && alliances.length > 0) {
    const numColumns = alliances[0].TeamIds.length + 1;
    $.each(alliances, function (k, v) {
      v.Index = k + 1;
    });
    $("#allianceSelection").html(allianceSelectionTemplate({alliances: alliances, numColumns: numColumns}));
  }
  if (rankedTeams) {
    let text = "";
    $.each(rankedTeams, function (i, v) {
      if (!v.Picked) {
        text += `<div class="unpicked"><div class="unpicked-rank">${v.Rank}.</div>` +
          `<div class="unpicked-team">${v.TeamId}</div></div>`;
      }
    });
    $("#allianceRankings").html(text);
  }

  if (data.ShowTimer) {
    $("#allianceSelectionTimer").text(getCountdownString(data.TimeRemainingSec));
  } else {
    $("#allianceSelectionTimer").html("&nbsp;");
  }
};

// Handles a websocket message to populate and/or show/hide a lower third.
const handleLowerThird = function (data) {
  if (data.LowerThird !== null) {
    if (data.LowerThird.BottomText === "") {
      $("#lowerThirdTop").hide();
      $("#lowerThirdBottom").hide();
      $("#lowerThirdSingle").text(data.LowerThird.TopText);
      $("#lowerThirdSingle").show();
    } else {
      $("#lowerThirdSingle").hide();
      $("#lowerThirdTop").text(data.LowerThird.TopText);
      $("#lowerThirdBottom").text(data.LowerThird.BottomText);
      $("#lowerThirdTop").show();
      $("#lowerThirdBottom").show();
    }
  }

  const lowerThirdElement = $("#lowerThird");
  if (data.ShowLowerThird && !lowerThirdElement.is(":visible")) {
    lowerThirdElement.show();
    lowerThirdElement.transition({queue: false, left: "150px"}, 750, "ease");
  } else if (!data.ShowLowerThird && lowerThirdElement.is(":visible")) {
    lowerThirdElement.transition({queue: false, left: "-1000px"}, 1000, "ease", function () {
      lowerThirdElement.hide();
    });
  }
};

const transitionAllianceSelectionToBlank = function (callback) {
  $('#allianceSelectionCentering').transition({queue: false, right: "-60em"}, 500, "ease", callback);
  $('#allianceRankingsCentering.enabled').transition({queue: false, left: "-60em"}, 500, "ease");
};

const transitionBlankToAllianceSelection = function (callback) {
  $('#allianceSelectionCentering').css("right", "-60em").show();
  $('#allianceSelectionCentering').transition({queue: false, right: "3em"}, 500, "ease", callback);
  $('#allianceRankingsCentering.enabled').css("left", "-60em").show();
  $('#allianceRankingsCentering.enabled').transition({queue: false, left: "3em"}, 500, "ease");
};

const transitionBlankToBracket = function (callback) {
  transitionBlankToLogo(function () {
    setTimeout(function () {
      transitionLogoToBracket(callback);
    }, 50);
  });
};

const transitionBlankToIntro = function (callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
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
  $(".blindsCenter.blank").css({rotateY: "0deg"});
  $(".blindsCenter.full").css({rotateY: "-180deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function () {
      $(".blindsCenter.blank").transition({queue: false, rotateY: "180deg"}, 500, "ease");
      $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 500, "ease", callback);
    }, 200);
  });
};

const transitionBlankToLogoLuma = function (callback) {
  $(".blindsCenter.blank").css({rotateY: "180deg"});
  $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 1000, "ease", callback);
};

const transitionBlankToMatch = function (callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
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
    });
  });
};

const transitionBlankToScore = function (callback) {
  transitionBlankToLogo(function () {
    setTimeout(function () {
      transitionLogoToScore(callback);
    }, 50);
  });
};

const transitionBlankToSponsor = function (callback) {
  $(".blindsCenter.blank").css({rotateY: "90deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function () {
      $("#sponsor").show();
      $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
    }, 200);
  });
};

const transitionBlankToTimeout = function (callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function () {
      $(".timeout-detail").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  });
};

const transitionBracketToBlank = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToBlank(callback);
  });
};

const transitionBracketToLogo = function (callback) {
  $("#bracket").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#bracket").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0, scale: 1}, 625, "ease", callback);
};

const transitionBracketToLogoLuma = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToLogoLuma(callback);
  });
};

const transitionBracketToScore = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop, scale: 1}, 1000, "ease");
  $("#bracket").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $("#bracket").hide();
    $("#finalScore").show();
    $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionBracketToSponsor = function (callback) {
  transitionBracketToLogo(function () {
    transitionLogoToSponsor(callback);
  });
};

const transitionIntroToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function () {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
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
  $(".blindsCenter.blank").transition({queue: false, rotateY: "360deg"}, 500, "ease");
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 500, "ease", function () {
    setTimeout(function () {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
    }, 200);
  });
};

const transitionLogoToBracket = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 625, "ease");
  $("#bracket").show();
  $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

const transitionLogoToLogoLuma = function (callback) {
  $(".blinds.left").removeClass("full");
  $(".blinds.right").show();
  $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", function () {
    if (callback) {
      callback();
    }
  });
};

const transitionLogoToScore = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop}, 625, "ease");
  $("#finalScore").show();
  $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

const transitionLogoToSponsor = function (callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "90deg"}, 750, "ease", function () {
    $("#sponsor").show();
    $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionLogoLumaToBlank = function (callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 1000, "ease", callback);
};

const transitionLogoLumaToBracket = function (callback) {
  transitionLogoLumaToLogo(function () {
    transitionLogoToBracket(callback);
  });
};

const transitionLogoLumaToLogo = function (callback) {
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function () {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    if (callback) {
      callback();
    }
  });
};

const transitionLogoLumaToScore = function (callback) {
  transitionLogoLumaToLogo(function () {
    transitionLogoToScore(callback);
  });
};

const transitionMatchToBlank = function (callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function () {
      $(".teams").hide();
      $(".score-fields").hide();
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
    });
  });
};

const transitionMatchToIntro = function (callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
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

const transitionScoreToBlank = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToBlank(callback);
  });
};

const transitionScoreToBracket = function (callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 1000, "ease");
  $("#finalScore").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $("#finalScore").hide();
    $("#bracket").show();
    $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

const transitionScoreToLogo = function (callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "ease", function () {
    $("#finalScore").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0}, 625, "ease", callback);
};

const transitionScoreToLogoLuma = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToLogoLuma(callback);
  });
};

const transitionScoreToSponsor = function (callback) {
  transitionScoreToLogo(function () {
    transitionLogoToSponsor(callback);
  });
};

const transitionSponsorToBlank = function (callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    setTimeout(function () {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
      $("#sponsor").hide();
    }, 200);
  });
};

const transitionSponsorToBracket = function (callback) {
  transitionSponsorToLogo(function () {
    transitionLogoToBracket(callback);
  });
};

const transitionSponsorToLogo = function (callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function () {
    $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 750, "ease", callback);
    $("#sponsor").hide();
  });
};

const transitionSponsorToScore = function (callback) {
  transitionSponsorToLogo(function () {
    transitionLogoToScore(callback);
  });
};

const transitionTimeoutToBlank = function (callback) {
  $(".timeout-detail").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function () {
    $("#timeoutDetails").transition({queue: false, width: timeoutDetailsIn}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function () {
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
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

// Loads sponsor slide data and builds the slideshow HTML.
const initializeSponsorDisplay = function () {
  $.getJSON("/api/sponsor_slides", function (slides) {
    $("#sponsorContainer").empty();

    // Inject the HTML for each slide into the DOM.
    $.each(slides, function (index, slide) {
      slide.DisplayTimeMs = slide.DisplayTimeSec * 1000;
      slide.First = index === 0;

      let slideHtml;
      if (slide.Image) {
        slideHtml = sponsorImageTemplate(slide);
      } else {
        slideHtml = sponsorTextTemplate(slide);
      }
      $("#sponsorContainer").append(slideHtml);
    });
  });
};

const getAvatarUrl = function (teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

const setTeamInfo = function (side, position, teamId, cards, rankings) {
  const teamNumberElement = $(`#${side}FinalTeam${position}`);
  teamNumberElement.html(teamId);
  teamNumberElement.toggle(teamId > 0);
  const avatarElement = $(`#${side}FinalTeam${position}Avatar`);
  avatarElement.attr("src", getAvatarUrl(teamId));
  avatarElement.toggle(teamId > 0);

  const cardElement = $(`#${side}FinalTeam${position}Card`);
  cardElement.attr("data-card", cards[teamId.toString()] || "");

  const ranking = rankings[teamId];
  let rankIndicator = "";
  let rankNumber = "";
  if (ranking !== undefined && ranking !== null && ranking.Rank !== 0) {
    rankNumber = ranking.Rank;
    if (rankNumber > ranking.PreviousRank && ranking.PreviousRank > 0) {
      rankIndicator = "rank-down";
    } else if (rankNumber < ranking.PreviousRank) {
      rankIndicator = "rank-up";
    }
  }

  const rankIndicatorElement = $(`#${side}FinalTeam${position}RankIndicator`);
  rankIndicatorElement.attr("src", rankIndicator === "" ? "" : `/static/img/${rankIndicator}.svg`);
  rankIndicatorElement.toggle(rankIndicator !== "" && teamId > 0);

  const rankNumberElement = $(`#${side}FinalTeam${position}RankNumber`);
  rankNumberElement.text(rankNumber);
  rankNumberElement.toggle(teamId > 0);
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
  if (urlParams.get("overlayLocation") === "top") {
    overlayCenteringHideParams = overlayCenteringTopHideParams;
    overlayCenteringShowParams = overlayCenteringTopShowParams;
    $("#overlayCentering").css("top", overlayCenteringTopUp);
  } else {
    overlayCenteringHideParams = overlayCenteringBottomHideParams;
    overlayCenteringShowParams = overlayCenteringBottomShowParams;
  }

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/audience/websocket", {
    allianceSelection: function (event) {
      handleAllianceSelection(event.data);
    },
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
    lowerThird: function (event) {
      handleLowerThird(event.data);
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
    playSound: function (event) {
      handlePlaySound(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scorePosted: function (event) {
      handleScorePosted(event.data);
    },
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    allianceSelection: {
      blank: transitionAllianceSelectionToBlank,
    },
    blank: {
      allianceSelection: transitionBlankToAllianceSelection,
      bracket: transitionBlankToBracket,
      intro: transitionBlankToIntro,
      logo: transitionBlankToLogo,
      logoLuma: transitionBlankToLogoLuma,
      match: transitionBlankToMatch,
      score: transitionBlankToScore,
      sponsor: transitionBlankToSponsor,
      timeout: transitionBlankToTimeout,
    },
    bracket: {
      blank: transitionBracketToBlank,
      logo: transitionBracketToLogo,
      logoLuma: transitionBracketToLogoLuma,
      score: transitionBracketToScore,
      sponsor: transitionBracketToSponsor,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    logo: {
      blank: transitionLogoToBlank,
      bracket: transitionLogoToBracket,
      logoLuma: transitionLogoToLogoLuma,
      score: transitionLogoToScore,
      sponsor: transitionLogoToSponsor,
    },
    logoLuma: {
      blank: transitionLogoLumaToBlank,
      bracket: transitionLogoLumaToBracket,
      logo: transitionLogoLumaToLogo,
      score: transitionLogoLumaToScore,
    },
    match: {
      blank: transitionMatchToBlank,
      intro: transitionMatchToIntro,
    },
    score: {
      blank: transitionScoreToBlank,
      bracket: transitionScoreToBracket,
      logo: transitionScoreToLogo,
      logoLuma: transitionScoreToLogoLuma,
      sponsor: transitionScoreToSponsor,
    },
    sponsor: {
      blank: transitionSponsorToBlank,
      bracket: transitionSponsorToBracket,
      logo: transitionSponsorToLogo,
      score: transitionSponsorToScore,
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro,
    },
  }
});
