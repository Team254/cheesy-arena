// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field monitor display.

var websocket;
var redSide;
var blueSide;
var lowBatteryThreshold = 8;

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function(data) {
  $.each(data.AllianceStations, function(station, stationStatus) {
    // Select the DOM elements corresponding to the team station.
    var teamElementPrefix;
    if (station[0] === "R") {
      teamElementPrefix = "#" + redSide + "Team" + station[1];
    } else {
      teamElementPrefix = "#" + blueSide + "Team" + station[1];
    }
    var teamIdElement = $(teamElementPrefix + "Id");
    var teamDsElement = $(teamElementPrefix + "Ds");
    var teamRadioElement = $(teamElementPrefix + "Radio");
    var teamRadioTextElement = $(teamElementPrefix + "Radio span");
    var teamRobotElement = $(teamElementPrefix + "Robot");
    var teamBypassElement = $(teamElementPrefix + "Bypass");

    if (stationStatus.Team) {
      // Set the team number and status.
      teamIdElement.text(stationStatus.Team.Id);
      var status = "no-link";
      if (stationStatus.Bypass) {
        status = "";
      } else if (stationStatus.DsConn) {
        if (stationStatus.DsConn.RobotLinked) {
          status = "robot-linked";
        } else if (stationStatus.DsConn.RadioLinked) {
          status = "radio-linked";
        } else if (stationStatus.DsConn.DsLinked) {
          status = "ds-linked";
        }
      }
      teamIdElement.attr("data-status", status);
    } else {
      // No team is present in this position for this match; blank out the status.
      teamIdElement.text("");
      teamIdElement.attr("data-status", "");
    }

    var wifiStatus = data.TeamWifiStatuses[station];
    teamRadioTextElement.text(wifiStatus.TeamId);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      var dsConn = stationStatus.DsConn;
      teamDsElement.attr("data-status-ok", dsConn.DsLinked);

      // Format the radio status box according to the connection status of the robot radio.
      var radioOkay = stationStatus.Team && stationStatus.Team.Id === wifiStatus.TeamId && wifiStatus.RadioLinked;
      teamRadioElement.attr("data-status-ok", radioOkay);

      // Format the robot status box.
      var robotOkay = dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked;
      teamRobotElement.attr("data-status-ok", robotOkay);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        teamRobotElement.text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        teamRobotElement.text(dsConn.BatteryVoltage.toFixed(1) + "V");
      }
    } else {
      teamDsElement.attr("data-status-ok", "");
      teamRobotElement.attr("data-status-ok", "");
      teamRobotElement.text("RBT");

      // Format the robot status box according to whether the AP is configured with the correct SSID.
      var expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
      if (wifiStatus.TeamId === expectedTeamId) {
        if (wifiStatus.RadioLinked) {
          teamRadioElement.attr("data-status-ok", true);
        } else {
          teamRadioElement.attr("data-status-ok", "");
        }
      } else {
        teamRadioElement.attr("data-status-ok", false);
      }
    }

    if (stationStatus.Estop) {
      teamBypassElement.attr("data-status-ok", false);
      teamBypassElement.text("ES");
    } else if (stationStatus.Bypass) {
      teamBypassElement.attr("data-status-ok", false);
      teamBypassElement.text("BYP");
    } else {
      teamBypassElement.attr("data-status-ok", true);
      teamBypassElement.text("ES");
    }
  });
};

$(function() {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
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
  websocket = new CheesyWebsocket("/displays/field_monitor/websocket", {
    arenaStatus: function(event) { handleArenaStatus(event.data); }
  });
});
