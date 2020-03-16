// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
var foulTeamButton;
var foulRuleButton;
var firstMatchLoad = true;

// Handles a click on a team button.
var setFoulTeam = function(teamButton) {
  if (foulTeamButton) {
    foulTeamButton.attr("data-selected", false);
  }
  foulTeamButton = $(teamButton);
  foulTeamButton.attr("data-selected", true);

  $("#commit").prop("disabled", !(foulTeamButton && foulRuleButton));
};

// Handles a click on a rule button.
var setFoulRule = function(ruleButton) {
  if (foulRuleButton) {
    foulRuleButton.attr("data-selected", false);
  }
  foulRuleButton = $(ruleButton);
  foulRuleButton.attr("data-selected", true);

  $("#commit").prop("disabled", !(foulTeamButton && foulRuleButton));
};

// Resets the buttons to their default selections.
var clearFoul = function() {
  if (foulTeamButton) {
    foulTeamButton.attr("data-selected", false);
    foulTeamButton = null;
  }
  if (foulRuleButton) {
    foulRuleButton.attr("data-selected", false);
    foulRuleButton = null;
  }
  $("#commit").prop("disabled", true);
};

// Sends the foul to the server to add it to the list.
var commitFoul = function() {
  websocket.send("addFoul", {Alliance: foulTeamButton.attr("data-alliance"),
      TeamId: parseInt(foulTeamButton.attr("data-team")), RuleId: parseInt(foulRuleButton.attr("data-rule-id"))});
};

// Removes the foul with the given parameters from the list.
var deleteFoul = function(alliance, team, ruleId, timeSec) {
  websocket.send("deleteFoul", {Alliance: alliance, TeamId: parseInt(team), RuleId: parseInt(ruleId),
      TimeInMatchSec: timeSec});
};

// Cycles through no card, yellow card, and red card.
var cycleCard = function(cardButton) {
  var newCard = "";
  if ($(cardButton).attr("data-card") === "") {
    newCard = "yellow";
  } else if ($(cardButton).attr("data-card") === "yellow") {
    newCard = "red";
  }
  websocket.send("card", {Alliance: $(cardButton).attr("data-alliance"),
      TeamId: parseInt($(cardButton).attr("data-card-team")), Card: newCard});
  $(cardButton).attr("data-card", newCard);
};

// Signals to the volunteers that they may enter the field.
var signalVolunteers = function() {
  websocket.send("signalVolunteers");
};

// Signals to the teams that they may enter the field.
var signalReset = function() {
  websocket.send("signalReset");
};

// Signals the scorekeeper that foul entry is complete for this match.
var commitMatch = function() {
  websocket.send("commitMatch");
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  // Since the server always sends a matchLoad message upon establishing the websocket connection, ignore the first one.
  if (!firstMatchLoad) {
    location.reload();
  }
  firstMatchLoad = false;
};

$(function() {
  // Activate tooltips above the rule buttons.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/referee/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data) }
  });

  clearFoul();
});
