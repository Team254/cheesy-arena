// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
let isFirstScorePosted = true;

// Handles a websocket message to hide the score dialog once the next match is being introduced.
var handleAudienceDisplayMode = function(targetScreen) {
  // Hide the final results so that they aren't blocking the current teams when the announcer needs them most.
  if (targetScreen === "intro" || targetScreen === "match") {
    $("#matchResult").modal("hide");
  }
};

// Handles a websocket message to update the event status message.
const handleEventStatus = function(data) {
  if (data.CycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
  }
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);

  const teams = $("#teams");
  teams.empty();

  fetch("/displays/announcer/match_load")
    .then(response => response.text())
    .then(html => teams.html(html));
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.EndgamePoints);
  $("#blueScore").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.EndgamePoints);
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function(data) {
  if (isFirstScorePosted) {
    // Don't show the final score dialog when the page is first loaded.
    isFirstScorePosted = false;
    return;
  }

  const matchResult = $("#matchResult");
  fetch("/displays/announcer/score_posted")
    .then(response => response.text())
    .then(html => matchResult.html(html));
  matchResult.modal("show");

  // Activate tooltips above the foul listings.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/announcer/websocket", {
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    eventStatus: function(event) { handleEventStatus(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scorePosted: function(event) { handleScorePosted(event.data); }
  });

  // Make the score blink.
  setInterval(function() {
    var blinkOn = $("#savedMatchResult").attr("data-blink") === "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
