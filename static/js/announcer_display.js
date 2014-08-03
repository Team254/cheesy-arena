// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
var teamTemplate = Handlebars.compile($("#teamTemplate").html());
var matchResultTemplate = Handlebars.compile($("#matchResultTemplate").html());

var handleSetAudienceDisplay = function(targetScreen) {
  // Hide the final results so that they aren't blocking the current teams when the announcer needs them most.
  if (targetScreen == "intro" || targetScreen == "match") {
    $("#matchResult").modal("hide");
  }
};

var handleSetMatch = function(data) {
  $("#matchName").text(data.MatchType + " Match " + data.MatchDisplayName);
  $("#red1").html(teamTemplate(formatTeam(data.Red1)));
  $("#red2").html(teamTemplate(formatTeam(data.Red2)));
  $("#red3").html(teamTemplate(formatTeam(data.Red3)));
  $("#blue1").html(teamTemplate(formatTeam(data.Blue1)));
  $("#blue2").html(teamTemplate(formatTeam(data.Blue2)));
  $("#blue3").html(teamTemplate(formatTeam(data.Blue3)));
};

var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
  });
};

var handleRealtimeScore = function(data) {
  $("#redScore").text(data.RedScore);
  $("#blueScore").text(data.BlueScore);
};

var handleSetFinalScore = function(data) {
console.log(data);
  $("#scoreMatchName").text(data.MatchType + " Match " + data.MatchDisplayName);
  $("#redScoreDetails").html(matchResultTemplate({score: data.RedScoreSummary, fouls: data.RedFouls}));
  $("#blueScoreDetails").html(matchResultTemplate({score: data.BlueScoreSummary, fouls: data.BlueFouls}));
  $("#matchResult").modal("show");
};

var postMatchResult = function(data) {
  $("#savedMatchResult").attr("data-blink", false);
  websocket.send("setAudienceDisplay", "score");
}

// Replaces newlines in team fields with HTML line breaks.
var formatTeam = function(team) {
  team.Accomplishments = team.Accomplishments.replace(/[\r\n]+/g, "<br />");
  return team;
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/announcer/websocket", {
    setMatch: function(event) { handleSetMatch(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    setFinalScore: function(event) { handleSetFinalScore(event.data); },
    setAudienceDisplay: function(event) { handleSetAudienceDisplay(event.data); }
  });

  // Make the score blink.
  setInterval(function() {
    var blinkOn = $("#savedMatchResult").attr("data-blink") == "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
