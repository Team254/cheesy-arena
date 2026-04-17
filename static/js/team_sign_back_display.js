// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the "team_sign_back" display.

var websocket;
var currentMatch;
var currentMatchTimeData = null;
var currentScoreData = null;
var displayAlliance = "Red"; // Which alliance's data to show, set from URL param
var showInactive = false; // If true, show inactive fuel count instead of auto climb points

// Game timing constants (must match game/match_timing.go)
const transitionDurationSec = 10; // First 10 seconds of teleop when both hubs are active
const shiftDurationSec = 25;
const endGameDurationSec = 30; // Last 30 seconds of teleop when both hubs are active

// RP thresholds (must match game/score.go)
const energizedRPThreshold = 100;
const superchargedRPThreshold = 360;

// Handles a websocket message to change which screen is displayed.
const handleAudienceDisplayMode = function (targetScreen) {
  // For this display, we mostly care about match state which comes from matchTime.
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  currentMatch = data.Match;
  updateDisplay();
};

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  currentMatchTimeData = data;
  updateDisplay();
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data) {
  currentScoreData = data;
  updateDisplay();
};

// Determines if red won auto based on current score data.
const didRedWinAuto = function () {
  if (!currentScoreData) {
    return false;
  }
  const redAutoPoints = currentScoreData.Red.ScoreSummary.AutoPoints;
  const blueAutoPoints = currentScoreData.Blue.ScoreSummary.AutoPoints;
  if (redAutoPoints > blueAutoPoints) {
    return true;
  }
  if (redAutoPoints === blueAutoPoints && currentScoreData.AutoTieWinner === "red") {
    return true;
  }
  return false;
};

// Returns shift info: { shiftChar, timeLeft, activeAlliance }
// shiftChar: A=Auto, T=Transition, R=Red active, B=Blue active, E=End game
// activeAlliance: "Red", "Blue", or "Both"
const getShiftInfo = function () {
  if (!matchTiming || !currentMatchTimeData) {
    return { shiftChar: "P", timeLeft: 0, activeAlliance: "Both" };
  }

  const matchTimeSec = currentMatchTimeData.MatchTimeSec;
  const matchState = currentMatchTimeData.MatchState;

  // Calculate key time boundaries using matchTiming from websocket
  const warmupDurationSec = matchTiming.WarmupDurationSec;
  const autoDurationSec = matchTiming.AutoDurationSec;
  const pauseDurationSec = matchTiming.PauseDurationSec;
  const teleopDurationSec = matchTiming.TeleopDurationSec;

  const autoEndSec = warmupDurationSec + autoDurationSec;
  const teleopStartSec = autoEndSec + pauseDurationSec;
  const transitionEndSec = teleopStartSec + transitionDurationSec;
  const teleopEndSec = teleopStartSec + teleopDurationSec;
  const endGameStartSec = teleopEndSec - endGameDurationSec;

  // PRE_MATCH, START_MATCH, WARMUP_PERIOD
  if (matchState === 0 || matchState === 1 || matchState === 2) {
    return { shiftChar: "P", timeLeft: 0, activeAlliance: "Both" };
  }

  // AUTO_PERIOD - both hubs active
  if (matchState === 3) {
    const timeLeft = Math.max(0, Math.ceil(autoEndSec - matchTimeSec));
    return { shiftChar: "A", timeLeft: timeLeft, activeAlliance: "Both" };
  }

  // PAUSE_PERIOD - both hubs active, show T for transition coming
  if (matchState === 4) {
    const timeLeft = Math.max(0, Math.ceil(transitionEndSec - matchTimeSec));
    return { shiftChar: "T", timeLeft: timeLeft, activeAlliance: "Both" };
  }

  // TELEOP_PERIOD
  if (matchState === 5) {
    // Transition period (first 10 seconds of teleop) - both hubs active
    if (matchTimeSec < transitionEndSec) {
      const timeLeft = Math.max(0, Math.ceil(transitionEndSec - matchTimeSec));
      return { shiftChar: "T", timeLeft: timeLeft, activeAlliance: "Both" };
    }

    // End game (last 30 seconds) - both hubs active
    if (matchTimeSec >= endGameStartSec) {
      const timeLeft = Math.max(0, Math.ceil(teleopEndSec - matchTimeSec));
      return { shiftChar: "E", timeLeft: timeLeft, activeAlliance: "Both" };
    }

    // Alternating shifts between transition end and end game start
    const postTransitionSec = matchTimeSec - transitionEndSec;
    const shift = Math.floor(postTransitionSec / shiftDurationSec);
    const timeInShift = postTransitionSec - (shift * shiftDurationSec);
    const timeLeft = Math.max(0, Math.ceil(shiftDurationSec - timeInShift));

    // Determine which hub is active
    // The alliance that LOST auto has their hub active first
    const redWonAuto = didRedWinAuto();
    let redHubActive;
    if (redWonAuto) {
      // Red won auto, so Red hub is INACTIVE first (even shifts), ACTIVE on odd shifts
      redHubActive = (shift % 2 === 1);
    } else {
      // Blue won auto or tie, so Red hub is ACTIVE first (even shifts), INACTIVE on odd shifts
      redHubActive = (shift % 2 === 0);
    }

    if (redHubActive) {
      return { shiftChar: "R", timeLeft: timeLeft, activeAlliance: "Red" };
    } else {
      return { shiftChar: "B", timeLeft: timeLeft, activeAlliance: "Blue" };
    }
  }

  // POST_MATCH
  if (matchState === 6) {
    return { shiftChar: "E", timeLeft: 0, activeAlliance: "Both" };
  }

  return { shiftChar: "P", timeLeft: 0, activeAlliance: "Both" };
};

const updateDisplay = function () {
  if (!currentMatch) {
    $("#displayText").text("WAITING");
    return;
  }

  // Only show during Test, Practice and Qualification matches (not Playoff)
  if (currentMatch.Type === matchTypePlayoff) {
    $("#displayText").text("QUAL ONLY");
    return;
  }

  if (!currentMatchTimeData || !matchTiming) {
    $("#displayText").text("P00 0/100 0 0:00");
    return;
  }

  // 1. Current SHIFT and time remaining in that period
  const shiftInfo = getShiftInfo();
  const timeLeftStr = shiftInfo.timeLeft < 10 ? "0" + shiftInfo.timeLeft : String(shiftInfo.timeLeft);
  const shiftDisplay = shiftInfo.shiftChar + timeLeftStr;

  // 2. Progress towards the FUEL Ranking Points
  // Always show the alliance specified by the URL param
  // Only AutoFuel + ActiveFuel count towards RP (not InactiveFuel)
  let fuelForRP = 0;
  let fuelThreshold = energizedRPThreshold;
  let autoTowerPoints = 0;

  let thirdFieldValue = 0;

  if (currentScoreData) {
    const score = currentScoreData[displayAlliance].Score;
    fuelForRP = score.AutoFuel + score.ActiveFuel;
    // Once we pass energized threshold, show progress towards supercharged
    if (fuelForRP >= energizedRPThreshold) {
      fuelThreshold = superchargedRPThreshold;
    }

    if (showInactive) {
      // Show inactive fuel count
      thirdFieldValue = score.InactiveFuel;
    } else {
      // Show auto climb points
      thirdFieldValue = currentScoreData[displayAlliance].ScoreSummary.AutoClimbPoints;
    }
  }

  const rpProgress = fuelForRP + "/" + fuelThreshold;

  // 3. Either AUTO TOWER points or inactive fuel count depending on show_inactive param
  const thirdFieldDisplay = String(thirdFieldValue);

  // 4. Remaining MATCH period time
  let matchTimeRemaining = "0:00";
  translateMatchTime(currentMatchTimeData, function (state, stateText, countdownSec) {
    matchTimeRemaining = getCountdownString(countdownSec);
  });

  const fullText = shiftDisplay + " " + rpProgress + " " + thirdFieldDisplay + " " + matchTimeRemaining;
  $("#displayText").text(fullText);
};

$(function () {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");

  // Set which alliance's data to display (default to Red)
  const allianceParam = urlParams.get("alliance");
  if (allianceParam && allianceParam.toLowerCase() === "blue") {
    displayAlliance = "Blue";
    $("#displayText").css("color", "#00f"); // Blue text
  } else {
    displayAlliance = "Red";
    $("#displayText").css("color", "#f00"); // Red text
  }

  // Set whether to show inactive fuel instead of auto climb points
  const showInactiveParam = urlParams.get("show_inactive");
  showInactive = (showInactiveParam === "true");

  updateDisplay();

  websocket = new CheesyWebsocket("/displays/team_sign_back/websocket", {
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
    }
  });
});
