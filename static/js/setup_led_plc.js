// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field setup page.

var websocket;

// Sends a websocket message to change the LED display mode.
var setLedMode = function() {
  websocket.send("setLedMode", {LedMode: parseInt($("input[name=ledMode]:checked").val()),
      VaultLedMode: parseInt($("input[name=vaultLedMode]:checked").val())});
};

// Handles a websocket message to update the LED test mode.
var handleLedMode = function(data) {
  $("input[name=ledMode]:checked").prop("checked", false);
  $("input[name=ledMode][value=" + data.LedMode + "]").prop("checked", true);

  $("input[name=vaultLedMode]:checked").prop("checked", false);
  $("input[name=vaultLedMode][value=" + data.VaultLedMode + "]").prop("checked", true);
};

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function(data) {
  $.each(data.Inputs, function(index, input) {
    $("#input" + index).text(input)
  });

  $.each(data.Registers, function(index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function(index, coil) {
    $("#coil" + index).text(coil)
  });
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/led_plc/websocket", {
    ledMode: function(event) {handleLedMode(event.data); },
    plcIoChange: function(event) { handlePlcIoChange(event.data); }
  });
});
