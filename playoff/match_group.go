// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Interface representing a stage of a playoff tournament containing one or more related matches.

package playoff

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"sort"
)

type MatchGroup interface {
	// Id returns the unique identifier for the match group.
	Id() string

	// MatchSpecs returns the list of match specifications that define the matches in the match group.
	MatchSpecs() []*matchSpec

	// update updates the state of each match group based on the results of the given played matches.
	update(playoffMatchResults map[int]playoffMatchResult)

	// traverse performs a depth-first traversal of the playoff graph and invokes the given function after visiting each
	// match group's children.
	traverse(visitFunction func(MatchGroup) error) error
}

type matchSpec struct {
	longName            string
	shortName           string
	nameDetail          string
	matchGroupId        string
	order               int
	durationSec         int
	useTiebreakCriteria bool
	isHidden            bool
	tbaMatchKey         model.TbaMatchKey
	redAllianceId       int
	blueAllianceId      int
}

// collectMatchGroups returns a map of all match groups including and below the given root match group, keyed by ID.
func collectMatchGroups(rootMatchGroup MatchGroup) (map[string]MatchGroup, error) {
	matchGroups := make(map[string]MatchGroup)
	err := rootMatchGroup.traverse(func(matchGroup MatchGroup) error {
		if _, ok := matchGroups[matchGroup.Id()]; ok {
			return fmt.Errorf("match group with ID %q defined more than once", matchGroup.Id())
		}
		matchGroups[matchGroup.Id()] = matchGroup
		return nil
	})
	return matchGroups, err
}

// collectMatches returns a slice of all matches including and below the given root match group, in order of play.
func collectMatchSpecs(rootMatchGroup MatchGroup) ([]*matchSpec, error) {
	uniqueLongNames := make(map[string]struct{})
	uniqueShortNames := make(map[string]struct{})
	uniqueOrders := make(map[int]struct{})
	uniqueTbaKeys := make(map[model.TbaMatchKey]struct{})

	var matches []*matchSpec
	err := rootMatchGroup.traverse(func(matchGroup MatchGroup) error {
		for _, match := range matchGroup.MatchSpecs() {
			if _, ok := uniqueLongNames[match.longName]; ok {
				return fmt.Errorf("match with long name %q defined more than once", match.longName)
			}
			if _, ok := uniqueShortNames[match.shortName]; ok {
				return fmt.Errorf("match with short name %q defined more than once", match.shortName)
			}
			if _, ok := uniqueOrders[match.order]; ok {
				return fmt.Errorf("match with order %d defined more than once", match.order)
			}
			if _, ok := uniqueTbaKeys[match.tbaMatchKey]; ok {
				return fmt.Errorf("match with TBA key %q defined more than once", match.tbaMatchKey)
			}

			match.matchGroupId = matchGroup.Id()
			matches = append(matches, match)
			uniqueLongNames[match.longName] = struct{}{}
			uniqueShortNames[match.shortName] = struct{}{}
			uniqueOrders[match.order] = struct{}{}
			uniqueTbaKeys[match.tbaMatchKey] = struct{}{}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].order < matches[j].order
	})
	return matches, nil
}
