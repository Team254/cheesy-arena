// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents a scheduled break in the playoff match schedule.

package playoff

type breakSpec struct {
	orderBefore int
	durationSec int
	description string
}
