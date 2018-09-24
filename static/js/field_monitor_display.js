// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field monitor display.

var websocket;
var redSide;
var blueSide;

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function(data) {
  $.each(data.AllianceStations, function(station, stationStatus) {
    // Select the DOM element corresponding to the team station.
    var teamElement;
    if (station[0] === "R") {
      teamElement = $("#" + redSide + "Team" + station[1]);
    } else {
      teamElement = $("#" + blueSide + "Team" + station[1]);
    }

    if (stationStatus.Team) {
      // Set the team number and status.
      teamElement.text(stationStatus.Team.Id);
      var status = "no-link";
      if (stationStatus.Bypass) {
        status = "";
      } else if (stationStatus.DsConn) {
        if (stationStatus.DsConn.RobotLinked) {
          status = "robot-linked";
        } else if (stationStatus.DsConn.DsLinked) {
          status = "ds-linked";
        }
      }
      teamElement.attr("data-status", status);
    } else {
      // No team is present in this position for this match; blank out the status.
      teamElement.text("");
      teamElement.attr("data-status", "");
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
