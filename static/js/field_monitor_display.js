// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field monitor display.

var websocket;
var currentMatchId;
var redSide;
var blueSide;
var lowBatteryThreshold = 8;
var highBtuThreshold = 4.0;


var handleArenaStatus = function(data) {
  // If getting data for the wrong match (e.g. after a server restart), reload the page.
  if (currentMatchId == null) {
    currentMatchId = data.MatchId;
  } else if (currentMatchId !== data.MatchId) {
    location.reload();
  }
	
  $.each(data.AllianceStations, function(station, stationStatus) {
    // Select the DOM elements corresponding to the team station.
    var teamElementPrefix;
    if (station[0] === "R") {
      teamElementPrefix = "#" + redSide + "Team" + station[1];
    } else {
      teamElementPrefix = "#" + blueSide + "Team" + station[1];
    }
    var teamIdElement = $(teamElementPrefix + "Id");
    var teamNotesElement = $(teamElementPrefix + "Notes");
    var teamNotesTextElement = $(teamElementPrefix + "Notes div");
    var teamEthernetElement = $(teamElementPrefix + "Ethernet");
    var teamDsElement = $(teamElementPrefix + "Ds");
    var teamRadioElement = $(teamElementPrefix + "Radio");
    var teamRadioTextElement = $(teamElementPrefix + "Radio span");
    var teamRobotElement = $(teamElementPrefix + "Robot");
    var teamBypassElement = $(teamElementPrefix + "Bypass");
    var teamBandwidthElement = $(teamElementPrefix + "Bandwidth");

    teamNotesTextElement.attr("data-station", station);

    if (stationStatus.Team) {
      // Set the team number and status.
      teamIdElement.text(stationStatus.Team.Id);
      var status = "no-link";
      if (stationStatus.Bypass) {
        status = "";
      } else if (stationStatus.DsConn) {
        if (stationStatus.DsConn.WrongStation) {
          status = "wrong-station";
        } else if (stationStatus.DsConn.RobotLinked) {
          status = "robot-linked";
        } else if (stationStatus.DsConn.RadioLinked) {
          status = "radio-linked";
        } else if (stationStatus.DsConn.DsLinked) {
          status = "ds-linked";
        }
      }
      teamIdElement.attr("data-status", status);
      teamNotesTextElement.text(stationStatus.Team.FtaNotes);
      teamNotesElement.attr("data-status", status);
    } else {
      // No team is present in this position for this match; blank out the status.
      teamIdElement.text("");
      teamNotesTextElement.text("");
      teamNotesElement.attr("data-status", "");
    }

    // Format the Ethernet status box.
    teamEthernetElement.attr("data-status-ok", stationStatus.Ethernet ? "true" : "");
    if (stationStatus.DsConn && stationStatus.DsConn.DsRobotTripTimeMs > 0) {
      teamEthernetElement.text(stationStatus.DsConn.DsRobotTripTimeMs);
    } else {
      teamEthernetElement.text("ETH");
    }

    var wifiStatus = data.TeamWifiStatuses[station];
    teamRadioTextElement.text(wifiStatus.TeamId);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      var dsConn = stationStatus.DsConn;
      teamDsElement.attr("data-status-ok", dsConn.DsLinked);
      teamDsElement.text(dsConn.MissedPacketCount);

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
      var btuOkay = wifiStatus.MBits < highBtuThreshold && dsConn.RobotLinked;
      if (wifiStatus.MBits >= 0.01) {
        teamBandwidthElement.text(wifiStatus.MBits.toFixed(2)+ "Mb");
        teamBandwidthElement.attr("data-status-ok", btuOkay);
      } else {
        teamBandwidthElement.text("-");
        teamBandwidthElement.attr("data-status-ok", btuOkay);
      }
    } else {
      teamDsElement.attr("data-status-ok", "");
      teamDsElement.text("DS");
      teamRobotElement.attr("data-status-ok", "");
      teamRobotElement.text("RBT");
      teamBandwidthElement.attr("data-status-ok", "");
      teamBandwidthElement.text("-");

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

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
    });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data,reversed) {

    if (reversed === "true") {
      $("#rightScore").text(data.Red.ScoreSummary.Score);
      $("#leftScore").text(data.Blue.ScoreSummary.Score);
    } else {
      $("#rightScore").text(data.Blue.ScoreSummary.Score);
      $("#leftScore").text(data.Red.ScoreSummary.Score);

    }
};

// Handles a websocket message to update current match
var handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);
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

// Makes the team notes section editable and handles saving edits to the server.
var editFtaNotes = function(element) {
  var teamNotesTextElement = $(element);
  var textArea = $("<textarea />");
  textArea.val(teamNotesTextElement.text());
  teamNotesTextElement.replaceWith(textArea);
  textArea.focus();
  textArea.blur(function() {
    textArea.replaceWith(teamNotesTextElement);
    if (textArea.val() !== teamNotesTextElement.text()) {
      websocket.send("updateTeamNotes", { station: teamNotesTextElement.attr("data-station"), notes: textArea.val()});
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
  $(".fta-dependent").attr("data-fta", urlParams.get("fta"));

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/field_monitor/websocket", {
    arenaStatus: function(event) { handleArenaStatus(event.data); },
    eventStatus: function(event) { handleEventStatus(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data,reversed); },
  });
});
