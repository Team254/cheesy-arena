// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared client-side logic for interpreting match state and timing notifications.

// MatchType enum values.
const matchTypeTest = 0;
const matchTypePractice = 1;
const matchTypeQualification = 2;
const matchTypePlayoff = 3;

var matchStates = {
  0: "PRE_MATCH",
  1: "START_MATCH",
  2: "WARMUP_PERIOD",
  3: "AUTO_PERIOD",
  4: "PAUSE_PERIOD",
  5: "TELEOP_PERIOD",
  6: "POST_MATCH",
  7: "TIMEOUT_ACTIVE",
  8: "POST_TIMEOUT"
};
var matchTiming;

// Handles a websocket message containing the length of each period in the match.
var handleMatchTiming = function(data) {
  matchTiming = data;
};

// Converts the raw match state and time into a human-readable state and per-period time. Calls the provided
// callback with the result.
var translateMatchTime = function(data, callback) {
  var matchStateText;
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      matchStateText = "PRE-MATCH";
      break;
    case "START_MATCH":
    case "WARMUP_PERIOD":
      matchStateText = "WARMUP";
      break;
    case "AUTO_PERIOD":
      matchStateText = "AUTONOMOUS";
      break;
    case "PAUSE_PERIOD":
      matchStateText = "PAUSE";
      break;
    case "TELEOP_PERIOD":
      matchStateText = "TELEOPERATED";
      break;
    case "POST_MATCH":
      matchStateText = "POST-MATCH";
      break;
    case "TIMEOUT_ACTIVE":
    case "POST_TIMEOUT":
      matchStateText = "TIMEOUT";
      break;
  }
  callback(matchStates[data.MatchState], matchStateText, getCountdown(data.MatchState, data.MatchTimeSec));
};

// Returns the per-period countdown for the given match state and overall time into the match.
var getCountdown = function(matchState, matchTimeSec) {
  switch (matchStates[matchState]) {
    case "PRE_MATCH":
    case "START_MATCH":
    case "WARMUP_PERIOD":
      return matchTiming.AutoDurationSec;
    case "AUTO_PERIOD":
      return matchTiming.WarmupDurationSec + matchTiming.AutoDurationSec - matchTimeSec;
    case "TELEOP_PERIOD":
      return matchTiming.WarmupDurationSec + matchTiming.AutoDurationSec + matchTiming.TeleopDurationSec +
          matchTiming.PauseDurationSec - matchTimeSec;
    case "TIMEOUT_ACTIVE":
      return matchTiming.TimeoutDurationSec - matchTimeSec;
    default:
      return 0;
  }
};
