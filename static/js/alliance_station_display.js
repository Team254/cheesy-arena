// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance station display.

// A unique id to differentiate this station's display from its peers.
var displayId;
var allianceStation = "";
var websocket;

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
    } else {
      $("#teamId").attr("data-alliance", allianceStation[0]);
      $("#teamName").attr("data-alliance", allianceStation[0]);
      $("#teamId").text(data.Teams[allianceStation].Id);
      $("#teamName").text(data.Teams[allianceStation].Nickname);
    }
    $("#displayIdRow").hide();
    $("#teamIdRow").show();
    $("#teamNameRow").show();
  } else {
    // Show the display ID so that someone can assign it to a station from the configuration interface.
    $("#teamId").text("");
    $("#teamName").text("");
    $("#displayIdRow").show();
    $("#teamIdRow").hide();
    $("#teamNameRow").hide();
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
      $("#teamNameRow").show();
      $("#matchInfoRow").hide();
    } else {
      $("#teamNameRow").hide();
      $("#matchInfoRow").show();
    }
  });
};

var handleRealtimeScore = function(data) {
  $("#redScore").text(data.RedScore);
  $("#blueScore").text(data.BlueScore);
};

$(function() {
  displayId = Math.floor(Math.random() * 10000);
  $("#displayId").text(displayId);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/alliance_station/websocket?displayId=" + displayId, {
    setMatch: function(event) { handleSetMatch(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); }
  });
});
