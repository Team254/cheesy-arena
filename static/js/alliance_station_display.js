// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance station display.

var allianceStation = "";
var blinkInterval;
var currentScreen = "blank";
var websocket;

// Handles a websocket message to change which screen is displayed.
var handleSetAllianceStationDisplay = function(targetScreen) {
  currentScreen = targetScreen;
  if (allianceStation == "") {
    // Don't show anything if this screen hasn't been assigned a position yet.
    targetScreen = "blank";
  }
  $("body").attr("data-mode", targetScreen);
  switch (allianceStation[1]) {
    case "1":
      $("body").attr("data-position", "right");
      break;
    case "2":
      $("body").attr("data-position", "middle");
      break;
    case "3":
      $("body").attr("data-position", "left");
      break;
    case "T":
      if (allianceStation[0] == "R") {
        $("body").attr("data-position", "redTower");
      } else {
        $("body").attr("data-position", "blueTower");
      }
      break;
  }
};

// Handles a websocket message to update the team to display.
var handleSetMatch = function(data) {
  if (allianceStation != "" && data.AllianceStation == "") {
    // The client knows better what display this should be; let the server know.
    websocket.send("setAllianceStation", allianceStation);
  } else if (allianceStation != data.AllianceStation) {
    // The server knows better what display this should be; sync up.
    allianceStation = data.AllianceStation;
    handleSetAllianceStationDisplay(currentScreen);
  }

  if (allianceStation != "") {
    var team = data.Teams[allianceStation];
    if (team) {
      $("#teamNumber").text(team.Id);
      $("#teamNameText").attr("data-alliance-bg", allianceStation[0]).text(team.Nickname);

      var ranking = data.Rankings[team.Id];
      if (ranking) {
        var rankingText = ranking.Rank + "-" + ranking.QualificationAverage;
        $("#teamRank").attr("data-alliance-bg", allianceStation[0]).text(rankingText);
      } else {
        $("#teamRank").attr("data-alliance-bg", allianceStation[0]).text("");
      }
    } else {
      $("#teamNumber").text("");
      $("#teamNameText").attr("data-alliance-bg", allianceStation[0]).text("");
      $("#teamRank").attr("data-alliance-bg", allianceStation[0]).text("");
    }
  } else {
    $("body").attr("data-mode", "displayId");
  }
};

// Handles a websocket message to update the team connection status.
var handleStatus = function(data) {
  stationStatus = data.AllianceStations[allianceStation];
  var blink = false;
  if (stationStatus && stationStatus.Bypass) {
    $("#match").attr("data-status", "bypass");
  } else if (stationStatus && stationStatus.DsConn) {
    if (!stationStatus.DsConn.DsLinked) {
      $("#match").attr("data-status", allianceStation[0]);
    } else if (!stationStatus.DsConn.RobotLinked) {
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

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length == 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#timeRemaining").text(countdownString);
    $("#match").attr("data-state", matchState);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.RedScoreFields.Score);
  $("#blueScore").text(data.BlueScoreFields.Score);
  $("#redTower").text(data.RedScoreFields.TowerStrength);
  $("#blueTower").text(data.BlueScoreFields.TowerStrength);
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
