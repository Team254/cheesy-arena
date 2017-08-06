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
  $("input[name=" + alliance + "AutoMobility]").val(result.score.AutoMobility);
  $("input[name=" + alliance + "AutoRotors]").val(result.score.AutoRotors);
  $("input[name=" + alliance + "AutoFuelLow]").val(result.score.AutoFuelLow);
  $("input[name=" + alliance + "AutoFuelHigh]").val(result.score.AutoFuelHigh);

  $("input[name=" + alliance + "Rotors]").val(result.score.Rotors);
  $("input[name=" + alliance + "FuelLow]").val(result.score.FuelLow);
  $("input[name=" + alliance + "FuelHigh]").val(result.score.FuelHigh);
  $("input[name=" + alliance + "Takeoffs]").val(result.score.Takeoffs);

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      $("input[name=" + alliance + "Foul" + k + "Team][value=" + v.TeamId + "]").prop("checked", true);
      $("input[name=" + alliance + "Foul" + k + "RuleNumber]").val(v.RuleNumber);
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

  result.score.AutoMobility = parseInt(formData[alliance + "AutoMobility"]);
  result.score.AutoRotors = parseInt(formData[alliance + "AutoRotors"]);
  result.score.AutoFuelLow = parseInt(formData[alliance + "AutoFuelLow"]);
  result.score.AutoFuelHigh = parseInt(formData[alliance + "AutoFuelHigh"]);
  result.score.Rotors = parseInt(formData[alliance + "Rotors"]);
  result.score.FuelLow = parseInt(formData[alliance + "FuelLow"]);
  result.score.FuelHigh = parseInt(formData[alliance + "FuelHigh"]);
  result.score.Takeoffs = parseInt(formData[alliance + "Takeoffs"]);

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), RuleNumber: formData[prefix + "RuleNumber"],
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
