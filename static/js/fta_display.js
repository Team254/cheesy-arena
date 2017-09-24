// Copyright 2014 Team 254. All Rights Reserved.
// Author: austin.linux@gmail.com (Austin Schuh)
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the FTA diagnostic display.

var websocket;

// Handles a websocket message to update the team connection status.
var handleStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    if (stationStatus.Team) {
      $("#status" + station + " .team").text(stationStatus.Team.Id);
    } else {
      $("#status" + station + " .team").text("");
    }

    if (stationStatus.DsConn) {
      var dsConn = stationStatus.DsConn;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsConn.DsLinked);
      $("#status" + station + " .ds-status").text(dsConn.MBpsToRobot.toFixed(1) + "/" + dsConn.MBpsFromRobot.toFixed(1));
      $("#status" + station + " .radio-status").attr("data-status-ok", dsConn.RadioLinked);
      $("#status" + station + " .robot-status").attr("data-status-ok", dsConn.RobotLinked);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        $("#status" + station + " .robot-status").text("");
      }
      var lowBatteryThreshold = 6;
      if (matchStates[data.MatchState] == "PRE_MATCH") {
        lowBatteryThreshold = 12;
      }
      $("#status" + station + " .battery-status").attr("data-status-ok",
          dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked);
      $("#status" + station + " .battery-status").text(dsConn.BatteryVoltage.toFixed(1) + "V");
      $("#status" + station + " .trip-time").attr("data-status-ok", true);
      $("#status" + station + " .trip-time").text(dsConn.DsRobotTripTimeMs.toFixed(1) + "ms");
      $("#status" + station + " .packet-loss").attr("data-status-ok", true);
      $("#status" + station + " .packet-loss").text(dsConn.MissedPacketCount);
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .ds-status").text("");
      $("#status" + station + " .radio-status").attr("data-status-ok", "");
      $("#status" + station + " .radio-status").text("");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");
      $("#status" + station + " .battery-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").text("");
      $("#status" + station + " .trip-time").attr("data-status-ok", "");
      $("#status" + station + " .trip-time").text("");
      $("#status" + station + " .packet-loss").attr("data-status-ok", "");
      $("#status" + station + " .packet-loss").text("");
    }

    if (stationStatus.Estop) {
      $("#status" + station + " .bypass-status-fta").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status-fta").text("ES");
    } else if (stationStatus.Bypass) {
      $("#status" + station + " .bypass-status-fta").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status-fta").text("B");
    } else {
      $("#status" + station + " .bypass-status-fta").attr("data-status-ok", true);
      $("#status" + station + " .bypass-status-fta").text("");
    }
  });
};

$(function() {
  // Activate tooltips above the status headers.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/fta/websocket", {
    status: function(event) { handleStatus(event.data); }
  });
});
