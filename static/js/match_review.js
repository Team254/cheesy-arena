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
  $("input[name=" + alliance + "AutoDefense1Crossings]").val(result.score.AutoDefensesCrossed[0]);
  $("input[name=" + alliance + "AutoDefense2Crossings]").val(result.score.AutoDefensesCrossed[1]);
  $("input[name=" + alliance + "AutoDefense3Crossings]").val(result.score.AutoDefensesCrossed[2]);
  $("input[name=" + alliance + "AutoDefense4Crossings]").val(result.score.AutoDefensesCrossed[3]);
  $("input[name=" + alliance + "AutoDefense5Crossings]").val(result.score.AutoDefensesCrossed[4]);
  $("input[name=" + alliance + "AutoDefensesReached]").val(result.score.AutoDefensesReached);
  $("input[name=" + alliance + "AutoHighGoals]").val(result.score.AutoHighGoals);
  $("input[name=" + alliance + "AutoLowGoals]").val(result.score.AutoLowGoals);

  $("input[name=" + alliance + "Defense1Crossings]").val(result.score.DefensesCrossed[0]);
  $("input[name=" + alliance + "Defense2Crossings]").val(result.score.DefensesCrossed[1]);
  $("input[name=" + alliance + "Defense3Crossings]").val(result.score.DefensesCrossed[2]);
  $("input[name=" + alliance + "Defense4Crossings]").val(result.score.DefensesCrossed[3]);
  $("input[name=" + alliance + "Defense5Crossings]").val(result.score.DefensesCrossed[4]);
  $("input[name=" + alliance + "HighGoals]").val(result.score.HighGoals);
  $("input[name=" + alliance + "LowGoals]").val(result.score.LowGoals);
  $("input[name=" + alliance + "Challenges]").val(result.score.Challenges);
  $("input[name=" + alliance + "Scales]").val(result.score.Scales);

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      $("input[name=" + alliance + "Foul" + k + "Team][value=" + v.TeamId + "]").prop("checked", true);
      $("input[name=" + alliance + "Foul" + k + "Rule]").val(v.Rule);
      $("input[name=" + alliance + "Foul" + k + "IsTechnical]").prop("checked", v.IsTechnical);
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

  result.score.AutoDefensesCrossed = [];
  result.score.AutoDefensesCrossed.push(parseInt(formData[alliance + "AutoDefense1Crossings"]));
  result.score.AutoDefensesCrossed.push(parseInt(formData[alliance + "AutoDefense2Crossings"]));
  result.score.AutoDefensesCrossed.push(parseInt(formData[alliance + "AutoDefense3Crossings"]));
  result.score.AutoDefensesCrossed.push(parseInt(formData[alliance + "AutoDefense4Crossings"]));
  result.score.AutoDefensesCrossed.push(parseInt(formData[alliance + "AutoDefense5Crossings"]));
  result.score.AutoDefensesReached = parseInt(formData[alliance + "AutoDefensesReached"]);
  result.score.AutoHighGoals = parseInt(formData[alliance + "AutoHighGoals"]);
  result.score.AutoLowGoals = parseInt(formData[alliance + "AutoLowGoals"]);
  result.score.DefensesCrossed = [];
  result.score.DefensesCrossed.push(parseInt(formData[alliance + "Defense1Crossings"]));
  result.score.DefensesCrossed.push(parseInt(formData[alliance + "Defense2Crossings"]));
  result.score.DefensesCrossed.push(parseInt(formData[alliance + "Defense3Crossings"]));
  result.score.DefensesCrossed.push(parseInt(formData[alliance + "Defense4Crossings"]));
  result.score.DefensesCrossed.push(parseInt(formData[alliance + "Defense5Crossings"]));
  result.score.HighGoals = parseInt(formData[alliance + "HighGoals"]);
  result.score.LowGoals = parseInt(formData[alliance + "LowGoals"]);
  result.score.Challenges = parseInt(formData[alliance + "Challenges"]);
  result.score.Scales = parseInt(formData[alliance + "Scales"]);

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), Rule: formData[prefix + "Rule"],
                IsTechnical: formData[prefix + "IsTechnical"] == "on",
                TimeInMatchSec: parseFloat(formData[prefix + "Time"])};
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function(i, team) {
    result.cards[team] = formData[alliance + "Team" + team + "Card"];
  });
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
