// Copyright 2026 Team 254. All Rights Reserved.
// 2026 REBUILT Version

const scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
const allianceResults = {};
let matchResult;

// 攔截表單提交，將資料轉為 JSON 以利伺服器解析
$("form").submit(function () {
  updateResults("red");
  updateResults("blue");

  matchResult.RedScore = allianceResults["red"].score;
  matchResult.BlueScore = allianceResults["blue"].score;
  matchResult.RedCards = allianceResults["red"].cards;
  matchResult.BlueCards = allianceResults["blue"].cards;
  const matchResultJson = JSON.stringify(matchResult);

  // 注入隱藏輸入項
  $("<input />").attr("type", "hidden").attr("name", "matchResultJson").attr("value", matchResultJson).appendTo("form");

  return true;
});

// 渲染特定聯盟的結果到頁面
const renderResults = function (alliance) {
  const result = allianceResults[alliance];
  const scoreContent = scoreTemplate(result);
  $(`#${alliance}Score`).html(scoreContent);

  // 1. Fuel 數量 (Auto/Teleop)
  getInputElement(alliance, "AutoFuelCount").val(result.score.AutoFuelCount || 0);
  getInputElement(alliance, "TeleopFuelCount").val(result.score.TeleopFuelCount || 0);

  // 2. 處理 3 個隊伍的狀態
  // 注意：result.score 陣列索引為 0, 1, 2；但 HTML 欄位名稱使用 1, 2, 3 (對應 Payload)
  for (let i = 0; i < 3; i++) {
    const htmlIdx = i + 1; 

    // A. 機器人是否被 Bypassed
    if (result.score.RobotsBypassed) {
      getInputElement(alliance, `RobotsBypassed${htmlIdx}`).prop("checked", result.score.RobotsBypassed[i]);
    }

    // B. Autonomous Tower Level 1
    if (result.score.AutoTowerLevel1) {
      getInputElement(alliance, `AutoTowerLevel1${htmlIdx}`).prop("checked", result.score.AutoTowerLevel1[i]);
    }

    // C. Endgame Status (支援含 alliance 或不含 alliance 的欄位名)
    if (result.score.EndgameStatuses) {
      let el = getInputElement(alliance, `EndgameStatuses${htmlIdx}`, result.score.EndgameStatuses[i]);
      if (el.length === 0) { // 兼容 Payload 顯示的 EndgameStatuses1 (無 alliance 前綴)
         $(`input[name=EndgameStatuses${htmlIdx}][value=${result.score.EndgameStatuses[i]}]`).prop("checked", true);
      } else {
         el.prop("checked", true);
      }
    }
  }

  // 3. 犯規列表
  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function (k, v) {
      getInputElement(alliance, `Foul${k}IsMajor`).prop("checked", v.IsMajor);
      getInputElement(alliance, `Foul${k}Team`, v.TeamId).prop("checked", true);
      getSelectElement(alliance, `Foul${k}RuleId`).val(v.RuleId);
    });
  }

  // 4. 卡片 (Cards)
  if (result.cards != null) {
    $.each([result.team1, result.team2, result.team3], function (i, team) {
      getInputElement(alliance, `Team${team}Card`, result.cards[team]).prop("checked", true);
    });
  }
};

// 從表單更新緩存的 JSON 資料結構
const updateResults = function (alliance) {
  const result = allianceResults[alliance];
  const formData = {};
  $.each($("form").serializeArray(), function (k, v) {
    formData[v.name] = v.value;
  });

  // 初始化陣列結構 (Go 端預期長度 3)
  result.score.RobotsBypassed = [false, false, false];
  result.score.AutoTowerLevel1 = [false, false, false];
  result.score.EndgameStatuses = [0, 0, 0];

  // 讀取 Fuel Count
  result.score.AutoFuelCount = parseInt(formData[`${alliance}AutoFuelCount`]) || 0;
  result.score.TeleopFuelCount = parseInt(formData[`${alliance}TeleopFuelCount`]) || 0;

  for (let i = 0; i < 3; i++) {
    const htmlIdx = i + 1; // 根據 Payload，HTML 名稱為 redRobotsBypassed1...3
    
    // 抓取 Bypassed
    result.score.RobotsBypassed[i] = formData[`${alliance}RobotsBypassed${htmlIdx}`] === "on";
    
    // 抓取 AutoTower
    result.score.AutoTowerLevel1[i] = formData[`${alliance}AutoTowerLevel1${htmlIdx}`] === "on";
    
    // 抓取 Endgame (優先抓取帶聯盟前綴的，若無則抓取不帶前綴的)
    let endgameVal = formData[`${alliance}EndgameStatuses${htmlIdx}`] || formData[`EndgameStatuses${htmlIdx}`];
    result.score.EndgameStatuses[i] = parseInt(endgameVal) || 0;
  }

  // 處理犯規
  result.score.Fouls = [];
  for (let i = 0; formData[`${alliance}Foul${i}Index`]; i++) {
    const prefix = `${alliance}Foul${i}`;
    result.score.Fouls.push({
      IsMajor: formData[`${prefix}IsMajor`] === "on",
      TeamId: parseInt(formData[`${prefix}Team`]) || 0,
      RuleId: parseInt(formData[`${prefix}RuleId`]) || 0,
    });
  }

  // 處理卡片
  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function (i, team) {
    result.cards[team] = formData[`${alliance}Team${team}Card`] || "";
  });
};

const addFoul = function (alliance) {
  updateResults(alliance);
  allianceResults[alliance].score.Fouls.push({IsMajor: false, TeamId: 0, RuleId: 0});
  renderResults(alliance);
};

const deleteFoul = function (alliance, index) {
  updateResults(alliance);
  allianceResults[alliance].score.Fouls.splice(index, 1);
  renderResults(alliance);
};

const getInputElement = function (alliance, name, value) {
  let selector = `input[name=${alliance}${name}]`;
  if (value !== undefined) selector += `[value=${value}]`;
  return $(selector);
};

const getSelectElement = function (alliance, name) {
  return $(`select[name=${alliance}${name}]`);
};