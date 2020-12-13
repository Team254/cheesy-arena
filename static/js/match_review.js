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
  for (var i = 0; i < 4; i++) {
    var i1 = i + 1;

    if (i < 2) {
      getInputElement(alliance, "AutoCellsBottom" + i1).val(result.score.AutoCellsBottom[i]);
      getInputElement(alliance, "AutoCellsOuter" + i1).val(result.score.AutoCellsOuter[i]);
      getInputElement(alliance, "AutoCellsInner" + i1).val(result.score.AutoCellsInner[i]);
    }

    if (i < 3) {
      getInputElement(alliance, "ExitedInitiationLine" + i1).prop("checked", result.score.ExitedInitiationLine[i]);
      getInputElement(alliance, "EndgameStatuses" + i1, result.score.EndgameStatuses[i]).prop("checked", true);
    }

    getInputElement(alliance, "TeleopCellsBottom" + i1).val(result.score.TeleopCellsBottom[i]);
    getInputElement(alliance, "TeleopCellsOuter" + i1).val(result.score.TeleopCellsOuter[i]);
    getInputElement(alliance, "TeleopCellsInner" + i1).val(result.score.TeleopCellsInner[i]);
  }
  getInputElement(alliance, "ControlPanelStatus", result.score.ControlPanelStatus).prop("checked", true);
  getInputElement(alliance, "RungIsLevel").prop("checked", result.score.RungIsLevel);

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      getInputElement(alliance, "Foul" + k + "Team", v.TeamId).prop("checked", true);
      getSelectElement(alliance, "Foul" + k + "RuleId").val(v.RuleId);
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

  result.score.ExitedInitiationLine = [];
  result.score.AutoCellsBottom = [];
  result.score.AutoCellsOuter = [];
  result.score.AutoCellsInner = [];
  result.score.TeleopCellsBottom = [];
  result.score.TeleopCellsOuter = [];
  result.score.TeleopCellsInner = [];
  result.score.EndgameStatuses = [];
  for (var i = 0; i < 4; i++) {
    var i1 = i + 1;

    if (i < 2) {
      result.score.AutoCellsBottom[i] = parseInt(formData[alliance + "AutoCellsBottom" + i1]);
      result.score.AutoCellsOuter[i] = parseInt(formData[alliance + "AutoCellsOuter" + i1]);
      result.score.AutoCellsInner[i] = parseInt(formData[alliance + "AutoCellsInner" + i1]);
    }

    if (i < 3) {
      result.score.ExitedInitiationLine[i] = formData[alliance + "ExitedInitiationLine" + i1] === "on";
      result.score.EndgameStatuses[i] = parseInt(formData[alliance + "EndgameStatuses" + i1]);
    }

    result.score.TeleopCellsBottom[i] = parseInt(formData[alliance + "TeleopCellsBottom" + i1]);
    result.score.TeleopCellsOuter[i] = parseInt(formData[alliance + "TeleopCellsOuter" + i1]);
    result.score.TeleopCellsInner[i] = parseInt(formData[alliance + "TeleopCellsInner" + i1]);
  }
  result.score.ControlPanelStatus = parseInt(formData[alliance + "ControlPanelStatus"]);
  result.score.RungIsLevel = formData[alliance + "RungIsLevel"] === "on";

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), RuleId: parseInt(formData[prefix + "RuleId"]),
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
