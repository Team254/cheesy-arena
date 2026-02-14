// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let nearSide;
let committed = false;
let currentRealtimeScore = null; // Store current score for toggle operations

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;
// True when teleop-only scoring controls should be available
let inTeleop = false;
// True when post-auto and in edit auto mode
let editingAuto = false;

let localFoulCounts = {
  "red-minor": 0,
  "blue-minor": 0,
  "red-major": 0,
  "blue-major": 0,
}

// Handle controls to open/close the endgame dialog
const endgameDialog = $("#endgame-dialog")[0];
const showEndgameDialog = function () {
  endgameDialog.showModal();
}
const closeEndgameDialog = function () {
  endgameDialog.close();
}
const closeEndgameDialogIfOutside = function (event) {
  if (event.target === endgameDialog) {
    closeEndgameDialog();
  }
}

const foulsDialog = $("#fouls-dialog")[0];
const showFoulsDialog = function () {
  foulsDialog.showModal();
}
const closeFoulsDialog = function () {
  foulsDialog.close();
}
const closeFoulsDialogIfOutside = function (event) {
  if (event.target === foulsDialog) {
    closeFoulsDialog();
  }
}

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
  if (alliance === "red") {
    $(".team-1 .team-num").text(data.Match.Red1);
    $(".team-2 .team-num").text(data.Match.Red2);
    $(".team-3 .team-num").text(data.Match.Red3);
  } else {
    $(".team-1 .team-num").text(data.Match.Blue1);
    $(".team-2 .team-num").text(data.Match.Blue2);
    $(".team-3 .team-num").text(data.Match.Blue3);
  }
};

const renderLocalFoulCounts = function () {
  for (const foulType in localFoulCounts) {
    const count = localFoulCounts[foulType];
    $(`#foul-${foulType} .fouls-local`).text(count);
  }
}

const resetFoulCounts = function () {
  localFoulCounts["red-minor"] = 0;
  localFoulCounts["blue-minor"] = 0;
  localFoulCounts["red-major"] = 0;
  localFoulCounts["blue-major"] = 0;
  renderLocalFoulCounts();
}

const addFoul = function (alliance, isMajor) {
  const foulType = `${alliance}-${isMajor ? "major" : "minor"}`;
  localFoulCounts[foulType] += 1;
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
  renderLocalFoulCounts();
}

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = false;
      editingAuto = false;
      committed = false;
      break;
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = true;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
        inTeleop = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      inTeleop = false;
      editingAuto = false;
      committed = false;
      resetFoulCounts();
  }
  updateUIMode();
};

// Switch in and out of autonomous editing mode
const toggleEditAuto = function () {
  editingAuto = !editingAuto;
  updateUIMode();
}

// Clear any local ephemeral state that is not maintained by the server
const resetLocalState = function () {
  committed = false;
  editingAuto = false;
  updateUIMode();
}

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $(".scoring-auto-button").prop('disabled', !scoringAvailable); // Auto climb always available when scoring
  $(".scoring-teleop-button").prop('disabled', !(inTeleop && scoringAvailable));
  $("#commit").prop('disabled', !commitAvailable);
  $("#edit-auto").prop('disabled', !(inTeleop && scoringAvailable));
  $(".container").attr("data-scoring-auto", (!inTeleop || editingAuto) && scoringAvailable);
  $(".container").attr("data-in-teleop", inTeleop && scoringAvailable);
  $("#edit-auto").text(editingAuto ? "Save Auto" : "Edit Auto");
}

const climbLevelNames = [
  "None",
  "L1",
  "L2",
  "L3",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  currentRealtimeScore = data; // Store for toggle operations

  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;

    // Display auto climb status on auto button (top left)
    const autoClimb = climbLevelNames[score.AutoClimbStatuses[i]];
    $(`#auto-status-${i1} > .team-text`).text(autoClimb);
    $(`#auto-status-${i1}`).attr("data-selected", score.AutoClimbStatuses[i] != 0);

    // Display teleop climb status on endgame button (top right)
    const teleopClimb = climbLevelNames[score.TeleopClimbStatuses[i]];
    $(`#endgame-status-${i1} > .team-text`).text(teleopClimb);
    $(`#endgame-status-${i1}`).attr("data-selected", score.TeleopClimbStatuses[i] != 0);

    // Update teleop climb button selection in modal
    for (let j = 0; j <= 3; j++) {
      $(`#endgame-input-${i1} .teleop-climb-${j}`).attr("data-selected", j == score.TeleopClimbStatuses[i]);
    }
  }

  // Update FUEL counters (only shown when PLC is not enabled)
  $(`#autoFuel .counter-value`).text(score.AutoFuel);
  $(`#activeFuel .counter-value`).text(score.ActiveFuel);
  $(`#inactiveFuel .counter-value`).text(score.InactiveFuel);

  redFouls = data.Red.Score.Fouls || [];
  blueFouls = data.Blue.Score.Fouls || [];
  $(`#foul-blue-minor .fouls-global`).text(blueFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-blue-major .fouls-global`).text(blueFouls.filter(foul => foul.IsMajor).length)
  $(`#foul-red-minor .fouls-global`).text(redFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-red-major .fouls-global`).text(redFouls.filter(foul => foul.IsMajor).length)
};

// Websocket message senders for various buttons
const handleCounterClick = function (command, adjustment) {
  websocket.send(command, {
    Adjustment: adjustment,
    Current: true,
    Autonomous: command === "autoFuel" ? true : (!inTeleop || editingAuto),
    NearSide: nearSide
  });
}
// Toggle auto climb between None (0) and L1 (1)
const handleAutoClimbClick = function (teamPosition) {
  // Get current state from the score
  let currentScore;
  if (alliance === "red") {
    currentScore = currentRealtimeScore.Red.Score;
  } else {
    currentScore = currentRealtimeScore.Blue.Score;
  }

  const currentLevel = currentScore.AutoClimbStatuses[teamPosition - 1];
  const newLevel = currentLevel === 0 ? 1 : 0; // Toggle between None and L1

  websocket.send("autoClimb", {TeamPosition: teamPosition, EndgameStatus: newLevel});
}

const handleTeleopClimbClick = function (teamPosition, climbLevel) {
  websocket.send("teleopClimb", {TeamPosition: teamPosition, EndgameStatus: climbLevel});
}

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  inTeleop = false;
  editingAuto = false;
  updateUIMode();
};

$(function () {
  position = window.location.href.split("/").slice(-1)[0];
  [alliance, side] = position.split("_");
  $(".container").attr("data-alliance", alliance);
  nearSide = side === "near";
  resetLocalState();

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + position + "/websocket", {
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    resetLocalState: function (event) {
      resetLocalState();
    },
  });
});
