// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for editing a match in the match review page.

const scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
const allianceResults = {};
let matchResult;

// Hijack the form submission to inject the data in JSON form so that it's easier for the server to parse.
$("form").submit(function() {
  updateResults("red");
  updateResults("blue");

  matchResult.RedScore = allianceResults["red"].score;
  matchResult.BlueScore = allianceResults["blue"].score;
  matchResult.RedCards = allianceResults["red"].cards;
  matchResult.BlueCards = allianceResults["blue"].cards;
  const matchResultJson = JSON.stringify(matchResult);

  // Inject the JSON data into the form as hidden inputs.
  $("<input />").attr("type", "hidden").attr("name", "matchResultJson").attr("value", matchResultJson).appendTo("form");

  return true;
});

// Draws the match-editing form for one alliance based on the cached result data.
const renderResults = function(alliance) {
  const result = allianceResults[alliance];
  const scoreContent = scoreTemplate(result);
  $("#" + alliance + "Score").html(scoreContent);

  // Set the values of the form fields from the JSON results data.
  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;

    getInputElement(alliance, "MobilityStatuses" + i1).prop("checked", result.score.MobilityStatuses[i]);
    getInputElement(alliance, "AutoDockStatuses" + i1).prop("checked", result.score.AutoDockStatuses[i]);
    getInputElement(alliance, "EndgameStatuses" + i1, result.score.EndgameStatuses[i]).prop("checked", true);

    for (let j = 0; j < 9; j++) {
      getInputElement(alliance, `GridAutoScoringRow${i}Node${j}`).prop("checked", result.score.Grid.AutoScoring[i][j]);
      getSelectElement(alliance, `GridNodeStatesRow${i}Node${j}`).val(result.score.Grid.Nodes[i][j]);
    }
  }

  getInputElement(alliance, "AutoChargeStationLevel").prop("checked", result.score.AutoChargeStationLevel);
  getInputElement(alliance, "EndgameChargeStationLevel").prop("checked", result.score.EndgameChargeStationLevel);

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      getInputElement(alliance, "Foul" + k + "IsTechnical").prop("checked", v.IsTechnical);
      getInputElement(alliance, "Foul" + k + "Team", v.TeamId).prop("checked", true);
      getSelectElement(alliance, "Foul" + k + "RuleId").val(v.RuleId);
    });
  }

  if (result.cards != null) {
    $.each(result.cards, function(k, v) {
      getInputElement(alliance, "Team" + k + "Card", v).prop("checked", true);
    });
  }
};

// Converts the current form values back into JSON structures and caches them.
const updateResults = function(alliance) {
  const result = allianceResults[alliance];
  const formData = {};
  $.each($("form").serializeArray(), function(k, v) {
    formData[v.name] = v.value;
  });

  result.score.MobilityStatuses = [];
  result.score.Grid = {AutoScoring: [], Nodes: []};
  result.score.AutoDockStatuses = [];
  result.score.EndgameStatuses = [];
  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;

    result.score.MobilityStatuses[i] = formData[alliance + "MobilityStatuses" + i1] === "on";
    result.score.AutoDockStatuses[i] = formData[alliance + "AutoDockStatuses" + i1] === "on";
    result.score.EndgameStatuses[i] = parseInt(formData[alliance + "EndgameStatuses" + i1]);

    result.score.Grid.AutoScoring[i] = [];
    result.score.Grid.Nodes[i] = [];
    for (let j = 0; j < 9; j++) {
      result.score.Grid.AutoScoring[i][j] = formData[alliance + `GridAutoScoringRow${i}Node${j}`] === "on";
      result.score.Grid.Nodes[i][j] = parseInt(formData[alliance + `GridNodeStatesRow${i}Node${j}`]);
    }
  }

  result.score.AutoChargeStationLevel = formData[alliance + "AutoChargeStationLevel"] === "on";
  result.score.EndgameChargeStationLevel = formData[alliance + "EndgameChargeStationLevel"] === "on";

  result.score.Fouls = [];

  for (let i = 0; formData[alliance + "Foul" + i + "Index"]; i++) {
    const prefix = alliance + "Foul" + i;
    const foul = {
      IsTechnical: formData[prefix + "IsTechnical"] === "on",
      TeamId: parseInt(formData[prefix + "Team"]),
      RuleId: parseInt(formData[prefix + "RuleId"]),
    };
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function(i, team) {
    result.cards[team] = formData[alliance + "Team" + team + "Card"];
  });
};

// Appends a blank foul to the end of the list.
const addFoul = function(alliance) {
  updateResults(alliance);
  const result = allianceResults[alliance];
  result.score.Fouls.push({IsTechnical: false, TeamId: 0, Rule: 0});
  renderResults(alliance);
};

// Removes the given foul from the list.
const deleteFoul = function(alliance, index) {
  updateResults(alliance);
  const result = allianceResults[alliance];
  result.score.Fouls.splice(index, 1);
  renderResults(alliance);
};

// Returns the form input element having the given parameters.
const getInputElement = function(alliance, name, value) {
  let selector = "input[name=" + alliance + name + "]";
  if (value !== undefined) {
    selector += "[value=" + value + "]";
  }
  return $(selector);
};

// Returns the form select element having the given parameters.
const getSelectElement = function(alliance, name) {
  const selector = "select[name=" + alliance + name + "]";
  return $(selector);
};
