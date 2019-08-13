// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for editing a match in the match review page.

var scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
var allianceResults = {};

// Hijack the form submission to inject the data in JSON form so that it's easier for the server to parse.
$("form").submit(function() {
  updateResults("red");
  updateResults("blue");

  var redScoreJson = JSON.stringify(allianceResults["red"].score);
  var blueScoreJson = JSON.stringify(allianceResults["blue"].score);
  var redCardsJson = JSON.stringify(allianceResults["red"].cards);
  var blueCardsJson = JSON.stringify(allianceResults["blue"].cards);

  // Inject the JSON data into the form as hidden inputs.
  $("<input />").attr("type", "hidden").attr("name", "redScoreJson").attr("value", redScoreJson).appendTo("form");
  $("<input />").attr("type", "hidden").attr("name", "blueScoreJson").attr("value", blueScoreJson).appendTo("form");
  $("<input />").attr("type", "hidden").attr("name", "redCardsJson").attr("value", redCardsJson).appendTo("form");
  $("<input />").attr("type", "hidden").attr("name", "blueCardsJson").attr("value", blueCardsJson).appendTo("form");

  return true;
});

// Draws the match-editing form for one alliance based on the cached result data.
var renderResults = function(alliance) {
  var result = allianceResults[alliance];
  var scoreContent = scoreTemplate(result);
  $("#" + alliance + "Score").html(scoreContent);

  // Set the values of the form fields from the JSON results data.
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    getInputElement(alliance, "RobotStartLevel" + i1, result.score.RobotStartLevels[i]).prop("checked", true);
    getInputElement(alliance, "SandstormBonus" + i1).prop("checked", result.score.SandstormBonuses[i]);
    getInputElement(alliance, "RobotEndLevel" + i1, result.score.RobotEndLevels[i]).prop("checked", true);
    getSelectElement(alliance, "RocketNearLeftBay" + i1).val(result.score.RocketNearLeftBays[i]);
    getSelectElement(alliance, "RocketNearRightBay" + i1).val(result.score.RocketNearRightBays[i]);
    getSelectElement(alliance, "RocketFarLeftBay" + i1).val(result.score.RocketFarLeftBays[i]);
    getSelectElement(alliance, "RocketFarRightBay" + i1).val(result.score.RocketFarRightBays[i]);
  }
  for (var i = 0; i < 8; i++) {
    var i1 = i + 1;
    getSelectElement(alliance, "CargoBayPreMatch" + i1).val(result.score.CargoBaysPreMatch[i]);
    getSelectElement(alliance, "CargoBay" + i1).val(result.score.CargoBays[i]);
  }

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      getInputElement(alliance, "Foul" + k + "Team", v.TeamId).prop("checked", true);
      getInputElement(alliance, "Foul" + k + "RuleNumber").val(v.RuleNumber);
      getInputElement(alliance, "Foul" + k + "IsTechnical").prop("checked", v.IsTechnical);
      getInputElement(alliance, "Foul" + k + "IsRankingPoint").prop("checked", v.IsRankingPoint);
      getInputElement(alliance, "Foul" + k + "Time").val(v.TimeInMatchSec);
    });
  }

  if (result.cards != null) {
    $.each(result.cards, function(k, v) {
      getInputElement(alliance, "Team" + k + "Card", v).prop("checked", true);
    });
  }
};

// Converts the current form values back into JSON structures and caches them.
var updateResults = function(alliance) {
  var result = allianceResults[alliance];
  var formData = {};
  $.each($("form").serializeArray(), function(k, v) {
    formData[v.name] = v.value;
  });

  result.score.RobotStartLevels = [];
  result.score.SandstormBonuses = [];
  result.score.CargoBaysPreMatch = [];
  result.score.CargoBays = [];
  result.score.RocketNearLeftBays = [];
  result.score.RocketNearRightBays = [];
  result.score.RocketFarLeftBays = [];
  result.score.RocketFarRightBays = [];
  result.score.RobotEndLevels = [];
  for (var i = 0; i < 3; i++) {
    result.score.RobotStartLevels[i] = parseInt(formData[alliance + "RobotStartLevel" + (i + 1)]);
    result.score.SandstormBonuses[i] = formData[alliance + "SandstormBonus" + (i + 1)] === "on";
    result.score.RobotEndLevels[i] = parseInt(formData[alliance + "RobotEndLevel" + (i + 1)]);
    result.score.RocketNearLeftBays[i] = parseInt(formData[alliance + "RocketNearLeftBay" + (i + 1)]);
    result.score.RocketNearRightBays[i] = parseInt(formData[alliance + "RocketNearRightBay" + (i + 1)]);
    result.score.RocketFarLeftBays[i] = parseInt(formData[alliance + "RocketFarLeftBay" + (i + 1)]);
    result.score.RocketFarRightBays[i] = parseInt(formData[alliance + "RocketFarRightBay" + (i + 1)]);
  }
  for (var i = 0; i < 8; i++) {
    result.score.CargoBaysPreMatch[i] = parseInt(formData[alliance + "CargoBayPreMatch" + (i + 1)]);
    result.score.CargoBays[i] = parseInt(formData[alliance + "CargoBay" + (i + 1)]);
  }

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), RuleNumber: formData[prefix + "RuleNumber"],
                IsTechnical: formData[prefix + "IsTechnical"] === "on",
                IsRankingPoint: formData[prefix + "IsRankingPoint"] === "on",
                TimeInMatchSec: parseFloat(formData[prefix + "Time"])};
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function(i, team) {
    result.cards[team] = formData[alliance + "Team" + team + "Card"];
  });
};

// Appends a blank foul to the end of the list.
var addFoul = function(alliance) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.push({TeamId: 0, Rule: "", TimeInMatchSec: 0});
  renderResults(alliance);
};

// Removes the given foul from the list.
var deleteFoul = function(alliance, index) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.splice(index, 1);
  renderResults(alliance);
};

// Returns the form input element having the given parameters.
var getInputElement = function(alliance, name, value) {
  var selector = "input[name=" + alliance + name + "]";
  if (value !== undefined) {
    selector += "[value=" + value + "]";
  }
  return $(selector);
};

// Returns the form select element having the given parameters.
var getSelectElement = function(alliance, name) {
  var selector = "select[name=" + alliance + name + "]";
  return $(selector);
};
