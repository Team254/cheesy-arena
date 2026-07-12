// Copyright 2026 Team 254. All Rights Reserved.
//
// Client-side logic for the hub display. The hub is physically shared by both alliances, so this shows a single
// color representing whichever alliance currently has scoring access:
//   - Exclusive shift (only one alliance's hub active): that alliance's color (red/blue).
//   - Both active at once (e.g. Auto/Transition/Endgame, when the shared hub is open to everyone): neutral/off,
//     since no single alliance color would be accurate.
//   - Manual override (field clear/reset signaled from Match Play): purple/green, takes priority over the above.
var websocket;

// Combines the independently-tracked red/blue color state into the single color this shared display should show.
var combineHubColor = function (redColor, blueColor) {
    if (redColor === "purple" || blueColor === "purple") {
        return "purple";
    }
    if (redColor === "green" || blueColor === "green") {
        return "green";
    }
    if (redColor === "red" && blueColor !== "blue") {
        return "red";
    }
    if (blueColor === "blue" && redColor !== "red") {
        return "blue";
    }
    // Either both are active at once, or neither is -- show neutral in both cases.
    return "off";
};

// Handles a websocket message containing the current color of every hub and applies the combined result.
var handleHubColor = function (data) {
    var color = combineHubColor(data["red"] || "off", data["blue"] || "off");
    $("#hub").attr("data-color", color);
    $("#hubLabel").text(color === "off" ? "" : color.toUpperCase());
};

$(function () {
    // Set up the websocket back to the server.
    websocket = new CheesyWebsocket("/displays/hub/websocket", {
        hubColor: function (event) {
            handleHubColor(event.data);
        }
    });
});