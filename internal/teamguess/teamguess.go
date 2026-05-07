package teamguess

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var (
	lineTimePattern   = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| `)
	playerLinePattern = regexp.MustCompile(`Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)`)
	chatLinePattern   = regexp.MustCompile(`Chat\("([^"]+)"\(id=([^)]*)\)\)`)
	leadingTagPattern = regexp.MustCompile(`^\s*([\[\(\{])([^]\)\}]+)([\]\)\}])`)
	spacePattern      = regexp.MustCompile(`\s+`)
)

type Window struct {
	Start time.Duration
	End   time.Duration
}

func (w Window) Contains(clock time.Duration) bool {
	if w.End < w.Start {
		return clock >= w.Start || clock <= w.End
	}

	return clock >= w.Start && clock <= w.End
}

func ParseClock(value string) (time.Duration, error) {
	for _, layout := range []string{"15:04:05", time.RFC3339, time.RFC3339Nano} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return time.Duration(parsed.Hour())*time.Hour +
				time.Duration(parsed.Minute())*time.Minute +
				time.Duration(parsed.Second())*time.Second, nil
		}
	}

	return 0, fmt.Errorf("unsupported time format %q", value)
}

type Player struct {
	PlayerID   string
	LatestName string
	SeenNames  []string
	firstSeen  int
	lastSeen   int
}

type Member struct {
	PlayerID      string `yaml:"player_id"`
	PreferredName string `yaml:"preferred_name"`
	DisplayName   string `yaml:"display_name,omitempty"`
}

type TeamSuggestion struct {
	Name       string   `yaml:"name"`
	Confidence string   `yaml:"confidence"`
	Reason     string   `yaml:"reason"`
	Members    []Member `yaml:"members"`
}

type Output struct {
	Teams     []TeamSuggestion `yaml:"teams"`
	Ungrouped []Member         `yaml:"ungrouped"`
}

func ExtractPlayers(reader io.Reader, window Window) ([]Player, error) {
	return extractPlayers(reader, func(clock time.Duration, ok bool) bool {
		return ok && window.Contains(clock)
	})
}

func ExtractPlayersAll(reader io.Reader) ([]Player, error) {
	return extractPlayers(reader, func(_ time.Duration, ok bool) bool {
		return ok
	})
}

func extractPlayers(reader io.Reader, includeLine func(clock time.Duration, ok bool) bool) ([]Player, error) {
	scanner := bufio.NewScanner(reader)
	playersByID := map[string]*Player{}
	order := []string{}
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		clock, ok, err := parseLineClock(line)
		if err != nil {
			return nil, fmt.Errorf("parse line %d time: %w", lineNumber, err)
		}
		if !includeLine(clock, ok) {
			continue
		}

		for _, match := range playerLinePattern.FindAllStringSubmatch(line, -1) {
			recordPlayer(playersByID, &order, match[2], strings.TrimSpace(match[1]), lineNumber)
		}

		for _, match := range chatLinePattern.FindAllStringSubmatch(line, -1) {
			recordPlayer(playersByID, &order, match[2], strings.TrimSpace(match[1]), lineNumber)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	players := make([]Player, 0, len(order))
	for _, id := range order {
		players = append(players, *playersByID[id])
	}

	return players, nil
}

func recordPlayer(playersByID map[string]*Player, order *[]string, playerID string, name string, lineNumber int) {
	player, exists := playersByID[playerID]
	if !exists {
		player = &Player{
			PlayerID:   playerID,
			LatestName: name,
			SeenNames:  []string{name},
			firstSeen:  lineNumber,
			lastSeen:   lineNumber,
		}
		playersByID[playerID] = player
		*order = append(*order, playerID)
		return
	}

	player.lastSeen = lineNumber
	player.LatestName = name
	if !contains(player.SeenNames, name) {
		player.SeenNames = append(player.SeenNames, name)
	}
}

func parseLineClock(line string) (time.Duration, bool, error) {
	match := lineTimePattern.FindStringSubmatch(line)
	if len(match) == 0 {
		return 0, false, nil
	}

	clock, err := ParseClock(match[1])
	if err != nil {
		return 0, false, err
	}

	return clock, true, nil
}

type groupCandidate struct {
	DisplayName string
	Reason      string
	Confidence  string
	PlayerIDs   map[string]struct{}
	Score       int
}

func Suggest(players []Player) Output {
	groups := map[string]*groupCandidate{}

	for _, player := range players {
		if tag, ok := extractLeadingTag(player.LatestName); ok {
			addCandidate(groups, normalizeKey(tag), tag, "shared leading tag", "high", 1000, player.PlayerID)
		}
	}

	prefixCandidates := map[string]*groupCandidate{}
	for _, player := range players {
		for _, candidate := range buildPrefixCandidates(player.LatestName) {
			key := normalizeKey(candidate.DisplayName)
			existing := prefixCandidates[key]
			if existing == nil {
				existing = &groupCandidate{
					DisplayName: candidate.DisplayName,
					Reason:      candidate.Reason,
					Confidence:  candidate.Confidence,
					PlayerIDs:   map[string]struct{}{},
					Score:       candidate.Score,
				}
				prefixCandidates[key] = existing
			}

			existing.PlayerIDs[player.PlayerID] = struct{}{}
			if candidate.Score > existing.Score {
				existing.DisplayName = candidate.DisplayName
				existing.Reason = candidate.Reason
				existing.Confidence = candidate.Confidence
				existing.Score = candidate.Score
			}
		}
	}

	for key, candidate := range prefixCandidates {
		if len(candidate.PlayerIDs) < 2 && groups[key] == nil {
			continue
		}

		group := groups[key]
		if group == nil {
			group = &groupCandidate{
				DisplayName: candidate.DisplayName,
				Reason:      candidate.Reason,
				Confidence:  candidate.Confidence,
				PlayerIDs:   map[string]struct{}{},
				Score:       candidate.Score,
			}
			groups[key] = group
		}

		for playerID := range candidate.PlayerIDs {
			group.PlayerIDs[playerID] = struct{}{}
		}

		if candidate.Score > group.Score {
			group.DisplayName = candidate.DisplayName
			group.Reason = candidate.Reason
			group.Confidence = candidate.Confidence
			group.Score = candidate.Score
		}
	}

	output := Output{}
	groupedPlayers := map[string]struct{}{}
	playersByID := map[string]Player{}
	for _, player := range players {
		playersByID[player.PlayerID] = player
	}

	keys := make([]string, 0, len(groups))
	for key, group := range groups {
		if len(group.PlayerIDs) < 2 {
			continue
		}
		keys = append(keys, key)
	}
	keys = pruneBroadGroups(keys, groups)
	sort.Slice(keys, func(i, j int) bool {
		left := groups[keys[i]]
		right := groups[keys[j]]
		if left.DisplayName != right.DisplayName {
			return left.DisplayName < right.DisplayName
		}

		return len(left.PlayerIDs) > len(right.PlayerIDs)
	})

	for _, key := range keys {
		group := groups[key]
		memberIDs := sortedKeys(group.PlayerIDs)
		members := make([]Member, 0, len(memberIDs))
		for _, playerID := range memberIDs {
			player := playersByID[playerID]
			members = append(members, Member{
				PlayerID:      player.PlayerID,
				PreferredName: player.LatestName,
				DisplayName:   cleanMemberName(group.DisplayName, player.LatestName),
			})
			groupedPlayers[player.PlayerID] = struct{}{}
		}

		output.Teams = append(output.Teams, TeamSuggestion{
			Name:       group.DisplayName,
			Confidence: group.Confidence,
			Reason:     fmt.Sprintf("%s across %d players", group.Reason, len(members)),
			Members:    members,
		})
	}

	for _, player := range players {
		if _, ok := groupedPlayers[player.PlayerID]; ok {
			continue
		}
		output.Ungrouped = append(output.Ungrouped, Member{
			PlayerID:      player.PlayerID,
			PreferredName: player.LatestName,
		})
	}

	sort.Slice(output.Ungrouped, func(i, j int) bool {
		return output.Ungrouped[i].PreferredName < output.Ungrouped[j].PreferredName
	})

	return output
}

func pruneBroadGroups(keys []string, groups map[string]*groupCandidate) []string {
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		if !isBroadDuplicate(group, key, keys, groups) {
			filtered = append(filtered, key)
		}
	}

	return filtered
}

func isBroadDuplicate(target *groupCandidate, targetKey string, keys []string, groups map[string]*groupCandidate) bool {
	for playerID := range target.PlayerIDs {
		covered := false
		for _, candidateKey := range keys {
			if candidateKey == targetKey {
				continue
			}

			candidate := groups[candidateKey]
			if candidate.Score <= target.Score {
				continue
			}
			if _, ok := candidate.PlayerIDs[playerID]; ok {
				covered = true
				break
			}
		}

		if !covered {
			return false
		}
	}

	return true
}

func addCandidate(groups map[string]*groupCandidate, key string, displayName string, reason string, confidence string, score int, playerID string) {
	group := groups[key]
	if group == nil {
		group = &groupCandidate{
			DisplayName: displayName,
			Reason:      reason,
			Confidence:  confidence,
			PlayerIDs:   map[string]struct{}{},
			Score:       score,
		}
		groups[key] = group
	}

	group.PlayerIDs[playerID] = struct{}{}
	if score > group.Score {
		group.DisplayName = displayName
		group.Reason = reason
		group.Confidence = confidence
		group.Score = score
	}
}

func buildPrefixCandidates(name string) []groupCandidate {
	trimmed := strings.TrimSpace(name)
	tag, hasTag := extractLeadingTag(trimmed)
	if hasTag {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, matchedLeadingTag(trimmed)))
	}

	normalized := normalizeWords(trimmed)
	if normalized == "" {
		return nil
	}

	words := strings.Fields(normalized)
	if len(words) < 1 {
		return nil
	}

	candidates := []groupCandidate{}
	maxWords := len(words) - 1
	if maxWords > 3 {
		maxWords = 3
	}

	for wordCount := 1; wordCount <= maxWords; wordCount++ {
		prefix := strings.Join(words[:wordCount], " ")
		if !isStrongPrefix(prefix, wordCount, name) {
			continue
		}

		display := displayPrefix(name, words[:wordCount], tag)
		confidence := "medium"
		score := wordCount*100 + len(prefix)
		reason := "shared normalized prefix"
		if wordCount == 1 {
			confidence = "low"
			reason = "shared short prefix"
			score -= 25
		}

		candidates = append(candidates, groupCandidate{
			DisplayName: display,
			Reason:      reason,
			Confidence:  confidence,
			PlayerIDs:   map[string]struct{}{},
			Score:       score,
		})
	}

	return candidates
}

func extractLeadingTag(name string) (string, bool) {
	match := leadingTagPattern.FindStringSubmatch(name)
	if len(match) == 0 {
		return "", false
	}

	return strings.TrimSpace(match[2]), true
}

func matchedLeadingTag(name string) string {
	return leadingTagPattern.FindString(name)
}

func cleanMemberName(groupName string, preferredName string) string {
	trimmed := strings.TrimSpace(preferredName)
	if trimmed == "" {
		return ""
	}

	if tag, ok := extractLeadingTag(trimmed); ok && normalizeKey(tag) == normalizeKey(groupName) {
		remainder := trimNameRemainder(strings.TrimPrefix(trimmed, matchedLeadingTag(trimmed)))
		if remainder != "" {
			return remainder
		}
	}

	if remainder, ok := trimGroupPrefix(trimmed, groupName); ok {
		return remainder
	}

	return trimmed
}

func trimGroupPrefix(name string, groupName string) (string, bool) {
	groupWords := strings.Fields(normalizeWords(groupName))
	if len(groupWords) == 0 {
		return "", false
	}

	index := 0
	for _, groupWord := range groupWords {
		nextIndex, token, ok := nextNormalizedToken(name, index)
		if !ok || token != groupWord {
			return "", false
		}
		index = nextIndex
	}

	remainder := trimNameRemainder(name[index:])
	if remainder == "" {
		return "", false
	}

	return remainder, true
}

func nextNormalizedToken(value string, start int) (int, string, bool) {
	index := start
	for index < len(value) {
		r, width := utf8.DecodeRuneInString(value[index:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		index += width
	}

	if index >= len(value) {
		return start, "", false
	}

	begin := index
	for index < len(value) {
		r, width := utf8.DecodeRuneInString(value[index:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			break
		}
		index += width
	}

	return index, normalizeWords(value[begin:index]), true
}

func trimNameRemainder(value string) string {
	return strings.TrimLeftFunc(strings.TrimSpace(value), func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_' || r == ':' || r == ']' || r == '\''
	})
}

func displayPrefix(original string, words []string, tag string) string {
	if tag != "" && normalizeKey(tag) == normalizeKey(strings.Join(words, " ")) {
		return tag
	}

	if len(words) == 1 {
		firstToken := firstRawToken(original)
		if firstToken != "" && normalizeKey(firstToken) == normalizeKey(words[0]) && isUpperToken(firstToken) {
			return firstToken
		}
	}

	displayWords := make([]string, 0, len(words))
	for _, word := range words {
		displayWords = append(displayWords, strings.ToUpper(word[:1])+word[1:])
	}

	return strings.Join(displayWords, " ")
}

func isStrongPrefix(prefix string, wordCount int, original string) bool {
	if wordCount >= 2 {
		return true
	}

	if len(prefix) >= 4 {
		return true
	}

	firstToken := strings.FieldsFunc(strings.TrimSpace(original), func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_' || r == '[' || r == ']' || r == '(' || r == ')' || r == '{' || r == '}'
	})
	if len(firstToken) == 0 {
		return false
	}

	for _, r := range firstToken[0] {
		if !unicode.IsUpper(r) {
			return false
		}
	}

	return len(firstToken[0]) >= 3
}

func firstRawToken(original string) string {
	tokens := strings.FieldsFunc(strings.TrimSpace(original), func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_' || r == '[' || r == ']' || r == '(' || r == ')' || r == '{' || r == '}'
	})
	if len(tokens) == 0 {
		return ""
	}

	return tokens[0]
}

func isUpperToken(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}

	return value != ""
}

func normalizeWords(value string) string {
	var builder strings.Builder
	var last rune

	for _, r := range value {
		if unicode.IsUpper(r) && unicode.IsLower(last) {
			builder.WriteRune(' ')
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		} else {
			builder.WriteRune(' ')
		}

		last = r
	}

	return strings.TrimSpace(spacePattern.ReplaceAllString(builder.String(), " "))
}

func normalizeKey(value string) string {
	return normalizeWords(value)
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func MarshalYAML(output Output) ([]byte, error) {
	return yaml.Marshal(output)
}
