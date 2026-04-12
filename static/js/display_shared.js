// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared helpers for audience and wall displays.

(function (window) {
  window.DisplayShared = {
    applyDisplaySides: function (urlParams) {
      const reversed = urlParams.get("reversed");
      const redSide = reversed === "true" ? "right" : "left";
      const blueSide = reversed === "true" ? "left" : "right";
      $(".reversible-left").attr("data-reversed", reversed);
      $(".reversible-right").attr("data-reversed", reversed);
      return {redSide: redSide, blueSide: blueSide};
    },

    getAvatarUrl: function (teamId) {
      return "/api/teams/" + teamId + "/avatar";
    },

    handleMatchLoad: function (data, redSide, blueSide) {
      const currentMatch = data.Match;
      $(`#${redSide}Team1`).text(currentMatch.Red1);
      $(`#${redSide}Team1`).attr("data-yellow-card", data.Teams["R1"]?.YellowCard);
      $(`#${redSide}Team2`).text(currentMatch.Red2);
      $(`#${redSide}Team2`).attr("data-yellow-card", data.Teams["R2"]?.YellowCard);
      $(`#${redSide}Team3`).text(currentMatch.Red3);
      $(`#${redSide}Team3`).attr("data-yellow-card", data.Teams["R3"]?.YellowCard);
      $(`#${redSide}Team1Avatar`).attr("src", this.getAvatarUrl(currentMatch.Red1));
      $(`#${redSide}Team2Avatar`).attr("src", this.getAvatarUrl(currentMatch.Red2));
      $(`#${redSide}Team3Avatar`).attr("src", this.getAvatarUrl(currentMatch.Red3));
      $(`#${blueSide}Team1`).text(currentMatch.Blue1);
      $(`#${blueSide}Team1`).attr("data-yellow-card", data.Teams["B1"]?.YellowCard);
      $(`#${blueSide}Team2`).text(currentMatch.Blue2);
      $(`#${blueSide}Team2`).attr("data-yellow-card", data.Teams["B2"]?.YellowCard);
      $(`#${blueSide}Team3`).text(currentMatch.Blue3);
      $(`#${blueSide}Team3`).attr("data-yellow-card", data.Teams["B3"]?.YellowCard);
      $(`#${blueSide}Team1Avatar`).attr("src", this.getAvatarUrl(currentMatch.Blue1));
      $(`#${blueSide}Team2Avatar`).attr("src", this.getAvatarUrl(currentMatch.Blue2));
      $(`#${blueSide}Team3Avatar`).attr("src", this.getAvatarUrl(currentMatch.Blue3));

      if (currentMatch.Type === matchTypePlayoff) {
        $(`#${redSide}PlayoffAlliance`).text(currentMatch.PlayoffRedAlliance);
        $(`#${blueSide}PlayoffAlliance`).text(currentMatch.PlayoffBlueAlliance);
        $(".playoff-alliance").show();

        if (data.Matchup.NumWinsToAdvance > 1) {
          $(`#${redSide}PlayoffAllianceWins`).text(data.Matchup.RedAllianceWins);
          $(`#${blueSide}PlayoffAllianceWins`).text(data.Matchup.BlueAllianceWins);
          $("#playoffSeriesStatus").css("display", "flex");
        } else {
          $("#playoffSeriesStatus").hide();
        }
      } else {
        $(`#${redSide}PlayoffAlliance`).text("");
        $(`#${blueSide}PlayoffAlliance`).text("");
        $(".playoff-alliance").hide();
        $("#playoffSeriesStatus").hide();
      }

      let matchName = data.Match.LongName;
      if (data.Match.NameDetail !== "") {
        matchName += " &ndash; " + data.Match.NameDetail;
      }
      $("#matchName").html(matchName);
      $("#timeoutNextMatchName").html(matchName);
      $("#timeoutBreakDescription").text(data.BreakDescription);
      return currentMatch;
    },

    handleMatchTime: function (data) {
      translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
        $("#matchTime").text(getCountdownString(countdownSec));
      });
    },

    handle2026RealtimeScore: function (data, currentMatch, redSide, blueSide, updateHubActiveIndicator) {
      $(`#${redSide}ScoreNumber`).text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.TeleopTowerPoints);
      $(`#${blueSide}ScoreNumber`).text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.TeleopTowerPoints);

      $(`#${redSide}FuelNumerator`).text(data.Red.ScoreSummary.NumFuel);
      $(`#${redSide}FuelDenominator`).text(data.Red.ScoreSummary.NumFuelGoal);
      $(`#${blueSide}FuelNumerator`).text(data.Blue.ScoreSummary.NumFuel);
      $(`#${blueSide}FuelDenominator`).text(data.Blue.ScoreSummary.NumFuelGoal);
      if (currentMatch && currentMatch.Type === matchTypePlayoff) {
        $(`#${redSide}FuelDenominator`).hide();
        $(`#${blueSide}FuelDenominator`).hide();
      } else {
        $(`#${redSide}FuelDenominator`).show();
        $(`#${blueSide}FuelDenominator`).show();
      }

      updateHubActiveIndicator(redSide, data.Red.ActiveRemainingSec, data.Red.ActiveDurationSec);
      updateHubActiveIndicator(blueSide, data.Blue.ActiveRemainingSec, data.Blue.ActiveDurationSec);
    },

    createHubActiveController: function (getCurrentScreen) {
      const activeProgressLength = 158;
      const leftActiveProgressStartOffset = parseFloat($("#leftHubActive svg .active-progress").attr("stroke-dashoffset"));
      const rightActiveProgressStartOffset = parseFloat(
        $("#rightHubActive svg .active-progress").attr("stroke-dashoffset")
      );
      const activeFadeTimeMs = 300;
      const activeDwellTimeMs = 500;
      const hubActiveStateBySide = {
        left: {
          active: false, lastRemainingSec: 0, lastDurationSec: 0,
          activeUntilTimeMs: null,
          hideTimeoutId: null, resetTimeoutId: null, animationFrameId: null, pendingRestart: false,
        },
        right: {
          active: false, lastRemainingSec: 0, lastDurationSec: 0,
          activeUntilTimeMs: null,
          hideTimeoutId: null, resetTimeoutId: null, animationFrameId: null, pendingRestart: false,
        },
      };

      const getActiveProgressStartOffset = function (side) {
        return side === "left" ? leftActiveProgressStartOffset : rightActiveProgressStartOffset;
      };

      const getActiveProgressEndOffset = function (side) {
        return side === "left" ? activeProgressLength : -activeProgressLength;
      };

      const getActiveProgressOffset = function (side, activeRemainingSec, activeDurationSec) {
        const progressRatio = Math.max(0, Math.min(activeRemainingSec, activeDurationSec)) / activeDurationSec;
        const startOffset = getActiveProgressStartOffset(side);
        const endOffset = getActiveProgressEndOffset(side);
        return startOffset + (1 - progressRatio) * (endOffset - startOffset);
      };

      const getAdjustedActiveRemainingSec = function (state) {
        if (state.activeUntilTimeMs === null) {
          return state.lastRemainingSec;
        }

        return Math.max(0, (state.activeUntilTimeMs - Date.now()) / 1000);
      };

      const restartHubActiveAnimation = function (state, hubActiveCircle, side, activeRemainingSec, activeDurationSec) {
        if (state.animationFrameId !== null) {
          cancelAnimationFrame(state.animationFrameId);
          state.animationFrameId = null;
        }

        hubActiveCircle.stop(true, true);
        hubActiveCircle.css("transition", "none");
        hubActiveCircle.css("stroke-dashoffset", getActiveProgressOffset(side, activeRemainingSec, activeDurationSec));
        hubActiveCircle[0].getBoundingClientRect();

        state.animationFrameId = requestAnimationFrame(function () {
          hubActiveCircle.css("transition", `stroke-dashoffset ${activeRemainingSec * 1000}ms linear`);
          hubActiveCircle.css("stroke-dashoffset", getActiveProgressEndOffset(side));
          state.animationFrameId = null;
        });
      };

      return {
        restartPendingHubActiveIndicators: function () {
          $.each(hubActiveStateBySide, function (side, state) {
            const activeRemainingSec = getAdjustedActiveRemainingSec(state);
            if (!state.pendingRestart || !state.active || activeRemainingSec <= 0 || state.lastDurationSec <= 0) {
              return;
            }

            const hubActiveCircle = $(`#${side}HubActive svg .active-progress`);
            restartHubActiveAnimation(state, hubActiveCircle, side, activeRemainingSec, state.lastDurationSec);
            state.pendingRestart = false;
          });
        },

        updateHubActiveIndicator: function (side, activeRemainingSec, activeDurationSec) {
          const state = hubActiveStateBySide[side];
          const hubActiveDiv = $(`#${side}HubActive`);
          const hubActiveCircle = $(`#${side}HubActive svg .active-progress`);
          const hubActiveText = $(`#${side}HubActive svg text`);
          const wasActive = state.active;

          if (state.hideTimeoutId !== null) {
            clearTimeout(state.hideTimeoutId);
            state.hideTimeoutId = null;
          }
          if (state.resetTimeoutId !== null) {
            clearTimeout(state.resetTimeoutId);
            state.resetTimeoutId = null;
          }

          if (activeRemainingSec > 0 && activeDurationSec > 0) {
            const shouldRestartAnimation = !state.active ||
              activeRemainingSec > state.lastRemainingSec ||
              activeDurationSec !== state.lastDurationSec;

            state.active = true;
            state.activeUntilTimeMs = Date.now() + activeRemainingSec * 1000;
            state.pendingRestart = getCurrentScreen() !== "match";
            hubActiveDiv.attr("data-active", true);
            hubActiveText.text(activeRemainingSec);

            if (shouldRestartAnimation) {
              if (state.pendingRestart) {
                hubActiveCircle.stop(true, true);
                hubActiveCircle.css("transition", "");
                hubActiveCircle.css("stroke-dashoffset", getActiveProgressOffset(side, activeRemainingSec, activeDurationSec));
              } else {
                restartHubActiveAnimation(state, hubActiveCircle, side, activeRemainingSec, activeDurationSec);
              }
            }
          } else {
            state.active = false;
            state.activeUntilTimeMs = null;
            state.pendingRestart = false;
            hubActiveText.text(activeRemainingSec);
            if (state.animationFrameId !== null) {
              cancelAnimationFrame(state.animationFrameId);
              state.animationFrameId = null;
            }
            if (wasActive) {
              state.hideTimeoutId = setTimeout(function () {
                hubActiveDiv.attr("data-active", false);
                state.resetTimeoutId = setTimeout(function () {
                  if (!state.active) {
                    hubActiveCircle.stop(true, true);
                    hubActiveCircle.css("transition", "");
                    hubActiveCircle.css("stroke-dashoffset", getActiveProgressStartOffset(side));
                  }
                  state.resetTimeoutId = null;
                }, activeFadeTimeMs);
                state.hideTimeoutId = null;
              }, activeDwellTimeMs);
            } else {
              hubActiveCircle.stop(true, true);
              hubActiveCircle.css("transition", "");
              hubActiveCircle.css("stroke-dashoffset", getActiveProgressStartOffset(side));
              hubActiveDiv.attr("data-active", false);
            }
          }

          state.lastRemainingSec = activeRemainingSec;
          state.lastDurationSec = activeDurationSec;
        },
      };
    },
  };
})(window);
