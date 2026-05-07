// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
let redFoulsHashCode = 0;
let blueFoulsHashCode = 0;

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

// Cycles through no card, yellow card, and red card.
var cycleCard = function (cardButton) {
  var newCard = "";
  if ($(cardButton).attr("data-card") === "") {
    newCard = "yellow";
  } else if ($(cardButton).attr("data-card") === "yellow") {
    newCard = "red";
  }
  websocket.send(
    "card",
    {Alliance: $(cardButton).attr("data-alliance"), TeamId: parseInt($(cardButton).attr("data-team")), Card: newCard}
  );
  $(cardButton).attr("data-card", newCard);
};

// Sends a websocket message to signal to the volunteers that they may enter the field.
var signalVolunteers = function () {
  websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
var signalReset = function () {
  websocket.send("signalReset");
};

// Signals the scorekeeper that foul entry is complete for this match.
var commitMatch = function () {
  websocket.send("commitMatch");
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

  $("#redScoreSummary .team-1").text(data.Teams["R1"].Id);
  $("#redScoreSummary .team-2").text(data.Teams["R2"].Id);
  $("#redScoreSummary .team-3").text(data.Teams["R3"].Id);
  $("#blueScoreSummary .team-1").text(data.Teams["B1"].Id);
  $("#blueScoreSummary .team-2").text(data.Teams["B2"].Id);
  $("#blueScoreSummary .team-3").text(data.Teams["B3"].Id);
};

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  $(".control-button").attr("data-enabled", matchStates[data.MatchState] === "POST_MATCH");
};

const endgameStatusNames = [
  "None",
  "Level 1",
  "Level 2",
  "Level 3",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  if (!data) return;

  // 1. 取得紅藍兩隊的分數物件 (增加層級容錯)
  const redRealtime = data.RedRealtimeScore || data.Red || {};
  const blueRealtime = data.BlueRealtimeScore || data.Blue || {};
  
  const redScore = redRealtime.CurrentScore || redRealtime.Score || {};
  const blueScore = blueRealtime.CurrentScore || blueRealtime.Score || {};

  // --- 燃料 (Fuel) 更新邏輯 ---
  // 更新紅隊燃料數據
  $("#redScoreSummary .fuel-teleop").text(redScore.TeleopFuelCount || 0);
  $("#redScoreSummary .fuel-auto").text(redScore.AutoFuelCount || 0);

  // 更新藍隊燃料數據
  $("#blueScoreSummary .fuel-teleop").text(blueScore.TeleopFuelCount || 0);
  $("#blueScoreSummary .fuel-auto").text(blueScore.AutoFuelCount || 0);

  // 2. 檢查 Fouls 是否有更新
  const newRedFoulsHashCode = hashObject(redScore.Fouls || []);
  const newBlueFoulsHashCode = hashObject(blueScore.Fouls || []);

  if (newRedFoulsHashCode !== redFoulsHashCode || newBlueFoulsHashCode !== blueFoulsHashCode) {
    redFoulsHashCode = newRedFoulsHashCode;
    blueFoulsHashCode = newBlueFoulsHashCode;
    
    fetch("/panels/referee/foul_list")
      .then(response => response.text())
      .then(html => {
        // 放寬檢查：只要不是完整的 HTML 頁面就插入
        if (html.indexOf("<!DOCTYPE") === -1 && html.indexOf("<html") === -1) {
            $("#foulList").html(html);
        } else {
            console.error("Foul list error: Received full page instead of snippet.");
        }
      });
    }
  // --- 3 & 4. AutoTowerLevel1 狀態更新 (優化版) ---
  const updateAutoTowerUI = (allianceScore, containerId) => {
      if (allianceScore.AutoTowerLevel1) {
          for (let i = 0; i < 3; i++) {
              const isScored = allianceScore.AutoTowerLevel1[i];
              // 改用 containerId 下的特定 class 選擇器
              const element = $(`${containerId} .team-${i + 1}-tower`);
              element.text(isScored ? "✅" : "❌");
              element.attr("data-active", isScored);
              // 可選：直接改變顏色
              element.css("color", isScored ? "#28a745" : "#dc3545");
          }
      }
  };

  // 呼叫更新
  updateAutoTowerUI(redScore, "#redScoreSummary");
  updateAutoTowerUI(blueScore, "#blueScoreSummary");

  // --- 5. Endgame (Climb) 狀態更新 ---
  // 對應您 HTML 中的 .team-X-endgame
  const updateEndgameUI = (allianceScore, containerId) => {
    if (allianceScore.EndgameStatuses) {
      for (let i = 0; i < 3; i++) {
        const status = allianceScore.EndgameStatuses[i]; // 0:None, 1:Lvl1, 2:Lvl2, 3:Lvl3
        const statusText = endgameStatusNames[status] || "None";
        $(`${containerId} .team-${i + 1}-endgame`).text(statusText);
        
        // 可選：根據狀態改變顏色 (例如 None 為灰色，Level 3 為綠色)
        $(`${containerId} .team-${i + 1}-endgame`).attr("data-status", status);
      }
    }
  };
  updateEndgameUI(redScore, "#redScoreSummary");
  updateEndgameUI(blueScore, "#blueScoreSummary");

  // --- RP Status 更新邏輯 ---
  const updateRPUI = (allianceData, containerId) => {
      // 關鍵修正：RP 狀態通常在 Summary 欄位下，而不是 Score 欄位下
      const summary = allianceData.Summary || allianceData.ScoreSummary || allianceData;

      // 更新 Energized RP 狀態
      const EnergizedElement = $(`${containerId} .Energized-status`);
      const isEnergized = summary.EnergizedRankingPoint || summary.energizedRankingPoint || false;
      EnergizedElement.text(isEnergized ? "Yes" : "No");
      EnergizedElement.css("color", isEnergized ? "#28a745" : "inherit");

      // 更新 Supercharged RP 狀態
      const superchargedElement = $(`${containerId} .supercharged-status`);
      const isSupercharged = summary.SuperchargedRankingPoint || summary.superchargedRankingPoint || false;
      superchargedElement.text(isSupercharged ? "Yes" : "No");
      superchargedElement.css("color", isSupercharged ? "#28a745" : "inherit");

      // 更新 Traversal RP 狀態
      const traversalElement = $(`${containerId} .traversal-status`);
      const isTraversal = summary.TraversalRankingPoint || summary.traversalRankingPoint || false;
      traversalElement.text(isTraversal ? "Yes" : "No");
      traversalElement.css("color", isTraversal ? "#28a745" : "inherit");
  };

  // 在 handleRealtimeScore 呼叫時，傳入完整的 redRealtime 而不是只有 redScore
  updateRPUI(redRealtime, "#redScoreSummary");
  updateRPUI(blueRealtime, "#blueScoreSummary");
}

// Handles a websocket message to update the scoring commit status.
const handleScoringStatus = function (data) {
  if (data.RefereeScoreReady) {
    $("#commitButton").attr("data-enabled", false);
  }
  updateScoreStatus(data, "red_near", "#redNearScoreStatus", "Red Near");
  updateScoreStatus(data, "red_far", "#redFarScoreStatus", "Red Far");
  updateScoreStatus(data, "blue_near", "#blueNearScoreStatus", "Blue Near");
  updateScoreStatus(data, "blue_far", "#blueFarScoreStatus", "Blue Far");
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
  const cardButton = $(`#${alliance}Team${position}Card`);
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
  });
});
