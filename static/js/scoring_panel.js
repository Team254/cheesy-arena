// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let nearSide;
let committed = false;

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
// 新增一個輔助函式來處理 UI 顏色與文字
const updateHubUI = function(redActive, blueActive) {
    const indicator = $("#hub-status-indicator");
    const card = $("#hub-status-card");

    if (redActive && blueActive) {
        indicator.text("BOTH ACTIVE").css({"background-color": "#198754", "color": "white"}); // 綠色
        card.css("border-color", "#198754");
    } else if (redActive) {
        indicator.text("RED ACTIVE").css({"background-color": "#dc3545", "color": "white"}); // 紅色
        card.css("border-color", "#dc3545");
    } else if (blueActive) {
        indicator.text("BLUE ACTIVE").css({"background-color": "#0d6efd", "color": "white"}); // 藍色
        card.css("border-color", "#0d6efd");
    } else {
        indicator.text("HUB INACTIVE").css({"background-color": "#6c757d", "color": "white"}); // 灰色
        card.css("border-color", "#6c757d");
    }
};

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
  $(".scoring-teleop-button").prop('disabled', !(inTeleop && scoringAvailable));
  $("#commit").prop('disabled', !commitAvailable);
  $("#edit-auto").prop('disabled', !(inTeleop && scoringAvailable));
  $(".container").attr("data-scoring-auto", (!inTeleop || editingAuto) && scoringAvailable);
  $(".container").attr("data-in-teleop", inTeleop && scoringAvailable);
  $("#edit-auto").text(editingAuto ? "Save Auto" : "Edit Auto");
}

const endgameStatusNames = [
  "None",    // 0
  "Level 1", // 1
  "Level 2", // 2
  "Level 3", // 3
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  
  // 取得後端傳來的 2026 核心分數結構
  const score = realtimeScore.Score;
  if (!score) return;

  // 這裡我們需要同時看 data.Red 和 data.Blue
    const redActive = data.Red.Score.HubActive;
    const blueActive = data.Blue.Score.HubActive;
    updateHubUI(redActive, blueActive);

  // 1. 同步 Fuel (燃料) 計數
  $("#auto_fuel_count").text(score.AutoFuelCount || 0);
  $("#teleop_fuel_count").text(score.TeleopFuelCount || 0);

  // 2. 同步 Auto Tower (自動階段塔) 勾選狀態
  // 假設後端 AutoTowerLevel1 是一個包含 3 個布林值的陣列
  if (score.AutoTowerLevel1) {
    for (let i = 0; i < 3; i++) {
      $(`#auto_tower_${i}`).prop("checked", score.AutoTowerLevel1[i]);
    }
  }

  // 3. 同步 Endgame / Climb (攀爬) 狀態
  // 這是解決你 "Level 1 無法點擊" 的關鍵
  if (score.EndgameStatuses) {
    for (let i = 0; i < 3; i++) {
      const status = score.EndgameStatuses[i]; // 0=None, 1=Lvl1, 2=Lvl2, 3=Lvl3
      
      // 更新 HTML 上的 Radio 按鈕
      // 注意：你的 HTML name 是 climb_0, climb_1, climb_2
      $(`input[name='climb_${i}']`).filter(`[value='${status}']`).prop("checked", true);
      
      // 如果你有顯示文字標籤，也可以更新它
      $(`#endgame-status-${i+1} > .team-text`).text(endgameStatusNames[status] || "None");
    }
  }

  // 4. 同步 Fouls (犯規)
  const redFouls = data.Red.Score.Fouls || [];
  const blueFouls = data.Blue.Score.Fouls || [];
  
  $(`#foul-blue-minor .fouls-global`).text(blueFouls.filter(f => !f.IsMajor).length);
  $(`#foul-blue-major .fouls-global`).text(blueFouls.filter(f => f.IsMajor).length);
  $(`#foul-red-minor .fouls-global`).text(redFouls.filter(f => !f.IsMajor).length);
  $(`#foul-red-major .fouls-global`).text(redFouls.filter(f => f.IsMajor).length);
};

// Websocket message senders for various buttons
const handleCounterClick = function (command, adjustment) {
  websocket.send(command, {
    Adjustment: adjustment,
    Current: true,
    Autonomous: !inTeleop || editingAuto,
    NearSide: nearSide
  });
}
const handleLeaveClick = function (teamPosition) {
  websocket.send("leave", {TeamPosition: teamPosition});
}
const handleEndgameClick = function (teamPosition, endgameStatus) {
  websocket.send("endgame", {TeamPosition: teamPosition, EndgameStatus: endgameStatus});
}


// 對應 HTML: updateFuel(isAuto, delta)
const updateFuel = function (isAuto, delta) {
  websocket.send("fuel", {
    Adjustment: delta,
    Autonomous: isAuto
  });
};

// 對應 HTML: updateAutoTower(robotIdx, checked)
const updateAutoTower = function (robotIdx, checked) {
  websocket.send("auto_tower", {
    RobotIndex: robotIdx,
    Adjustment: checked ? 1 : 0
  });
};

// 對應 HTML: updateClimb(robotIdx, level)
// 這是解決 Level 1 點擊沒反應的最關鍵修正
const updateClimb = function (robotIdx, level) {
  console.log("Sending climb command:", robotIdx, level);
  websocket.send("climb", {
    RobotIndex: parseInt(robotIdx),
    Level: parseInt(level)
  });
};

const handleReefClick = function (reefPosition, reefLevel) {
  websocket.send("reef", {
    ReefPosition: reefPosition,
    ReefLevel: reefLevel,
    Current: !editingAuto,
    Autonomous: !inTeleop || editingAuto,
    NearSide: nearSide
  });
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
