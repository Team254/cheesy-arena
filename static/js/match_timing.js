// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared client-side logic for interpreting match state and timing notifications.

var matchStates = {
  0: "PRE_MATCH",
  1: "START_MATCH",
  2: "AUTO_PERIOD",
  3: "PAUSE_PERIOD",
  4: "TELEOP_PERIOD",
  5: "ENDGAME_PERIOD",
  6: "POST_MATCH"
};
var matchTiming;

var handleMatchTiming = function(data) {
  matchTiming = data;
};

var translateMatchTime = function(data, callback) {
  var matchStateText;
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      matchStateText = "PRE-MATCH";
      break;
    case "START_MATCH":
    case "AUTO_PERIOD":
      matchStateText = "AUTONOMOUS";
      break;
    case "PAUSE_PERIOD":
      matchStateText = "PAUSE";
      break;
    case "TELEOP_PERIOD":
    case "ENDGAME_PERIOD":
      matchStateText = "TELEOPERATED";
      break;
    case "POST_MATCH":
      matchStateText = "POST-MATCH";
      break;
  }
  callback(matchStates[data.MatchState], matchStateText, getCountdown(data.MatchState, data.MatchTimeSec));
};

var getCountdown = function(matchState, matchTimeSec) {
  switch (matchStates[matchState]) {
    case "PRE_MATCH":
      return matchTiming.AutoDurationSec;
    case "START_MATCH":
    case "AUTO_PERIOD":
      return matchTiming.AutoDurationSec - matchTimeSec;
    case "TELEOP_PERIOD":
    case "ENDGAME_PERIOD":
      return matchTiming.TeleopDurationSec + matchTiming.AutoDurationSec + matchTiming.PauseDurationSec -
          matchTimeSec;
    default:
      return 0;
  }
};
