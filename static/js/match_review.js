// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for editing a match in the match review page.

const scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
const allianceResults = {};
let matchResult;

const NUM_ROBOTS = 3;
const NUM_HUB_SHIFTS = 8;

// Hijack the form submission to inject the data in JSON form so that it's easier for the server to parse.
$("form").submit(function () {
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
const renderResults = function (alliance) {
  const result = allianceResults[alliance];
  result.score = normalizeScore(result.score);
  const scoreContent = scoreTemplate(result);
  $(`#${alliance}Score`).html(scoreContent);

  // Set the values of the form fields from the JSON results data.
  getInputElement(alliance, "HubWonAuto").prop("checked", result.score.Hub.WonAuto);
  for (let i = 0; i < NUM_HUB_SHIFTS; i++) {
    getInputElement(alliance, `HubShiftCount${i}`).val(result.score.Hub.ShiftCounts[i]);
  }

  for (let i = 0; i < NUM_ROBOTS; i++) {
    const i1 = i + 1;

    getInputElement(alliance, `AutoTowerStatuses${i1}`, result.score.AutoTowerStatuses[i]).prop("checked", true);
    getInputElement(alliance, `EndgameTowerStatuses${i1}`, result.score.EndgameTowerStatuses[i]).prop("checked", true);
  }

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function (k, v) {
      getInputElement(alliance, `Foul${k}IsMajor`).prop("checked", v.IsMajor);
      getInputElement(alliance, `Foul${k}Team`, v.TeamId).prop("checked", true);
      getSelectElement(alliance, `Foul${k}RuleId`).val(v.RuleId);
    });
  }

  if (result.cards != null) {
    $.each(result.cards, function (k, v) {
      getInputElement(alliance, `Team${k}Card`, v).prop("checked", true);
    });
  }
};

// Converts the current form values back into JSON structures and caches them.
const updateResults = function (alliance) {
  const result = allianceResults[alliance];
  const formData = {};
  $.each($("form").serializeArray(), function (k, v) {
    formData[v.name] = v.value;
  });

  result.score.AutoTowerStatuses = [];
  result.score.Hub = {
    WonAuto: formData[`${alliance}HubWonAuto`] === "on",
    ShiftCounts: [],
  };
  result.score.EndgameTowerStatuses = [];
  for (let i = 0; i < NUM_HUB_SHIFTS; i++) {
    result.score.Hub.ShiftCounts[i] = parseFormInt(formData[`${alliance}HubShiftCount${i}`]);
  }
  for (let i = 0; i < NUM_ROBOTS; i++) {
    const i1 = i + 1;

    result.score.AutoTowerStatuses[i] = parseFormInt(formData[`${alliance}AutoTowerStatuses${i1}`]);
    result.score.EndgameTowerStatuses[i] = parseFormInt(formData[`${alliance}EndgameTowerStatuses${i1}`]);
  }

  result.score.Fouls = [];

  for (let i = 0; formData[`${alliance}Foul${i}Index`]; i++) {
    const prefix = `${alliance}Foul${i}`;
    const foul = {
      IsMajor: formData[`${prefix}IsMajor`] === "on",
      TeamId: parseFormInt(formData[`${prefix}Team`]),
      RuleId: parseFormInt(formData[`${prefix}RuleId`]),
    };
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function (i, team) {
    result.cards[team] = formData[`${alliance}Team${team}Card`];
  });
};

// Appends a blank foul to the end of the list.
const addFoul = function (alliance) {
  updateResults(alliance);
  const result = allianceResults[alliance];
  result.score.Fouls.push({IsMajor: false, TeamId: 0, RuleId: 0});
  renderResults(alliance);
};

// Removes the given foul from the list.
const deleteFoul = function (alliance, index) {
  updateResults(alliance);
  const result = allianceResults[alliance];
  result.score.Fouls.splice(index, 1);
  renderResults(alliance);
};

// Returns the form input element having the given parameters.
const getInputElement = function (alliance, name, value) {
  let selector = `input[name=${alliance}${name}]`;
  if (value !== undefined) {
    selector += `[value=${value}]`;
  }
  return $(selector);
};

// Returns the form select element having the given parameters.
const getSelectElement = function (alliance, name) {
  const selector = `select[name=${alliance}${name}]`;
  return $(selector);
};

const normalizeScore = function (score) {
  score = score || {};
  score.AutoTowerStatuses = normalizeArray(score.AutoTowerStatuses, NUM_ROBOTS, 0);
  score.EndgameTowerStatuses = normalizeArray(score.EndgameTowerStatuses, NUM_ROBOTS, 0);
  score.Hub = score.Hub || {};
  score.Hub.WonAuto = !!score.Hub.WonAuto;
  score.Hub.ShiftCounts = normalizeArray(score.Hub.ShiftCounts, NUM_HUB_SHIFTS, 0);
  score.Fouls = score.Fouls || [];
  return score;
};

const normalizeArray = function (array, length, defaultValue) {
  array = array || [];
  for (let i = 0; i < length; i++) {
    if (array[i] === undefined || array[i] === null) {
      array[i] = defaultValue;
    }
  }
  return array;
};

const parseFormInt = function (value) {
  const parsed = parseInt(value, 10);
  if (isNaN(parsed)) {
    return 0;
  }
  return parsed;
};
