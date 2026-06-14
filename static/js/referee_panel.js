// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
let redFoulsHashCode = 0;
let blueFoulsHashCode = 0;
let scoreIsReady = false;
let isPostMatch = false;

// Sends the foul to the server to add it to the list.
const addFoul = function (alliance, isMajor) {
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
}

// Toggles the foul type between minor and major.
const toggleFoulType = function (alliance, index) {
  websocket.send("toggleFoulType", {Alliance: alliance, Index: index});
}

// Updates the team that the foul is attributed to.
const updateFoulTeam = function (alliance, index, teamId) {
  websocket.send("updateFoulTeam", {Alliance: alliance, Index: index, TeamId: teamId});
}

// Updates the rule that the foul is for.
const updateFoulRule = function (alliance, index, ruleId) {
  websocket.send("updateFoulRule", {Alliance: alliance, Index: index, RuleId: ruleId});
}

// Removes the foul with the given parameters from the list.
var deleteFoul = function (alliance, index) {
  websocket.send("deleteFoul", {Alliance: alliance, Index: index});
};

// Cycles through the card options for the selected team.
var cycleCard = function (cardButton) {
  if(isPostMatch) {
    // Cycle card.
    const currentCard = $(cardButton).attr("data-card");
    const hasOldYellowCard = $(cardButton).attr("data-old-yellow-card") === "true";
    let newCard = "";
    if (currentCard === "" && hasOldYellowCard) {
      newCard = "red";
    } else if (currentCard === "") {
      newCard = "yellow";
    } else if (currentCard === "yellow") {
      newCard = "red";
    }
    websocket.send(
      "card",
      {Alliance: $(cardButton).attr("data-alliance"), TeamId: parseInt($(cardButton).attr("data-team")), Card: newCard}
    );
    $(cardButton).attr("data-card", newCard);
    return;
  }

  // Toggle bypass.
  const isDisabled = $(cardButton).hasClass("bypassed-status");
  const team = $(cardButton).attr("data-team");
  $("#confirmBypassTitle").text(`${isDisabled ? "Enable" : "Disable"} ${team}?`);
  $("#confirmBypassAction").text(isDisabled ? "Enable" : "Disable")
  $("#confirmBypass").attr("data-station", $(cardButton).attr("data-station")?.toUpperCase());

  if(team === "0") {
    toggleBypass();
  } else {
    $("#confirmBypass").modal("show");
  }
};

const toggleBypass = function() {
  const station = $("#confirmBypass").attr("data-station");
  websocket.send("toggleBypass", station);
}

// Sends a websocket message to signal to the volunteers that they may enter the field.
var signalVolunteers = function () {
  websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
var signalReset = function () {
  websocket.send("signalReset");
};

// Shows confirmation modal if not all scores are ready, otherwise directly commits and posts.
var confirmCommit = function () {
  if (scoreIsReady) {
    commitAndPost();
    return;
  }

  $("#confirmCommit").modal("show");
};

// Commits the score and posts results to the audience.
var commitAndPost = function () {
  websocket.send("commitAndPost");
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);

  setTeamCard("red", 1, data.Teams["R1"]);
  setTeamCard("red", 2, data.Teams["R2"]);
  setTeamCard("red", 3, data.Teams["R3"]);
  setTeamCard("blue", 1, data.Teams["B1"]);
  setTeamCard("blue", 2, data.Teams["B2"]);
  setTeamCard("blue", 3, data.Teams["B3"]);

  $("#redScoreSummary .team-1").text(data.Teams["R1"]?.Id || "");
  $("#redScoreSummary .team-2").text(data.Teams["R2"]?.Id || "");
  $("#redScoreSummary .team-3").text(data.Teams["R3"]?.Id || "");
  $("#blueScoreSummary .team-1").text(data.Teams["B1"]?.Id || "");
  $("#blueScoreSummary .team-2").text(data.Teams["B2"]?.Id || "");
  $("#blueScoreSummary .team-3").text(data.Teams["B3"]?.Id || "");
};

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  isPostMatch = matchStates[data.MatchState] === "POST_MATCH";
  $(".control-button").attr("data-enabled", isPostMatch);

  let title = "Red/Yellow Cards";
  if(!isPostMatch) {
    title = matchStates[data.MatchState] === "PRE_MATCH" ? "Bypass" : "Disable";
  }

  $("#teamTitle").text(title)
};

const towerStatusNames = [
  "None",
  "Level 1",
  "Level 2",
  "Level 3",
];

const setTowerStatus = function (selector, status) {
  $(selector).text(towerStatusNames[status]);
  $(selector).attr("data-status", status);
};

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  for (const [teamId, card] of Object.entries(Object.assign(data.RedCards, data.BlueCards))) {
    $(`[data-team="${teamId}"]`).attr("data-card", card);
  }

  const newRedFoulsHashCode = hashObject(data.Red.Score.Fouls);
  const newBlueFoulsHashCode = hashObject(data.Blue.Score.Fouls);
  if (newRedFoulsHashCode !== redFoulsHashCode || newBlueFoulsHashCode !== blueFoulsHashCode) {
    redFoulsHashCode = newRedFoulsHashCode;
    blueFoulsHashCode = newBlueFoulsHashCode;
    fetch("/panels/referee/foul_list")
      .then(response => response.text())
      .then(svg => $("#foulList").html(svg));
  }

  for (alliance of ["red", "blue"]) {
    let score;
    if (alliance === "red") {
      score = data.Red.Score;
    } else {
      score = data.Blue.Score;
    }

    let scoreRoot = `${alliance}ScoreSummary`;
    setTowerStatus(`#${scoreRoot} .team-1-auto-tower`, score.AutoTowerStatuses[0]);
    setTowerStatus(`#${scoreRoot} .team-2-auto-tower`, score.AutoTowerStatuses[1]);
    setTowerStatus(`#${scoreRoot} .team-3-auto-tower`, score.AutoTowerStatuses[2]);
    setTowerStatus(`#${scoreRoot} .team-1-endgame-tower`, score.EndgameTowerStatuses[0]);
    setTowerStatus(`#${scoreRoot} .team-2-endgame-tower`, score.EndgameTowerStatuses[1]);
    setTowerStatus(`#${scoreRoot} .team-3-endgame-tower`, score.EndgameTowerStatuses[2]);
  }
}

// Handles a websocket message to update the scoring commit status.
const handleScoringStatus = function (data) {
  if (data.RefereeScoreReady) {
    $("#commitButton").attr("data-enabled", false);
  }
  updateScoreStatus(data, "red", "#redScoreStatus", "Red");
  updateScoreStatus(data, "blue", "#blueScoreStatus", "Blue");

  scoreIsReady = Object.values(data.PositionStatuses).every(status => status.Ready);

  // Make the button visually distinct if not all refs have committed.
  // HR can still press the button with confirm modal.
  if (scoreIsReady) {
    $("#commitButton").removeClass("disabled");
  } else {
    $("#commitButton").addClass("disabled");
  }
}

const handleArenaStatus = function (data) {
  setTeamBypassedStatus("red1", data.AllianceStations["R1"]?.Bypass);
  setTeamBypassedStatus("red2", data.AllianceStations["R2"]?.Bypass);
  setTeamBypassedStatus("red3", data.AllianceStations["R3"]?.Bypass);
  setTeamBypassedStatus("blue1", data.AllianceStations["B1"]?.Bypass);
  setTeamBypassedStatus("blue2", data.AllianceStations["B2"]?.Bypass);
  setTeamBypassedStatus("blue3", data.AllianceStations["B3"]?.Bypass);
};

const setTeamBypassedStatus = function (station, bypassed) {
  const cardButton = $(`#${station}Card`);
  cardButton.toggleClass("bypassed-status", bypassed && !isPostMatch);
}

// Helper function to update a badge that shows scoring panel commit status.
const updateScoreStatus = function (data, position, element, displayName) {
  const status = data.PositionStatuses[position];
  $(element).text(`${displayName} ${status.NumPanelsReady}/${status.NumPanels}`);
  $(element).attr("data-present", status.NumPanels > 0);
  $(element).attr("data-ready", status.Ready);
};

// Populates the red/yellow card button for a given team.
const setTeamCard = function (alliance, position, team) {
  const cardButton = $(`#${alliance}${position}Card`);
  if (team === null) {
    cardButton.text(0);
    cardButton.attr("data-team", 0)
    cardButton.attr("data-old-yellow-card", "");
  } else {
    cardButton.text(team.Id);
    cardButton.attr("data-team", team.Id)
    cardButton.attr("data-old-yellow-card", team.YellowCard);
  }
  cardButton.attr("data-card", "");
}

// Produces a hash code of the given object for use in equality comparisons.
const hashObject = function (object) {
  const s = JSON.stringify(object);
  let h = 0;
  for (let i = 0; i < s.length; i++) {
    h = Math.imul(31, h) + s.charCodeAt(i) | 0;
  }
  return h;
}

$(function () {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  $(".headRef-dependent").attr("data-hr", urlParams.get("hr"));

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/referee/websocket", {
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scoringStatus: function (event) {
      handleScoringStatus(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    },
  });
});
