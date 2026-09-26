// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Run with: node --test static/js/stream_displays_test.js

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

function loadDisplay(name, query, playerApi) {
  const state = {sockets: [], text: ""};
  const context = vm.createContext({
    URLSearchParams,
    window: {location: {search: query, origin: "http://localhost:8080"}, innerWidth: 1920, innerHeight: 1080},
    $: function (value) {
      if (typeof value === "function") {
        value();
      } else {
        return {text: function (text) { state.text = text; }};
      }
    },
    CheesyWebsocket: function (endpoint) { state.sockets.push(endpoint); },
    ...playerApi
  });
  vm.runInContext(fs.readFileSync(path.join(__dirname, name + "_display.js"), "utf8"), context);
  return {context, state};
}

for (const captions of ["true", "false", ""]) {
  test("Twitch caption setting: " + (captions || "omitted"), function () {
    const calls = [];
    let ready;
    function Embed(id, options) {
      assert.equal(id, "twitchEmbed");
      assert.equal(options.channel, "team254");
      assert.equal(options.layout, "video");
      this.addEventListener = function (event, callback) {
        assert.equal(event, "ready");
        ready = callback;
      };
      this.getPlayer = function () {
        return {
          enableCaptions: function () { calls.push("enable"); },
          disableCaptions: function () { calls.push("disable"); }
        };
      };
    }
    Embed.VIDEO_READY = "ready";
    const {state} = loadDisplay("twitch", "?channel=team254" + (captions ? "&captions=" + captions : ""), {
      Twitch: {Embed}
    });
    assert.deepEqual(state.sockets, ["/displays/twitch/websocket"]);
    assert.deepEqual(calls, []);
    ready();
    assert.deepEqual(calls, captions === "true" ? ["enable"] : ["disable"]);
  });

  test("YouTube caption setting: " + (captions || "omitted"), function () {
    let options;
    const {context, state} = loadDisplay("youtube", "?videoId=liveVideoId" + (captions ? "&captions=" + captions : ""), {
      YT: {Player: function (id, playerOptions) {
        assert.equal(id, "youtubeEmbed");
        options = playerOptions;
      }}
    });
    // The server connection must not depend on the external API being ready.
    assert.deepEqual(state.sockets, ["/displays/youtube/websocket"]);
    assert.equal(options, undefined);
    context.onYouTubeIframeAPIReady();
    assert.equal(options.videoId, "liveVideoId");
    assert.equal(options.width, "100%");
    assert.equal(options.height, "100%");
    assert.equal(options.playerVars.origin, "http://localhost:8080");
    assert.equal(options.playerVars.cc_load_policy, captions === "true" ? 1 : 0);
    const calls = [];
    options.events.onReady({target: {
      mute: function () { calls.push("mute"); },
      playVideo: function () { calls.push("play"); }
    }});
    assert.deepEqual(calls, ["mute", "play"]);
  });
}

test("Unconfigured YouTube display stays available for remote configuration", function () {
  const {context, state} = loadDisplay("youtube", "?videoId=", {});
  context.onYouTubeIframeAPIReady();
  assert.match(state.text, /Set videoId/);
  assert.deepEqual(state.sockets, ["/displays/youtube/websocket"]);
});
