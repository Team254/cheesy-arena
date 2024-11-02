package plc

import (
	"time"

	"github.com/Team254/cheesy-arena/websocket"
)

type ManualScorePLC struct {
	ioChangeNotifier *websocket.Notifier
}

var RedAmpAmplify bool = false
var BlueAmpAmplify bool = false
var RedAmpCoop bool = false
var BlueAmpCoop bool = false

var RedAmpScore int = 0
var BlueAmpScore int = 0
var RedSpeakerScore int = 0
var BlueSpeakerScore int = 0

func (plc *ManualScorePLC) SetAddress(address string) {

}

// Returns true if the PLC is enabled in the configurations.
func (plc *ManualScorePLC) IsEnabled() bool {
	return true
}

// Returns true if the PLC is connected and responding to requests.
func (plc *ManualScorePLC) IsHealthy() bool {
	return true
}

// Returns a notifier which fires whenever the I/O values change.
func (plc *ManualScorePLC) IoChangeNotifier() *websocket.Notifier {
	return plc.ioChangeNotifier
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *ManualScorePLC) Run() {
	for {
		plc.ioChangeNotifier.Notify()
		time.Sleep(time.Millisecond * 200)
	}
}

// Returns a map of ArmorBlocks I/O module names to whether they are connected properly.
func (plc *ManualScorePLC) GetArmorBlockStatuses() map[string]bool {
	statuses := make(map[string]bool, armorBlockCount)
	return statuses
}

// Returns the state of the field emergency stop button (true if e-stop is active).
func (plc *ManualScorePLC) GetFieldEStop() bool {
	return false
}

// Returns the state of the red and blue driver station emergency stop buttons (true if E-stop is active).
func (plc *ManualScorePLC) GetTeamEStops() ([3]bool, [3]bool) {
	var redEStops, blueEStops [3]bool
	redEStops[0] = false
	redEStops[1] = false
	redEStops[2] = false
	blueEStops[0] = false
	blueEStops[1] = false
	blueEStops[2] = false
	return redEStops, blueEStops
}

// Returns the state of the red and blue driver station autonomous stop buttons (true if A-stop is active).
func (plc *ManualScorePLC) GetTeamAStops() ([3]bool, [3]bool) {
	var redAStops, blueAStops [3]bool
	redAStops[0] = false
	redAStops[1] = false
	redAStops[2] = false
	blueAStops[0] = false
	blueAStops[1] = false
	blueAStops[2] = false
	return redAStops, blueAStops
}

// Returns whether anything is connected to each station's designated Ethernet port on the SCC.
func (plc *ManualScorePLC) GetEthernetConnected() ([3]bool, [3]bool) {
	return [3]bool{
			true,
			true,
			true,
		},
		[3]bool{
			true,
			true,
			true,
		}
}

// Resets the internal state of the PLC to start a new match.
func (plc *ManualScorePLC) ResetMatch() {
	plc.ioChangeNotifier.Notify()
	RedAmpAmplify = false
	RedAmpCoop = false
	BlueAmpAmplify = false
	BlueAmpCoop = false

	RedAmpScore = 0
	BlueAmpScore = 0
	RedSpeakerScore = 0
	BlueSpeakerScore = 0
}

// Sets the on/off state of the stack lights on the scoring table.
func (plc *ManualScorePLC) SetStackLights(red, blue, orange, green bool) {

}

// Triggers the "match ready" chime if the state is true.
func (plc *ManualScorePLC) SetStackBuzzer(state bool) {

}

// Sets the on/off state of the field reset light.
func (plc *ManualScorePLC) SetFieldResetLight(state bool) {

}

func (plc *ManualScorePLC) GetCycleState(max, index, duration int) bool {
	return false
}

func (plc *ManualScorePLC) GetInputNames() []string {
	inputNames := make([]string, inputCount)

	return inputNames
}

func (plc *ManualScorePLC) GetRegisterNames() []string {
	registerNames := make([]string, registerCount)

	return registerNames
}

func (plc *ManualScorePLC) GetCoilNames() []string {
	coilNames := make([]string, coilCount)

	return coilNames
}

// Returns the state of the red amplify, red co-op, blue amplify, and blue co-op buttons, respectively.
func (plc *ManualScorePLC) GetAmpButtons() (bool, bool, bool, bool) {
	return RedAmpAmplify, RedAmpCoop, BlueAmpAmplify, BlueAmpCoop
}

// Returns the red amp, red speaker, blue amp, and blue speaker note counts, respectively.
func (plc *ManualScorePLC) GetAmpSpeakerNoteCounts() (int, int, int, int) {
	return RedAmpScore,
		RedSpeakerScore,
		BlueAmpScore,
		BlueSpeakerScore
}

// Sets the on/off state of the serializer motors within each speaker.
func (plc *ManualScorePLC) SetSpeakerMotors(state bool) {
}

// Sets the state of the amplification lights on the red and blue speakers.
func (plc *ManualScorePLC) SetSpeakerLights(redState, blueState bool) {

}

// Sets the state of the red and blue subwoofer countdown lights. When the state is set to true, the lights light up and
// begin the ten-second coundown sequence. When set to false before the countdown is complete, the lights will turn off.
func (plc *ManualScorePLC) SetSubwooferCountdown(redState, blueState bool) {

}

// Sets the state of the red and blue amp lights.
func (plc *ManualScorePLC) SetAmpLights(redLow, redHigh, redCoop, blueLow, blueHigh, blueCoop bool) {

}

// Sets the state of the post-match subwoofer lights.
func (plc *ManualScorePLC) SetPostMatchSubwooferLights(state bool) {
}
