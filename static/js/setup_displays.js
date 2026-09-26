// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the display configuration page.

var displayTemplate = Handlebars.compile($("#displayTemplate").html());
var websocket;
var displayRows = new Map();
var displayFields = ["Nickname", "Type", "Configuration"];
var nextSaveRequestId = 0;
var displayTableHovered = false;
var displaySocketConnected = false;

var displayValues = function (configuration) {
  return {
    Nickname: configuration.Nickname,
    Type: String(configuration.Type),
    Configuration: Object.keys(configuration.Configuration).sort().map(function (key) {
      return key + "=" + configuration.Configuration[key];
    }).join("&")
  };
};

var displayHasChanges = function (state) {
  return state.unconfirmed || displayFields.some(function (field) {
    return state.draft[field] !== state.baseline[field];
  });
};

var setDisplayText = function (element, value) {
  if (element.text() !== String(value)) {
    element.text(value);
  }
};

var updateDisplayEditStatus = function (state) {
  var changed = displayHasChanges(state);
  displayFields.forEach(function (field) {
    state.inputs[field].attr("data-changed", state.draft[field] !== state.baseline[field]);
  });
  state.row.find(".save-display").prop("disabled", !changed || !!state.pending || !displaySocketConnected);
  state.row.find(".discard-display").prop("disabled", !changed || !!state.pending);
  var status = state.error || (state.pending ? "Saving…" : changed ? "Unsaved changes" : state.saved ? "Saved" : "");
  var statusElement = state.row.find(".display-save-status");
  setDisplayText(statusElement, status);
  statusElement.toggleClass("text-danger", !!state.error);
};

// Merge server values only into untouched, inactive fields. Keep the same DOM nodes throughout an edit.
var syncDisplayInputs = function (state) {
  if (!state.pending && !state.unconfirmed && state.latest) {
    var values = displayValues(state.latest.DisplayConfiguration);
    displayFields.forEach(function (field) {
      var input = state.inputs[field][0];
      if (state.draft[field] === state.baseline[field] && document.activeElement !== input && !state.composing.has(field)) {
        state.baseline[field] = values[field];
        state.draft[field] = values[field];
        if (input.value !== values[field]) {
          input.value = values[field];
        }
      }
    });
  }
  updateDisplayEditStatus(state);
};

var finishDisplaySave = function (state, error, unconfirmed) {
  clearTimeout(state.pending.timer);
  state.pending = null;
  state.error = error;
  // After a timeout or disconnect, even a draft equal to the old baseline may need to undo an applied save.
  state.unconfirmed = state.unconfirmed || !!unconfirmed;
  updateDisplayEditStatus(state);
};

var configureDisplay = function (displayId) {
  var state = displayRows.get(displayId);
  if (!state || state.pending || !displayHasChanges(state)) {
    return;
  }
  var values = Object.assign({}, state.draft);
  var configurationMap = {};
  $.each(values.Configuration.split("&"), function (index, param) {
    var keyValuePair = param.split("=");
    if (keyValuePair[1] !== undefined) {
      configurationMap[keyValuePair[0]] = keyValuePair[1];
    }
  });
  var configuration = {
    Id: displayId,
    Nickname: values.Nickname,
    Type: parseInt(values.Type),
    Configuration: configurationMap
  };
  var requestId = String(++nextSaveRequestId);
  state.pending = {requestId: requestId, values: values, configuration: configuration};
  state.error = "";
  state.saved = false;
  state.pending.timer = setTimeout(function () {
    finishDisplaySave(state, "Save not confirmed. Changes retained; retry saving.", true);
  }, 10000);
  updateDisplayEditStatus(state);
  try {
    if (!displaySocketConnected) {
      throw new Error("Disconnected");
    }
    websocket.send("configureDisplay", Object.assign({RequestId: requestId}, configuration));
  } catch (error) {
    finishDisplaySave(state, "Could not save. Changes retained; retry when connected.", true);
  }
};

var handleDisplayConfigurationSaved = function (result) {
  var state = displayRows.get(result.Id);
  if (!state || !state.pending || state.pending.requestId !== result.RequestId) {
    return;
  }
  if (result.Error) {
    finishDisplaySave(state, result.Error);
    return;
  }
  // Acknowledge only the submitted values; edits made while saving remain dirty.
  state.baseline = state.pending.values;
  if (!state.latest || state.latest.Revision <= result.Revision) {
    state.latest = Object.assign({}, state.latest, {
      DisplayConfiguration: state.pending.configuration,
      Revision: result.Revision
    });
  }
  state.saved = true;
  state.unconfirmed = false;
  finishDisplaySave(state, "");
  syncDisplayInputs(state);
  reconcileDisplayRows();
};

var discardDisplayChanges = function (displayId) {
  var state = displayRows.get(displayId);
  if (!state || state.pending) {
    return;
  }
  var values = state.latest ? displayValues(state.latest.DisplayConfiguration) : state.baseline;
  state.baseline = Object.assign({}, values);
  state.draft = Object.assign({}, values);
  displayFields.forEach(function (field) {
    state.inputs[field].val(values[field]);
  });
  state.error = "";
  state.saved = false;
  state.unconfirmed = false;
  updateDisplayEditStatus(state);
  reconcileDisplayRows();
};

var reloadDisplay = function (displayId) {
  websocket.send("reloadDisplay", displayId);
};

var reloadAllDisplays = function () {
  websocket.send("reloadAllDisplays");
};

var createDisplayRow = function (display) {
  var id = display.DisplayConfiguration.Id;
  var values = displayValues(display.DisplayConfiguration);
  var state = {
    row: $(displayTemplate(display)), inputs: {}, baseline: Object.assign({}, values), draft: values,
    latest: display, present: true, pending: null, error: "", saved: false, unconfirmed: false, composing: new Set()
  };
  displayFields.forEach(function (field) {
    var input = state.row.find('[data-field="' + field + '"]');
    state.inputs[field] = input;
    input.val(values[field]);
    var recordEdit = function () {
      state.draft[field] = input.val();
      state.error = "";
      state.saved = false;
      updateDisplayEditStatus(state);
    };
    input.on("input change", recordEdit);
    input.on("compositionstart", function () { state.composing.add(field); });
    input.on("compositionend", function () {
      state.composing.delete(field);
      recordEdit();
    });
    input.on("blur", function () { setTimeout(function () { syncDisplayInputs(state); }, 0); });
  });
  state.row.find(".save-display").on("click", function () { configureDisplay(id); });
  state.row.find(".discard-display").on("click", function () { discardDisplayChanges(id); });
  state.row.find(".reload-display").on("click", function () { reloadDisplay(id); });
  displayRows.set(id, state);
  return state;
};

// Defer insertions, removals, and sorting while the table is being used, including pointer interaction.
var reconcileDisplayRows = function () {
  var container = document.getElementById("displayContainer");
  if (displayTableHovered || container.contains(document.activeElement)) {
    return;
  }
  displayRows.forEach(function (state, id) {
    if (!state.present && !state.pending && !displayHasChanges(state) && !state.error) {
      state.row.remove();
      displayRows.delete(id);
    }
  });
  var rows = Array.from(displayRows.entries()).sort(function (a, b) {
    var nicknameA = a[1].latest.DisplayConfiguration.Nickname;
    var nicknameB = b[1].latest.DisplayConfiguration.Nickname;
    if (nicknameA === "" && nicknameB !== "") {
      return 1;
    }
    if (nicknameA !== "" && nicknameB === "") {
      return -1;
    }
    return nicknameA.localeCompare(nicknameB) || a[0].localeCompare(b[0]);
  });
  rows.forEach(function (entry, index) {
    var row = entry[1].row[0];
    if (container.children[index] !== row) {
      container.insertBefore(row, container.children[index] || null);
    }
  });
};

var handleDisplayConfiguration = function (data) {
  displayRows.forEach(function (state) { state.present = false; });
  Object.keys(data).forEach(function (id) {
    var display = data[id];
    var state = displayRows.get(id) || createDisplayRow(display);
    state.present = true;
    // A status notification may have been queued before the save acknowledgement.
    if (!state.latest || display.Revision >= state.latest.Revision) {
      state.latest = display;
    }
    setDisplayText(state.row.find(".display-connection-count"), display.ConnectionCount);
    setDisplayText(state.row.find(".display-ip-address"), display.IpAddress);
    state.row.toggleClass("danger", !display.ConnectionCount);
    syncDisplayInputs(state);
  });
  displayRows.forEach(function (state) {
    if (!state.present) {
      setDisplayText(state.row.find(".display-connection-count"), "0");
    }
  });
  reconcileDisplayRows();
};

$(function () {
  $("#displayContainer").on("mouseenter", function () { displayTableHovered = true; });
  $("#displayContainer").on("mouseleave", function () {
    displayTableHovered = false;
    reconcileDisplayRows();
  });
  $("#displayContainer").on("focusout", function () { setTimeout(reconcileDisplayRows, 0); });
  window.addEventListener("beforeunload", function (event) {
    if (Array.from(displayRows.values()).some(function (state) { return state.pending || displayHasChanges(state); })) {
      event.preventDefault();
      event.returnValue = "";
    }
  });
  websocket = new CheesyWebsocket("/setup/displays/websocket", {
    displayConfiguration: function (event) { handleDisplayConfiguration(event.data); },
    displayConfigurationSaved: function (event) { handleDisplayConfigurationSaved(event.data); }
  }, {
    open: function () {
      displaySocketConnected = true;
      displayRows.forEach(function (state) {
        // Revisions belong to the server process, which may have restarted during the disconnect.
        state.latest.Revision = 0;
        updateDisplayEditStatus(state);
      });
    },
    close: function () {
      displaySocketConnected = false;
      displayRows.forEach(function (state) {
        if (state.pending) {
          finishDisplaySave(state, "Connection lost before save was confirmed. Changes retained; retry saving.", true);
        }
        updateDisplayEditStatus(state);
      });
    }
  });
});
