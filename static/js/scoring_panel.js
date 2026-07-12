// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let committed = false;

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;

let localFoulCounts = {
  "red-minor": 0,
  "blue-minor": 0,
  "red-major": 0,
  "blue-major": 0,
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
  } else {
    $(".team-1 .team-num").text(data.Match.Blue1);
    $(".team-2 .team-num").text(data.Match.Blue2);
  }
};

const renderLocalFoulCounts = function () {
  for (const foulType in localFoulCounts) {
    const count = localFoulCounts[foulType];
    $(`#foul-${foulType} .fouls-local`).text(count)
  }
}

const renderGlobalFoulCounts = function (redFouls, blueFouls) {
  $(`#foul-blue-minor .fouls-global`).text(blueFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-blue-major .fouls-global`).text(blueFouls.filter(foul => foul.IsMajor).length)
  $(`#foul-red-minor .fouls-global`).text(redFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-red-major .fouls-global`).text(redFouls.filter(foul => foul.IsMajor).length)
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
  renderLocalFoulCounts();
  websocket.send("addFoul", { Alliance: alliance, IsMajor: isMajor });
}

// Expose functions used by inline onclick handlers so templates can call them.
window.addFoul = addFoul;

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      committed = false;
      resetFoulCounts();
  }
  updateUIMode();
};

// Clear any local ephemeral state that is not maintained by the server
const resetLocalState = function () {
  committed = false;
  updateUIMode();
}

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $(".scoring-tower-button").prop('disabled', !scoringAvailable);
  $("#commit").prop('disabled', !commitAvailable);
}

const endgameStatusNames = [
  "None",
  "Park",
  "Complete",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  console.debug("realtimeScore received", data);
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 2; i++) {
    const i1 = i + 1;
    for (let j = 0; j < endgameStatusNames.length; j++) {
      $(`#auto-input-${i1} .tower-${j}`).attr("data-selected", j == score.AutoTowerStatuses[i]);
      $(`#endgame-input-${i1} .tower-${j}`).attr("data-selected", j == score.EndgameTowerStatuses[i]);
    }
  }

  const redFouls = data.Red.Score.Fouls || [];
  const blueFouls = data.Blue.Score.Fouls || [];
  renderGlobalFoulCounts(redFouls, blueFouls);
  // Render Auto and Manual totals for this alliance.
  let autoTotal = 0;
  if (score.Hub && score.Hub.ShiftCounts) {
    const shiftCounts = score.Hub.ShiftCounts;
    for (let i = 0; i < shiftCounts.length; i++) {
      autoTotal += shiftCounts[i] || 0;
    }
  }
  let manualTotal = 0;
  if (score.Hub && typeof score.Hub.ManualTotal === 'number') {
    manualTotal = score.Hub.ManualTotal;
  } else if (score.Hub && score.Hub.ManualShiftCounts) {
    const manualCounts = score.Hub.ManualShiftCounts;
    for (let i = 0; i < manualCounts.length; i++) {
      manualTotal += manualCounts[i] || 0;
    }
  }
  $(".auto-fuel-total").text(`Auto: ${autoTotal}`);
  $(".manual-fuel-total").text(`Manual: ${manualTotal}`);
};

// Websocket message senders for various buttons
const handleAutoTowerClick = function (teamPosition, autoTowerStatus) {
  websocket.send("autoPark", { TeamPosition: teamPosition, AutoParkStatus: autoTowerStatus });
}
const handleEndgameClick = function (teamPosition, endgameTowerStatus) {
  websocket.send("endgamePark", { TeamPosition: teamPosition, EndgameParkStatus: endgameTowerStatus });
}

// Send manual fuel delta (positive or negative) from the UI.
const addManualFuelDelta = function (delta) {
  websocket.send("manualFuel", { Count: delta });
}

// Expose handlers for inline onclicks
window.handleAutoTowerClick = handleAutoTowerClick;
window.handleEndgameClick = handleEndgameClick;
window.addManualFuelDelta = addManualFuelDelta;

// Send a manual fuel delta for a specific shift index (0=Auto, 6=Endgame, etc.).
const addManualFuelShift = function (shiftIndex, delta) {
  console.debug("sending manualFuel", { Count: delta, Shift: shiftIndex });
  websocket.send("manualFuel", { Count: delta, Shift: shiftIndex });
}
window.addManualFuelShift = addManualFuelShift;

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  updateUIMode();
};

window.commitMatchScore = commitMatchScore;

$(function () {
  position = window.location.href.split("/").slice(-1)[0];
  alliance = position;
  $(".container").attr("data-alliance", alliance);
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
