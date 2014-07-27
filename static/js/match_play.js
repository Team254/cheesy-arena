// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;

var substituteTeam = function(team, position) {
  websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

var toggleBypass = function(station) {
  websocket.send("toggleBypass", station);
};

var startMatch = function() {
  websocket.send("startMatch");
};

var abortMatch = function() {
  websocket.send("abortMatch");
};

var commitResults = function() {
  websocket.send("commitResults");
};

var discardResults = function() {
  websocket.send("discardResults");
};

var handleStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    if (stationStatus.DsConn) {
      var dsStatus = stationStatus.DsConn.DriverStationStatus;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsStatus.DsLinked);
      $("#status" + station + " .robot-status").attr("data-status-ok", dsStatus.RobotLinked);
      $("#status" + station + " .battery-status").attr("data-status-ok", dsStatus.BatteryVoltage > 0);
      $("#status" + station + " .battery-status").text(dsStatus.BatteryVoltage.toFixed(1) + "V");
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").text("");
    }

    if (stationStatus.EmergencyStop) {
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
      break;
    case "START_MATCH":
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
    case "ENDGAME_PERIOD":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      break;
    case "POST_MATCH":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#commitResults").prop("disabled", false);
      $("#discardResults").prop("disabled", false);
      break;
  }
};

var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

$(function() {
  // Activate tooltips above the status headers.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/match_play/websocket", {
    status: function(event) { handleStatus(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); }
  });
});
