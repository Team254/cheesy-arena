// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;
var plcOverrideAllowed = false;
var allowedOverrideMatchStates = [0, 5, 6, 7];
var disabledOverrideTooltipText = "Cannot override coil while match is in progress.";

// Sends a websocket message to play a given game sound on the audience display.
var playSound = function (sound) {
  websocket.send("playSound", sound);
};

// Sends a websocket message to set the selected LED test mode.
var setLedMode = function () {
  websocket.send("setLedMode", {
    RedMode: parseInt($("input[name=redLedMode]:checked").val(), 10),
    BlueMode: parseInt($("input[name=blueLedMode]:checked").val(), 10)
  });
};

var setPlcCoilOverride = function (index, override) {
  websocket.send("setPlcCoilOverride", { Index: index, Override: override });
};

var getNextOverrideState = function (overrideState) {
  switch (overrideState) {
    case "on":
      return "off";
    case "off":
      return "auto";
    default:
      return "on";
  }
};

var updateCoilOverrideTooltips = function () {
  $(".plc-coil-indicator").each(function (_index, element) {
    element.setAttribute("data-plc-override-allowed", plcOverrideAllowed);

    const tooltip = bootstrap.Tooltip.getInstance(element);
    if (plcOverrideAllowed) {
      if (tooltip) {
        tooltip.dispose();
      }
    } else if (!tooltip) {
      new bootstrap.Tooltip(element, { title: disabledOverrideTooltipText });
    }
  });
};

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function (data) {
  $.each(data.Inputs, function (index, input) {
    $("#input" + index).text(input)
    $("#input" + index).attr("data-plc-value", input);
  });

  $.each(data.Registers, function (index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    const coilIndicator = $("#coil" + index);
    coilIndicator.text(coil)
    coilIndicator.attr("data-plc-value", coil);
    coilIndicator.attr("data-plc-override", data.CoilOverrides[index]);
  });
};

var handleArenaStatus = function (data) {
  plcOverrideAllowed = allowedOverrideMatchStates.includes(data.MatchState);
  updateCoilOverrideTooltips();
};

// Handles a websocket message to update the LED mode selection.
var handLEDModeChange = function (data) {
  $("input[name=redLedMode][value=" + data.RedMode + "]").prop("checked", true);
  $("input[name=blueLedMode][value=" + data.BlueMode + "]").prop("checked", true);
}

$(function () {
  updateCoilOverrideTooltips();

  $(document).on("click", ".plc-coil-indicator", function (event) {
    if (!plcOverrideAllowed) {
      return;
    }

    const currentOverrideState = $(event.currentTarget).attr("data-plc-override");
    const nextOverrideState = getNextOverrideState(currentOverrideState);
    setPlcCoilOverride(parseInt($(event.currentTarget).attr("data-coil-index"), 10), nextOverrideState);
  });

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/field_testing/websocket", {
    plcIoChange: function (event) {
      handlePlcIoChange(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    },
    setLedMode: function (event) {
      handLEDModeChange(event.data);
    }
  });
});
