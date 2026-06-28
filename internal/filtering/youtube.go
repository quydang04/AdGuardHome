package filtering

import (
	"context"
	"strings"
)

const youtubeRuleMarker = "! [YouTube-managed]"

// AddRewrite adds a single DNS rewrite entry and persists it.
func (d *DNSFilter) AddRewrite(ctx context.Context, rw *LegacyRewrite) {
	d.confMu.Lock()
	defer d.confMu.Unlock()

	d.conf.Rewrites = append(d.conf.Rewrites, rw)
	d.logger.DebugContext(ctx, "youtube: added rewrite", "domain", rw.Domain, "answer", rw.Answer)
}

// RemoveRewritesByDomains removes all DNS rewrite entries whose Domain matches
// any of the given domains.
func (d *DNSFilter) RemoveRewritesByDomains(ctx context.Context, domains []string) {
	domainSet := make(map[string]bool, len(domains))
	for _, dom := range domains {
		domainSet[strings.ToLower(dom)] = true
	}

	d.confMu.Lock()
	defer d.confMu.Unlock()

	kept := make([]*LegacyRewrite, 0, len(d.conf.Rewrites))
	removed := 0
	for _, rw := range d.conf.Rewrites {
		if domainSet[strings.ToLower(rw.Domain)] {
			removed++

			continue
		}

		kept = append(kept, rw)
	}

	d.conf.Rewrites = kept

	if removed > 0 {
		d.logger.DebugContext(ctx, "youtube: removed rewrites", "count", removed)
	}
}

// AddYouTubeRules adds YouTube-managed blocking rules to the user rules list.
// Each rule is prefixed with the YouTube rule marker for later identification.
func (d *DNSFilter) AddYouTubeRules(ctx context.Context, rules []string) {
	d.RemoveYouTubeRules(ctx)

	d.conf.UserRules = append(d.conf.UserRules, youtubeRuleMarker)
	d.conf.UserRules = append(d.conf.UserRules, rules...)

	d.conf.ConfModifier.Apply(ctx)
	d.EnableFilters(true)

	d.logger.DebugContext(ctx, "youtube: added blocking rules", "count", len(rules))
}

// RemoveYouTubeRules removes all YouTube-managed rules from the user rules list.
func (d *DNSFilter) RemoveYouTubeRules(ctx context.Context) {
	kept := make([]string, 0, len(d.conf.UserRules))
	inYouTube := false

	for _, rule := range d.conf.UserRules {
		if rule == youtubeRuleMarker {
			inYouTube = true

			continue
		}

		if inYouTube {
			if strings.HasPrefix(rule, "||") && strings.HasSuffix(rule, "^") {
				continue
			}

			inYouTube = false
		}

		kept = append(kept, rule)
	}

	d.conf.UserRules = kept
	d.conf.ConfModifier.Apply(ctx)
	d.EnableFilters(true)
}
