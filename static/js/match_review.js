const scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
const allianceResults = {};
let matchResult;

// 攔截表單提交，將資料轉為 JSON
$("form").submit(function () {
  updateResults("red");
  updateResults("blue");

  matchResult.RedScore = allianceResults["red"].score;
  matchResult.BlueScore = allianceResults["blue"].score;
  matchResult.RedCards = allianceResults["red"].cards;
  matchResult.BlueCards = allianceResults["blue"].cards;
  const matchResultJson = JSON.stringify(matchResult);

  $("<input />").attr("type", "hidden").attr("name", "matchResultJson").attr("value", matchResultJson).appendTo("form");
  return true;
});

// 渲染結果到頁面
const renderResults = function (alliance) {
  const result = allianceResults[alliance];
  const scoreContent = scoreTemplate(result);
  $(`#${alliance}Score`).html(scoreContent);

  // 1. Fuel 數量
  getInputElement(alliance, "AutoFuelCount").val(result.score.AutoFuelCount || 0);
  getInputElement(alliance, "TeleopFuelCount").val(result.score.TeleopFuelCount || 0);

  // 2. 機器人狀態 (索引 0, 1, 2)
  for (let i = 0; i < 3; i++) {
    // RobotsBypassed
    if (result.score.RobotsBypassed) {
      getInputElement(alliance, `RobotsBypassed${i}`).prop("checked", result.score.RobotsBypassed[i]);
    }
    // AutoTowerLevel1
    if (result.score.AutoTowerLevel1) {
      getInputElement(alliance, `AutoTowerLevel1${i}`).prop("checked", result.score.AutoTowerLevel1[i]);
    }
    // EndgameStatuses (Radio)
    if (result.score.EndgameStatuses) {
      getInputElement(alliance, `EndgameStatuses${i}`, result.score.EndgameStatuses[i]).prop("checked", true);
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

  // 4. 卡片
  if (result.cards != null) {
    $.each(result.cards, function (k, v) {
      getInputElement(alliance, `Team${k}Card`, v).prop("checked", true);
    });
  }
};

// 更新 JS 緩存結構
const updateResults = function (alliance) {
  const result = allianceResults[alliance];
  const formData = {};
  $.each($("form").serializeArray(), function (k, v) {
    formData[v.name] = v.value;
  });

  // 初始化陣列結構
  result.score.RobotsBypassed = [false, false, false];
  result.score.AutoTowerLevel1 = [false, false, false];
  result.score.EndgameStatuses = [0, 0, 0];

  // 讀取 Fuel
  result.score.AutoFuelCount = parseInt(formData[`${alliance}AutoFuelCount`]) || 0;
  result.score.TeleopFuelCount = parseInt(formData[`${alliance}TeleopFuelCount`]) || 0;

  for (let i = 0; i < 3; i++) {
    // 修正：這裡的索引必須與 HTML 的 {{range $i := seq 3}} 產生的 0, 1, 2 一致
    result.score.RobotsBypassed[i] = formData[`${alliance}RobotsBypassed${i}`] === "on";
    result.score.AutoTowerLevel1[i] = formData[`${alliance}AutoTowerLevel1${i}`] === "on";
    result.score.EndgameStatuses[i] = parseInt(formData[`${alliance}EndgameStatuses${i}`]) || 0;
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
    result.cards[team] = formData[`${alliance}Team${team}Card`];
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