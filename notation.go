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
		args := strings.Split(c, "-")
		switch args[0] {
		case "p":
			if len(args) != 4 {
				return nil, errors.New("all play commands must have 4 args")
			}
			key, err := wave.ParseStringToKey(args[1])
			if err != nil {
				return nil, err
			}
			octave, err := strconv.Atoi(args[2])
			if err != nil {
				return nil, err
			}
			dur, err := strconv.ParseFloat(args[3], 64)
			if err != nil {
				return nil, err
			}
			timeline = append(timeline, timelineEntry{now, Play{Key: key, Octave: octave, Duration: dur / (bpm / 60.0), Instrument: instrument}})
		case "a":
			if len(args) != 2 {
				return nil, errors.New("all advance commands must have 2 args, got " + c)
			}
			t, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return nil, err
			}
			now += t / (bpm / 60.0)
		case "i":
			if len(args) != 2 {
				return nil, errors.New("all insert commands must have 2 args")
			}
			name := args[1]
			tl, err := addAdvancedToTimeline(now, instrument, timeline, name, blocks, bpm)
			if err != nil {
				return nil, err
			}
			timeline = tl
		case "t":
			if len(args) != 2 {
				return nil, errors.New("all insert commands must have 2 args")
			}
			instrument = args[1]
		}
	}
	return timeline, nil
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
