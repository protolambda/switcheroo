package switcher

import "fmt"

func handle(cfg *Config, em *Envelope, sourceName, targetName string) error {
	src, ok := cfg.Sources[sourceName]
	if !ok {
		return fmt.Errorf("unknown source %s", sourceName)
	}
	target, ok := cfg.Targets[targetName]
	if !ok {
		return fmt.Errorf("unknown target %s", targetName)
	}
	for _, ef := range src.Effects {
		if !ef.Direction.MatchSource() {
			continue
		}
		if em.Msg.Request != nil && !ef.Direction.MatchRequest() {
			continue
		}
		if em.Msg.Response != nil && ef.Direction.MatchResponse() {
			continue
		}
		// TODO apply effect
	}
	for _, ef := range target.Effects {
		if !ef.Direction.MatchTarget() {
			continue
		}
		if em.Msg.Request != nil && !ef.Direction.MatchRequest() {
			continue
		}
		if em.Msg.Response != nil && ef.Direction.MatchResponse() {
			continue
		}
		// TODO apply effect
	}
	return nil
}
