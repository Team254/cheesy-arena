// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for editing a match in the match review page.

const allianceResults = {};
let matchResult;

const ALLIANCES = ["red", "blue"];
const NUM_ROBOTS = 3;
const NUM_HUB_SHIFTS = 8;
const SUMMARY_REFRESH_DELAY_MS = 150;
const RANKING_POINT_SUMMARY_FIELDS = [
  "EnergizedBonusRankingPoint",
  "SuperchargedBonusRankingPoint",
  "TraversalBonusRankingPoint",
];

let summaryRefreshTimer;
let latestSummaryRequestId = 0;

// Hijack the form submission to inject the data in JSON form so that it's easier for the server to parse.
$("form").submit(function () {
  updateAllResults();

  const matchResultJson = JSON.stringify(matchResult);

  // Inject the JSON data into the form as hidden inputs.
  $("input[name=matchResultJson]").remove();
  $("<input />").attr("type", "hidden").attr("name", "matchResultJson").attr("value", matchResultJson).appendTo("form");

  return true;
});

$("form").on("input change", "input, select", function () {
  scheduleScoreSummaryRefresh();
});

// Sets up the match-editing form for one alliance based on the cached result data.
const renderResults = function (alliance) {
  const result = allianceResults[alliance];
  result.score = normalizeScore(result.score);
  result.cards = result.cards || {};

  getInputElement(alliance, "HubWonAuto").prop("checked", result.score.Hub.WonAuto);
  for (let i = 0; i < NUM_HUB_SHIFTS; i++) {
    getInputElement(alliance, `HubShiftCount${i}`).val(result.score.Hub.ShiftCounts[i]);
  }

  for (let i = 0; i < NUM_ROBOTS; i++) {
    const i1 = i + 1;

    getInputElement(alliance, `AutoTowerStatuses${i1}`, result.score.AutoTowerStatuses[i]).prop("checked", true);
    getInputElement(alliance, `EndgameTowerStatuses${i1}`, result.score.EndgameTowerStatuses[i]).prop("checked", true);
  }

  renderFouls(alliance);
  renderCards(alliance);
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
    result.score.Fouls.push({
      IsMajor: formData[`${prefix}IsMajor`] === "on",
      TeamId: parseFormInt(formData[`${prefix}Team`]),
      RuleId: parseFormInt(formData[`${prefix}RuleId`]),
    });
  }

  result.cards = {};
  $.each(result.teams, function (i, team) {
    result.cards[team] = formData[`${alliance}Team${team}Card`];
  });
};

const updateAllResults = function () {
  updateResults("red");
  updateResults("blue");

  matchResult.RedScore = allianceResults["red"].score;
  matchResult.BlueScore = allianceResults["blue"].score;
  matchResult.RedCards = allianceResults["red"].cards;
  matchResult.BlueCards = allianceResults["blue"].cards;
};

// Appends a blank foul to the end of the list.
const addFoul = function (alliance) {
  updateResults(alliance);
  allianceResults[alliance].score.Fouls.push({IsMajor: false, TeamId: 0, RuleId: 0});
  renderFouls(alliance);
  refreshScoreSummaries();
};

// Removes the given foul from the list.
const deleteFoul = function (alliance, index) {
  updateResults(alliance);
  allianceResults[alliance].score.Fouls.splice(index, 1);
  renderFouls(alliance);
  refreshScoreSummaries();
};

const renderFouls = function (alliance) {
  const result = allianceResults[alliance];
  const foulContainer = $(`#${alliance}Fouls`);
  foulContainer.empty();

  $.each(result.score.Fouls, function (index, foul) {
    foulContainer.append(buildFoulElement(alliance, index, foul));
  });
};

const buildFoulElement = function (alliance, index, foul) {
  const result = allianceResults[alliance];
  const prefix = `${alliance}Foul${index}`;
  const element = cloneTemplateElement("foulTemplate").addClass(`bg-dark-${alliance}`);

  element.find("[data-foul-field=index]").attr("name", `${prefix}Index`).val(index);
  element.find("[data-foul-action=delete]").on("click", function () {
    deleteFoul(alliance, index);
  });
  element.find("[data-foul-field=isMajor]").attr("name", `${prefix}IsMajor`).prop("checked", foul.IsMajor);
  element.find("[data-foul-field=ruleId]").attr("name", `${prefix}RuleId`).val(foul.RuleId);

  const teamContainer = element.find("[data-foul-teams]");
  $.each(result.teams, function (i, team) {
    const teamOption = cloneTemplateElement("foulTeamOptionTemplate");
    teamOption.find("[data-foul-field=team]").attr("name", `${prefix}Team`).attr("value", team).prop(
      "checked", team === foul.TeamId
    );
    teamOption.find("[data-foul-team-label]").text(`Team ${team}`);
    teamContainer.append(teamOption);
  });

  return element;
};

const renderCards = function (alliance) {
  const result = allianceResults[alliance];
  $.each(result.cards, function (team, card) {
    getInputElement(alliance, `Team${team}Card`, card).prop("checked", true);
  });
};

const scheduleScoreSummaryRefresh = function () {
  window.clearTimeout(summaryRefreshTimer);
  summaryRefreshTimer = window.setTimeout(refreshScoreSummaries, SUMMARY_REFRESH_DELAY_MS);
};

const refreshScoreSummaries = function () {
  updateAllResults();
  const requestId = ++latestSummaryRequestId;

  $.ajax({
    url: `/match_review/${matchId}/summary`,
    method: "POST",
    contentType: "application/json",
    data: JSON.stringify(matchResult),
    success: function (data) {
      if (requestId !== latestSummaryRequestId) {
        return;
      }
      updateSummaryCard("red", data.RedSummary);
      updateSummaryCard("blue", data.BlueSummary);
    },
  });
};

const updateSummaryCard = function (alliance, summary) {
  const summaryCard = $(`#${alliance}Summary`);
  $.each(summary, function (field, value) {
    summaryCard.find(`[data-summary-field=${field}]`).html(formatSummaryValue(field, value));
  });
};

// Returns the form input element having the given parameters.
const getInputElement = function (alliance, name, value) {
  let selector = `input[name=${alliance}${name}]`;
  if (value !== undefined) {
    selector += `[value=${value}]`;
  }
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

const cloneTemplateElement = function (id) {
  return $($(`#${id}`).prop("content").firstElementChild.cloneNode(true));
};

const formatSummaryValue = function (field, value) {
  if (RANKING_POINT_SUMMARY_FIELDS.includes(field)) {
    return value ? '<span class="score-summary-rp text-success">&#x2611;</span>' :
      '<span class="score-summary-rp text-danger">&#x2612;</span>';
  }
  return value;
};

const parseFormInt = function (value) {
  const parsed = parseInt(value, 10);
  if (isNaN(parsed)) {
    return 0;
  }
  return parsed;
};
