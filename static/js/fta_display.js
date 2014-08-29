// Copyright 2014 Team 254. All Rights Reserved.
// Author: austin.linux@gmail.com (Austin Schuh)
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the FTA diagnostic display.

var websocket;

var handleStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    if (stationStatus.DsConn) {
      $("#status" + station + " .team").text(stationStatus.DsConn.TeamId);
      var dsStatus = stationStatus.DsConn.DriverStationStatus;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsStatus.DsLinked);
      $("#status" + station + " .robot-status").attr("data-status-ok", dsStatus.RobotLinked);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        $("#status" + station + " .robot-status").text("");
      }
      $("#status" + station + " .battery-status").attr("data-status-ok",
                                                       dsStatus.BatteryVoltage > 6 && dsStatus.RobotLinked);
      $("#status" + station + " .battery-status").text(dsStatus.BatteryVoltage.toFixed(1) + "V");
      $("#status" + station + " .trip-time").attr("data-status-ok", true);
      $("#status" + station + " .trip-time").text(dsStatus.DsRobotTripTimeMs.toFixed(1) + "ms");
      $("#status" + station + " .packet-loss").attr("data-status-ok", true);
      $("#status" + station + " .packet-loss").text(dsStatus.MissedPacketCount);
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");
      $("#status" + station + " .battery-status").attr("data-status-ok", "");
      $("#status" + station + " .battery-status").text("");
      $("#status" + station + " .trip-time").attr("data-status-ok", "");
      $("#status" + station + " .trip-time").text("");
      $("#status" + station + " .packet-loss").attr("data-status-ok", "");
      $("#status" + station + " .packet-loss").text("");
    }

    if (stationStatus.EmergencyStop) {
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
  websocket = new CheesyWebsocket("/match_play/websocket", {
    status: function(event) { handleStatus(event.data); }
  });
});
