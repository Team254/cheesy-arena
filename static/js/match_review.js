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
  $("input[name=" + alliance + "AutoRobotSet]").prop("checked", result.score.AutoRobotSet);
  $("input[name=" + alliance + "AutoContainerSet]").prop("checked", result.score.AutoContainerSet);
  $("input[name=" + alliance + "AutoToteSet]").prop("checked", result.score.AutoToteSet);
  $("input[name=" + alliance + "AutoStackedToteSet]").prop("checked", result.score.AutoStackedToteSet);
  $("input[name=" + alliance + "CoopertitionSet]").prop("checked", result.score.CoopertitionSet);
  $("input[name=" + alliance + "CoopertitionStack]").prop("checked", result.score.CoopertitionStack);

  if (result.score.Stacks != null) {
    $.each(result.score.Stacks, function(k, v) {
      $("#" + alliance + "Stack" + k + "Title").text("Stack " + (k + 1));
      $("input[name=" + alliance + "Stack" + k + "Totes]").val(v.Totes);
      $("input[name=" + alliance + "Stack" + k + "Container]").prop("checked", v.Container);
      $("input[name=" + alliance + "Stack" + k + "Litter]").prop("checked", v.Litter);
    });
  }

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      $("input[name=" + alliance + "Foul" + k + "Team][value=" + v.TeamId + "]").prop("checked", true);
      $("input[name=" + alliance + "Foul" + k + "Rule]").val(v.Rule);
      $("input[name=" + alliance + "Foul" + k + "Time]").val(v.TimeInMatchSec);
    });
  }

  if (result.cards != null) {
    $.each(result.cards, function(k, v) {
      $("input[name=" + alliance + "Team" + k + "Card][value=" + v + "]").prop("checked", true);
    });
  }
}

// Converts the current form values back into JSON structures and caches them.
var updateResults = function(alliance) {
  var result = allianceResults[alliance];
  var formData = {}
  $.each($("form").serializeArray(), function(k, v) {
    formData[v.name] = v.value;
  });

  result.score.AutoRobotSet = formData[alliance + "AutoRobotSet"] == "on";
  result.score.AutoContainerSet = formData[alliance + "AutoContainerSet"] == "on";
  result.score.AutoToteSet = formData[alliance + "AutoToteSet"] == "on";
  result.score.AutoStackedToteSet = formData[alliance + "AutoStackedToteSet"] == "on";
  result.score.CoopertitionSet = formData[alliance + "CoopertitionSet"] == "on";
  result.score.CoopertitionStack = formData[alliance + "CoopertitionStack"] == "on";

  result.score.Stacks = [];
  for (var i = 0; formData[alliance + "Stack" + i + "Totes"]; i++) {
    var prefix = alliance + "Stack" + i;
    var stack = {Totes: parseInt(formData[prefix + "Totes"]),
        Container: formData[prefix + "Container"] == "on",
        Litter: formData[prefix + "Litter"] == "on"}
    result.score.Stacks.push(stack);
  }

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), Rule: formData[prefix + "Rule"],
                TimeInMatchSec: parseFloat(formData[prefix + "Time"])};
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function(i, team) {
    result.cards[team] = formData[alliance + "Team" + team + "Card"];
  });
}

// Appends a blank stack to the end of the list.
var addStack = function(alliance) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Stacks.push({Totes: 0, Container: false, Litter: false})
  renderResults(alliance);
}

// Removes the given stack from the list.
var deleteStack = function(alliance, index) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Stacks.splice(index, 1);
  renderResults(alliance);
}

// Appends a blank foul to the end of the list.
var addFoul = function(alliance) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.push({TeamId: 0, Rule: "", TimeInMatchSec: 0})
  renderResults(alliance);
}

// Removes the given foul from the list.
var deleteFoul = function(alliance, index) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.splice(index, 1);
  renderResults(alliance);
}
