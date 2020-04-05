// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
var currentMatchId;
var scoreIsReady;
var lowBatteryThreshold = 8;

// Sends a websocket message to load a team into an alliance station.
var substituteTeam = function(team, position) {
  websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

// Sends a websocket message to toggle the bypass status for an alliance station.
var toggleBypass = function(station) {
  websocket.send("toggleBypass", station);
};

// Sends a websocket message to start the match.
var startMatch = function() {
  websocket.send("startMatch",
      { muteMatchSounds: $("#muteMatchSounds").prop("checked") });
};

// Sends a websocket message to abort the match.
var abortMatch = function() {
  websocket.send("abortMatch");
};

// Sends a websocket message to commit the match score and load the next match.
var commitResults = function() {
  websocket.send("commitResults");
};

// Sends a websocket message to discard the match score and load the next match.
var discardResults = function() {
  websocket.send("discardResults");
};

// Sends a websocket message to change what the audience display is showing.
var setAudienceDisplay = function() {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

// Sends a websocket message to change what the alliance station display is showing.
var setAllianceStationDisplay = function() {
  websocket.send("setAllianceStationDisplay", $("input[name=allianceStationDisplay]:checked").val());
};

// Sends a websocket message to start the timeout.
var startTimeout = function() {
  var duration = $("#timeoutDuration").val().split(":");
  var durationSec = parseFloat(duration[0]);
  if (duration.length > 1) {
    durationSec = durationSec * 60 + parseFloat(duration[1]);
  }
  websocket.send("startTimeout", durationSec);
};

var confirmCommit = function(isReplay) {
  if (isReplay || !scoreIsReady) {
    // Show the appropriate message(s) in the confirmation dialog.
    $("#confirmCommitReplay").css("display", isReplay ? "block" : "none");
    $("#confirmCommitNotReady").css("display", scoreIsReady ? "none" : "block");
    $("#confirmCommitResults").modal("show");
  } else {
    commitResults();
  }
};

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function(data) {
  // If getting data for the wrong match (e.g. after a server restart), reload the page.
  if (currentMatchId == null) {
    currentMatchId = data.MatchId;
  } else if (currentMatchId !== data.MatchId) {
    location.reload();
  }

  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    var wifiStatus = data.TeamWifiStatuses[station];
    $("#status" + station + " .radio-status").text(wifiStatus.TeamId);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      var dsConn = stationStatus.DsConn;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsConn.DsLinked);

      // Format the radio status box according to the connection status of the robot radio.
      var radioOkay = stationStatus.Team && stationStatus.Team.Id === wifiStatus.TeamId && wifiStatus.RadioLinked;
      $("#status" + station + " .radio-status").attr("data-status-ok", radioOkay);

      // Format the robot status box.
      var robotOkay = dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked;
      $("#status" + station + " .robot-status").attr("data-status-ok", robotOkay);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        $("#status" + station + " .robot-status").text(dsConn.BatteryVoltage.toFixed(1) + "V");
      }
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");

      // Format the robot status box according to whether the AP is configured with the correct SSID.
      var expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
      if (wifiStatus.TeamId === expectedTeamId) {
        if (wifiStatus.RadioLinked) {
          $("#status" + station + " .radio-status").attr("data-status-ok", true);
        } else {
          $("#status" + station + " .radio-status").attr("data-status-ok", "");
        }
      } else {
        $("#status" + station + " .radio-status").attr("data-status-ok", false);
      }
    }

    if (stationStatus.Estop) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("ES");
    } else if (stationStatus.Bypass) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("B");
    } else {
      $("#status" + station + " .bypass-status").attr("data-status-ok", true);
      $("#status" + station + " .bypass-status").text("");
    }
  });

  // Enable/disable the buttons based on the current match state.
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      $("#startMatch").prop("disabled", !data.CanStartMatch);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", false);
      break;
    case "START_MATCH":
    case "WARMUP_PERIOD":
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
    case "POST_MATCH":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", false);
      $("#discardResults").prop("disabled", false);
      $("#editResults").prop("disabled", false);
      $("#startTimeout").prop("disabled", true);
      break;
    case "TIMEOUT_ACTIVE":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
    case "POST_TIMEOUT":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
  }

  if (data.PlcIsHealthy) {
    $("#plcStatus").text("Connected");
    $("#plcStatus").attr("data-ready", true);
  } else {
    $("#plcStatus").text("Not Connected");
    $("#plcStatus").attr("data-ready", false);
  }
  $("#fieldEstop").attr("data-ready", !data.FieldEstop);
  $.each(data.PlcArmorBlockStatuses, function(name, status) {
    $("#plc" + name + "Status").attr("data-ready", status);
  });
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.Red.ScoreSummary.Score);
  $("#blueScore").text(data.Blue.ScoreSummary.Score);
};

// Handles a websocket message to update the audience display screen selector.
var handleAudienceDisplayMode = function(data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to signal whether the referee and scorers have committed after the match.
var handleScoringStatus = function(data) {
  scoreIsReady = data.RefereeScoreReady && data.RedScoreReady && data.BlueScoreReady;
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  $("#redScoreStatus").text("Red Scoring " + data.NumRedScoringPanelsReady + "/" + data.NumRedScoringPanels);
  $("#redScoreStatus").attr("data-ready", data.RedScoreReady);
  $("#blueScoreStatus").text("Blue Scoring " + data.NumBlueScoringPanelsReady + "/" + data.NumBlueScoringPanels);
  $("#blueScoreStatus").attr("data-ready", data.BlueScoreReady);
};

// Handles a websocket message to update the alliance station display screen selector.
var handleAllianceStationDisplayMode = function(data) {
  $("input[name=allianceStationDisplay]:checked").prop("checked", false);
  $("input[name=allianceStationDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to update the event status message.
var handleEventStatus = function(data) {
  if (data.CycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
  }
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

$(function() {
  // Activate tooltips above the status headers.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/match_play/websocket", {
    allianceStationDisplayMode: function(event) { handleAllianceStationDisplayMode(event.data); },
    arenaStatus: function(event) { handleArenaStatus(event.data); },
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    eventStatus: function(event) { handleEventStatus(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scoringStatus: function(event) { handleScoringStatus(event.data); },
  });
});
