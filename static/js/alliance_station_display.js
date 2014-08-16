// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance station display.

var allianceStation = "";
var blinkInterval;
var websocket;

var handleSetAllianceStationDisplay = function(targetScreen) {
  switch (targetScreen) {
    case "logo":
      $("#match").hide();
      $("#logo").show();
      break;
    case "blank":
      $("#match").hide();
      $("#logo").hide();
      break;
    case "match":
      $("#match").show();
      $("#logo").hide();
      break;
  }
};

var handleSetMatch = function(data) {
  if (allianceStation != "" && data.AllianceStation == "") {
    // The client knows better what display this should be; let the server know.
    websocket.send("setAllianceStation", allianceStation);
  } else if (allianceStation != data.AllianceStation) {
    // The server knows better what display this should be; sync up.
    allianceStation = data.AllianceStation;
  }

  if (allianceStation != "") {
    team = data.Teams[allianceStation];
    if (team == null) {
      $("#teamId").text("");
      $("#teamName").text("");
      $("#teamName").attr("data-alliance", "");
    } else {
      $("#teamName").attr("data-alliance", allianceStation[0]);
      $("#teamId").text(data.Teams[allianceStation].Id);
      $("#teamName").text(data.Teams[allianceStation].Nickname);
    }
    $("#displayId").hide();
    $("#teamId").show();
    $("#teamName").show();
  } else {
    // Show the display ID so that someone can assign it to a station from the configuration interface.
    $("#teamId").text("");
    $("#teamName").text("");
    $("#displayId").show();
    $("#teamId").hide();
    $("#teamName").hide();
  }
};

var handleStatus = function(data) {
  stationStatus = data.AllianceStations[allianceStation];
  var blink = false;
  if (stationStatus.Bypass) {
    $("#match").attr("data-status", "bypass");
  } else if (stationStatus.DsConn) {
    if (!stationStatus.DsConn.DriverStationStatus.DsLinked) {
      $("#match").attr("data-status", allianceStation[0]);
    } else if (!stationStatus.DsConn.DriverStationStatus.RobotLinked) {
      blink = true;
      if (!blinkInterval) {
        blinkInterval = setInterval(function() {
          var status = $("#match").attr("data-status");
          $("#match").attr("data-status", (status == "") ? allianceStation[0] : "");
        }, 250);
      }
    } else {
      $("#match").attr("data-status", "");
    }
  }

  if (!blink && blinkInterval) {
    clearInterval(blinkInterval);
    blinkInterval = null;
  }
};

var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length == 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);

    if (matchState == "PRE_MATCH" || matchState == "POST_MATCH") {
      $("#teamName").show();
      $("#matchInfo").hide();
    } else {
      $("#teamName").hide();
      $("#matchInfo").show();
    }
  });
};

var handleRealtimeScore = function(data) {
  $("#redScore").text(data.RedScore);
  $("#blueScore").text(data.BlueScore);
};

$(function() {
  if (displayId == "") {
    displayId = Math.floor(Math.random() * 10000);
    window.location = "/displays/alliance_station?displayId=" + displayId;
  }
  $("#displayId").text(displayId);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/alliance_station/websocket?displayId=" + displayId, {
    setAllianceStationDisplay: function(event) { handleSetAllianceStationDisplay(event.data); },
    setMatch: function(event) { handleSetMatch(event.data); },
    status: function(event) { handleStatus(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); }
  });
});
