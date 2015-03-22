// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
var foulAlliance;
var foulTeam;
var foulRule;

// Handles a click on a team button.
var setFoulTeam = function(teamButton) {
  foulAlliance = $(teamButton).attr("data-alliance");
  foulTeam = $(teamButton).attr("data-team");
  setSelections();
};

// Handles a click on a rule button.
var setFoulRule = function(ruleButton) {
  foulRule = $(ruleButton).attr("data-rule");
  setSelections();
};

// Sets button styles to match the selection cached in the global variables.
var setSelections = function() {
  $("[data-team]").each(function(i, teamButton) {
    $(teamButton).attr("data-selected", $(teamButton).attr("data-team") == foulTeam);
  });

  $("[data-rule]").each(function(i, ruleButton) {
    $(ruleButton).attr("data-selected", $(ruleButton).attr("data-rule") == foulRule);
  });

  $("#commit").prop("disabled", (foulTeam == "" || foulRule == ""));
};

// Resets the buttons to their default selections.
var clearFoul = function() {
  foulTeam = "";
  foulRule = "";
  setSelections();
};

// Sends the foul to the server to add it to the list.
var commitFoul = function() {
  websocket.send("addFoul", {Alliance: foulAlliance, TeamId: parseInt(foulTeam), Rule: foulRule});
};

// Removes the foul with the given parameters from the list.
var deleteFoul = function(alliance, team, rule, timeSec) {
  websocket.send("deleteFoul", {Alliance: alliance, TeamId: parseInt(team), Rule: rule,
      TimeInMatchSec: timeSec});
};

// Cycles through no card, yellow card, and red card.
var cycleCard = function(cardButton) {
  var newCard = "";
  if ($(cardButton).attr("data-card") == "") {
    newCard = "yellow";
  } else if ($(cardButton).attr("data-card") == "yellow") {
    newCard = "red";
  }
  websocket.send("card", {Alliance: $(cardButton).attr("data-alliance"),
      TeamId: parseInt($(cardButton).attr("data-card-team")), Card: newCard});
  $(cardButton).attr("data-card", newCard);
};

// Signals to the teams that they may enter the field.
var signalReset = function() {
  websocket.send("signalReset");
};

// Signals the scorekeeper that foul entry is complete for this match.
var commitMatch = function() {
  websocket.send("commitMatch");
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/referee/websocket", {
  });

  clearFoul();
});
