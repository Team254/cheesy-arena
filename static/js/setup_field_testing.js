// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;

// Sends a websocket message to play a given game sound on the audience display.
var playSound = function (sound) {
  websocket.send("playSound", sound);
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
    $("#coil" + index).text(coil)
    $("#coil" + index).attr("data-plc-value", coil);
  });
};

// Handles a websocket message to update the hub LED status.
var handleHubLedChange = function (data) {
  $("#hubLedRed").text("rgb(" + data.Red.R + ", " + data.Red.G + ", " + data.Red.B + ")");
  $("#hubLedBlue").text("rgb(" + data.Blue.R + ", " + data.Blue.G + ", " + data.Blue.B + ")");
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/field_testing/websocket", {
    plcIoChange: function (event) {
      handlePlcIoChange(event.data);
    },
    hubLed: function (event) {
      handleHubLedChange(event.data);
    }
  });
});
