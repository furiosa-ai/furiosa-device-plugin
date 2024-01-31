package device_manager

import (
	"fmt"
	"regexp"
)

const (
	bdfPattern   = `(?P<domain>[0-9a-fA-F]{1,4}):(?P<bus>[0-9a-fA-F]+):(?P<function>[0-9a-fA-F]+\.[0-9])`
	subExpKeyBus = "bus"
)

var (
	bdfRegExp = regexp.MustCompile(bdfPattern)
)

func parseBusIDfromBDF(bdf string) (string, error) {
	if !bdfRegExp.MatchString(bdf) {
		return "", fmt.Errorf("couldn't parse the given string %s with bdf regex pattern: %s", bdf, bdfPattern)
	}

	matches := bdfRegExp.FindStringSubmatch(bdf)
	subExps := bdfRegExp.SubexpNames()

	namedMatches := map[string]string{}
	for i, match := range matches {
		subExp := subExps[i]
		if subExp == "" {
			continue
		}
		namedMatches[subExp] = match
	}

	busID, ok := namedMatches[subExpKeyBus]
	if !ok {
		return "", fmt.Errorf("couldn't parse bus id from the given bdf expression %s", bdf)
	}

	return busID, nil
}
