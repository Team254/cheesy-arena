package plc

var simulationState int

// Simulates the states of PLC input for a realistic match, for testing without a PLC and field.
func (plc *Plc) SimulateInput(matchTimeSec float64) {
	if matchTimeSec == 0 {
		simulationState = 0
	}

	switch simulationState {
	case 0:
		for i := 0; i < len(plc.inputs); i++ {
			plc.inputs[i] = false
		}
		for i := 0; i < len(plc.registers); i++ {
			plc.registers[i] = 0
		}
		if matchTimeSec > 0 {
			simulationState++
		}
	case 1:
		if matchTimeSec > 5 {
			plc.inputs[redSwitchNear] = true
			simulationState++
		}
	case 2:
		if matchTimeSec > 10 {
			plc.inputs[scaleNear] = true
			simulationState++
		}
	case 3:
		if matchTimeSec > 14 {
			plc.inputs[scaleNear] = false
			simulationState++
		}
	case 4:
		if matchTimeSec > 22 {
			plc.inputs[scaleFar] = true
			plc.registers[blueLevitateDistance] = 43
			simulationState++
		}
	case 5:
		if matchTimeSec > 25 {
			plc.registers[redForceDistance] = 98
			plc.registers[blueBoostDistance] = 98
			plc.inputs[blueLevitateActivate] = true
			simulationState++
		}
	case 6:
		if matchTimeSec > 28 {
			plc.registers[redForceDistance] = 58
			plc.registers[blueBoostDistance] = 58
			plc.inputs[blueLevitateActivate] = false
			simulationState++
		}
	case 7:
		if matchTimeSec > 31 {
			plc.registers[redForceDistance] = 43
			plc.registers[blueBoostDistance] = 43
			simulationState++
		}
	case 8:
		if matchTimeSec > 35 {
			plc.inputs[blueSwitchNear] = true
			plc.inputs[redForceActivate] = true
			simulationState++
		}
	case 9:
		if matchTimeSec > 36 {
			plc.inputs[redForceActivate] = false
			simulationState++
		}
	case 10:
		if matchTimeSec > 37 {
			plc.inputs[blueBoostActivate] = true
			simulationState++
		}
	case 11:
		if matchTimeSec > 38 {
			plc.inputs[blueBoostActivate] = false
			simulationState++
		}
	case 12:
		if matchTimeSec > 50 {
			plc.registers[redLevitateDistance] = 43
			plc.registers[redBoostDistance] = 58
			simulationState++
		}
	case 13:
		if matchTimeSec > 52 {
			plc.inputs[redBoostActivate] = true
			simulationState++
		}
	case 14:
		if matchTimeSec > 55 {
			plc.inputs[redBoostActivate] = false
			plc.inputs[redLevitateActivate] = true
			simulationState++
		}
	case 15:
		if matchTimeSec > 56 {
			plc.inputs[redLevitateActivate] = false
			simulationState++
		}
	case 16:
		if matchTimeSec > 155 {
			simulationState++
		}
	}

	plc.cycleCounter++
	if plc.cycleCounter == cycleCounterMax {
		plc.cycleCounter = 0
	}

	// Detect any changes in input or output and notify listeners if so.
	if plc.inputs != plc.oldInputs || plc.registers != plc.oldRegisters || plc.coils != plc.oldCoils {
		plc.IoChangeNotifier.Notify()
		plc.oldInputs = plc.inputs
		plc.oldRegisters = plc.registers
		plc.oldCoils = plc.coils
	}
}
