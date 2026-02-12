// scoring_panel.js
var websocket;

// 1. 初始化連線邏輯 (自動執行)
$(function () {
    if (typeof position === 'undefined') {
        position = window.location.href.split("/").slice(-1)[0];
    }
    
    // 使用 CheesyWebsocket 物件掛載所有接收事件
    websocket = new CheesyWebsocket("/panels/scoring/" + position + "/websocket", {
      onopen: function() {
        websocket.send("subscribe", {}); 
        // 嘗試觸發一次獲取分數的動作（這取決於後端支援什麼指令，有些是 "refresh" 或 "get_score"）
        websocket.send("fuel", { Adjustment: 0, Autonomous: true }); 
        console.log("Subscribed and sent initial refresh.");
        },  
      matchTime: function (event) {
            handleMatchTime(event.data);
      },
        realtimeScore: function (event) {
            handleRealtimeScore(event.data);
        },
        resetLocalState: function (event) {
            resetLocalState();
        }
    });
});

// 2. 處理時間顯示與 UI 狀態切換
function handleMatchTime(data) {
    // 使用傳入的數據計算分秒
    var totalSec = data.MatchTimeSec;
    var min = Math.floor(totalSec / 60);
    var sec = totalSec % 60;
    $("#match_time").text(min + ":" + (sec < 10 ? "0" : "") + sec);

    var stateText = "Pre-Match";
    // MatchState: Auto=3, Teleop=5, PostMatch=6
    switch(data.MatchState) {
        case 3: 
            stateText = "Autonomous"; 
            $("#auto-panel").css("opacity", "1");
            $("#teleop-panel").css("opacity", "0.5");
            break;
        case 5: 
            stateText = "Teleop"; 
            $("#auto-panel").css("opacity", "0.5");
            $("#teleop-panel").css("opacity", "1");
            break;
        case 6: 
            stateText = "Post-Match"; 
            $("#commit_btn").prop("disabled", false).text("COMMIT SCORE");
            break;
        default:
            stateText = "Pre-Match";
            break;
    }
    $("#match_state").text(stateText);
    // --- 新增：嘗試從時間包中直接同步 Hub 狀態 ---
    // 如果後端在傳送時間時也順便帶了 Hub 狀態，這裡就能抓到
    if (data.HubActiveRed !== undefined) {
        updateHubUI(data.HubActiveRed, data.HubActiveBlue);
    }
}

// 3. 處理分數同步更新
function handleRealtimeScore(data) {
    var myScore = (alliance === "red") ? data.Red.Score : data.Blue.Score;
    if (!myScore) return;

    // 更新 Hub 狀態顏色
    updateHubUI(data.Red.Score.HubActive, data.Blue.Score.HubActive);

    // 同步數值到畫面上
    $("#auto_fuel_count").text(myScore.AutoFuelCount);
    $("#teleop_fuel_count").text(myScore.TeleopFuelCount);

    for (var i = 0; i < 3; i++) {
        $("#auto_tower_" + i).prop("checked", myScore.AutoTowerLevel1[i]);
        var status = myScore.EndgameStatuses[i];
        $(`input[name=climb_${i}][value=${status}]`).prop("checked", true);
    }
}

// 4. Hub 狀態變色邏輯
function updateHubUI(redActive, blueActive) {
    var indicator = $("#hub-status-indicator");
    var card = $("#hub-status-card");

    // 如果後端傳來 undefined，預設為 false
    redActive = !!redActive; 
    blueActive = !!blueActive;

    if (redActive && blueActive) {
        indicator.text("BOTH ACTIVE").css({"background-color": "#198754", "color": "white"});
        card.css("border-color", "#198754");
    } else if (redActive) {
        indicator.text("RED ACTIVE").css({"background-color": "#dc3545", "color": "white"});
        card.css("border-color", "#dc3545");
    } else if (blueActive) {
        indicator.text("BLUE ACTIVE").css({"background-color": "#0d6efd", "color": "white"});
        card.css("border-color", "#0d6efd");
    } else {
        // 這是預設狀態：兩邊都沒啟動或資料尚未到達
        indicator.text("HUB INACTIVE").css({"background-color": "#6c757d", "color": "white"});
        card.css("border-color", "#6c757d");
    }
}

// 5. 按鈕指令發送函式 (與 HTML onclick 名稱對應)
function updateFuel(isAuto, delta) {
    websocket.send("fuel", { Adjustment: delta, Autonomous: isAuto });
}

function updateAutoTower(robotIdx, checked) {
    websocket.send("auto_tower", { RobotIndex: robotIdx, Adjustment: checked ? 1 : 0 });
}

function updateClimb(robotIdx, level) {
    websocket.send("climb", { RobotIndex: parseInt(robotIdx), Level: parseInt(level) });
}

function commitScore() {
    websocket.send("commitMatch", {});
    $("#commit_btn").text("SCORE COMMITTED").addClass("btn-secondary").removeClass("btn-primary").prop("disabled", true);
}

function resetLocalState() {
    $("#commit_btn").text("WAIT FOR MATCH END").prop("disabled", true).removeClass("btn-secondary").addClass("btn-primary");
}