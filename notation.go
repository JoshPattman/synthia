package main

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"wave"
)

type Action interface {
	action()
}

func (Play) action()    {}
func (Advance) action() {}

type Play struct {
	Key        wave.Key
	Octave     int
	Duration   float64
	Instrument string
}

type Advance struct {
	Duration float64
}

func ParseRTTTL(rtttl string) ([]Action, error) {
	// Parse Defaults
	sections := strings.Split(rtttl, ":")
	dFinder := regexp.MustCompile(`d\s*=\s*(\d+)`)
	oFinder := regexp.MustCompile(`o\s*=\s*(\d+)`)
	bFinder := regexp.MustCompile(`b\s*=\s*(\d+)`)
	dGroups := dFinder.FindStringSubmatch(sections[1])
	oGroups := oFinder.FindStringSubmatch(sections[1])
	bGroups := bFinder.FindStringSubmatch(sections[1])
	if dGroups[1] == "" || bGroups[1] == "" || oGroups[1] == "" {
		return nil, errors.New("must specify all defaults")
	}
	defaultO, err := strconv.Atoi(oGroups[1])
	if err != nil {
		return nil, err
	}
	defaultB, err := strconv.Atoi(bGroups[1])
	if err != nil {
		return nil, err
	}
	defaultD, err := strconv.Atoi(dGroups[1])
	if err != nil {
		return nil, err
	}

	// Parse notes
	rawNotes := strings.Split(sections[2], ",")
	actions := make([]Action, 0)
	noteDetector := regexp.MustCompile(`(\d+)?([cdefgabp]#?)(\d+)?(\.)?`)
	for _, rn := range rawNotes {
		noteString := strings.TrimSpace(rn)
		groups := noteDetector.FindStringSubmatch(noteString)
		if groups[1] == "" {
			groups[1] = fmt.Sprint(defaultD)
		}
		if groups[3] == "" {
			groups[3] = fmt.Sprint(defaultO)
		}
		d, err := strconv.Atoi(groups[1])
		if err != nil {
			return nil, err
		}
		duration := 60.0 / float64(defaultB*d)
		if groups[4] == "." {
			duration *= 1.5
		}
		if groups[2] != "p" {
			key, err := wave.ParseStringToKey(groups[2])
			if err != nil {
				return nil, err
			}
			o, err := strconv.Atoi(groups[3])
			if err != nil {
				return nil, err
			}
			actions = append(actions, Play{Key: key, Octave: o, Duration: duration})
		}
		actions = append(actions, Advance{Duration: duration})
	}
	return actions, nil
}

func ParseAdvancedNotation(code string, bpm float64) ([]Action, error) {
	commentsAndWhitespace := regexp.MustCompile(`(?:\[.*?\])|\s`)
	code = commentsAndWhitespace.ReplaceAllString(code, "")
	blockFinder := regexp.MustCompile(`<(\w+)>(.*?)</>`)
	blocks := blockFinder.FindAllStringSubmatch(code, 999)
	blocksMap := make(map[string]string)
	for _, b := range blocks {
		blocksMap[strings.ToLower(b[1])] = b[2]
	}
	timeline, err := addAdvancedToTimeline(0, "", make([]timelineEntry, 0), "Main", blocksMap, bpm)
	if err != nil {
		return nil, err
	}
	actions := buildActionsFromTimeline(timeline)
	return actions, nil
}

// Timeline is UNORDERED
func addAdvancedToTimeline(now float64, instrument string, timeline []timelineEntry, blockName string, blocks map[string]string, bpm float64) ([]timelineEntry, error) {
	blockName = strings.ToLower(blockName)
	code, ok := blocks[blockName]
	if !ok {
		return nil, fmt.Errorf("tried to insert '%s' but it did not exist, blocks are: %v", blockName, slices.Collect(maps.Keys(blocks)))
	}
	parts := strings.Split(code, ";")
	for _, c := range parts {
		c = strings.TrimSpace(strings.ToLower(c))
		if c == "" {
			continue
		}

		if isInstrumentChange, changeTo, err := tryParseInstrumentChange(c); isInstrumentChange {
			if err != nil {
				return nil, err
			}
			instrument = changeTo
		} else if isInsertSubsection, subsectionName, err := tryParseInsertSubsection(c); isInsertSubsection {
			if err != nil {
				return nil, err
			}
			tl, err := addAdvancedToTimeline(now, instrument, timeline, subsectionName, blocks, bpm)
			if err != nil {
				return nil, err
			}
			timeline = tl
		} else if isAdvance, advancedBy, err := tryParseAdvanceTimeline(c); isAdvance {
			if err != nil {
				return nil, err
			}
			now += advancedBy / (bpm / 60.0)
		} else if isNote, key, octave, duration, err := tryParsePlayNote(c); isNote {
			if err != nil {
				return nil, err
			}
			timeline = append(timeline, timelineEntry{now, Play{Key: key, Octave: octave, Duration: duration / (bpm / 60.0), Instrument: instrument}})
		} else {
			return nil, fmt.Errorf("cannot parse instruction '%s'", c)
		}
	}
	return timeline, nil
}

var instrumentChangeMatcher = regexp.MustCompile(`^\$(.+)$`)

func tryParseInstrumentChange(cmd string) (bool, string, error) {
	match := instrumentChangeMatcher.FindStringSubmatch(cmd)
	if match == nil {
		return false, "", nil
	}
	return true, match[1], nil
}

var useSubsectionMatcher = regexp.MustCompile(`^@(.+)$`)

func tryParseInsertSubsection(cmd string) (bool, string, error) {
	match := useSubsectionMatcher.FindStringSubmatch(cmd)
	if match == nil {
		return false, "", nil
	}
	return true, match[1], nil
}

var advanceTimelineMatcher = regexp.MustCompile(`^(.+)\+$`)

func tryParseAdvanceTimeline(cmd string) (bool, float64, error) {
	match := advanceTimelineMatcher.FindStringSubmatch(cmd)
	if match == nil {
		return false, 0, nil
	}
	dur, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return true, 0, err
	}
	return true, dur, nil
}

var playNoteMatcher = regexp.MustCompile(`^([a-zA-Z]#?)(\d+)-(.+)$`)

func tryParsePlayNote(cmd string) (bool, wave.Key, int, float64, error) {
	match := playNoteMatcher.FindStringSubmatch(cmd)
	if match == nil {
		return false, 0, 0, 0, nil
	}
	key, err := wave.ParseStringToKey(match[1])
	if err != nil {
		return true, 0, 0, 0, err
	}
	octave, err := strconv.Atoi(match[2])
	if err != nil {
		return true, 0, 0, 0, err
	}
	dur, err := strconv.ParseFloat(match[3], 64)
	if err != nil {
		return true, 0, 0, 0, err
	}
	return true, key, octave, dur, nil
}

func buildActionsFromTimeline(timeline []timelineEntry) []Action {
	slices.SortFunc(timeline, func(a, b timelineEntry) int {
		if a.time < b.time {
			return -1
		} else if a.time > b.time {
			return 1
		} else {
			return 0
		}
	})
	actions := make([]Action, 0)
	lastTime := 0.0
	for _, entry := range timeline {
		if entry.time != lastTime {
			actions = append(actions, Advance{Duration: entry.time - lastTime})
			lastTime = entry.time
		}
		actions = append(actions, entry.play)
	}
	if len(actions) > 0 {
		actions = append(actions, Advance{Duration: timeline[len(timeline)-1].play.Duration})
	}
	return actions
}

type timelineEntry struct {
	time float64
	play Play
}
