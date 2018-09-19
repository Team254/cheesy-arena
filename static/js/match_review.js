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
  $("input[name=" + alliance + "AutoRuns]").val(result.score.AutoRuns);
  $("input[name=" + alliance + "AutoScaleOwnershipSec]").val(result.score.AutoScaleOwnershipSec);
  $("input[name=" + alliance + "AutoSwitchOwnershipSec]").val(result.score.AutoSwitchOwnershipSec);
  $("input[name=" + alliance + "AutoEndSwitchOwnership]").prop("checked", result.score.AutoEndSwitchOwnership);

  $("input[name=" + alliance + "TeleopScaleOwnershipSec]").val(result.score.TeleopScaleOwnershipSec);
  $("input[name=" + alliance + "TeleopScaleBoostSec]").val(result.score.TeleopScaleBoostSec);
  $("input[name=" + alliance + "TeleopSwitchOwnershipSec]").val(result.score.TeleopSwitchOwnershipSec);
  $("input[name=" + alliance + "TeleopSwitchBoostSec]").val(result.score.TeleopSwitchBoostSec);
  $("input[name=" + alliance + "ForceCubes]").val(result.score.ForceCubes);
  $("input[name=" + alliance + "ForceCubesPlayed]").val(result.score.ForceCubesPlayed);
  $("input[name=" + alliance + "LevitateCubes]").val(result.score.LevitateCubes);
  $("input[name=" + alliance + "LevitatePlayed]").prop("checked", result.score.LevitatePlayed);
  $("input[name=" + alliance + "BoostCubes]").val(result.score.BoostCubes);
  $("input[name=" + alliance + "BoostCubesPlayed]").val(result.score.BoostCubesPlayed);
  $("input[name=" + alliance + "Climbs]").val(result.score.Climbs);
  $("input[name=" + alliance + "Parks]").val(result.score.Parks);

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
};

// Converts the current form values back into JSON structures and caches them.
var updateResults = function(alliance) {
  var result = allianceResults[alliance];
  var formData = {};
  $.each($("form").serializeArray(), function(k, v) {
    formData[v.name] = v.value;
  });

  result.score.AutoRuns = parseInt(formData[alliance + "AutoRuns"]);
  result.score.AutoScaleOwnershipSec = parseFloat(formData[alliance + "AutoScaleOwnershipSec"]);
  result.score.AutoSwitchOwnershipSec = parseFloat(formData[alliance + "AutoSwitchOwnershipSec"]);
  result.score.AutoEndSwitchOwnership = formData[alliance + "AutoEndSwitchOwnership"] === "on";
  result.score.TeleopScaleOwnershipSec = parseFloat(formData[alliance + "TeleopScaleOwnershipSec"]);
  result.score.TeleopScaleBoostSec = parseFloat(formData[alliance + "TeleopScaleBoostSec"]);
  result.score.TeleopSwitchOwnershipSec = parseFloat(formData[alliance + "TeleopSwitchOwnershipSec"]);
  result.score.TeleopSwitchBoostSec = parseFloat(formData[alliance + "TeleopSwitchBoostSec"]);
  result.score.ForceCubes = parseInt(formData[alliance + "ForceCubes"]);
  result.score.ForceCubesPlayed = parseInt(formData[alliance + "ForceCubesPlayed"]);
  result.score.LevitateCubes = parseInt(formData[alliance + "LevitateCubes"]);
  result.score.LevitatePlayed = formData[alliance + "LevitatePlayed"] === "on";
  result.score.BoostCubes = parseInt(formData[alliance + "BoostCubes"]);
  result.score.BoostCubesPlayed = parseInt(formData[alliance + "BoostCubesPlayed"]);
  result.score.Climbs = parseInt(formData[alliance + "Climbs"]);
  result.score.Parks = parseInt(formData[alliance + "Parks"]);

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), RuleNumber: formData[prefix + "RuleNumber"],
                IsTechnical: formData[prefix + "IsTechnical"] === "on",
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
