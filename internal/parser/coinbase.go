package parser

import (
	"encoding/binary"
	"encoding/hex"
	"regexp"
	"strings"
)

const unknownTag = "Unknown"

const (
	minTagLength  = 4
	minTagLetters = 3
)

var versionSuffix = regexp.MustCompile(`(?i)-v[0-9]+(?:\.[0-9]+)+$`)

// ExtractCoinbaseTag returns the most useful human-readable tag embedded in a
// coinbase scriptSig. This is a miner-supplied tag, not a claimed or inferred
// pool identity. Some miners put their tag in a conventional script push, while
// others append it to a binary extranonce field; both forms are supported.
func ExtractCoinbaseTag(coinbaseHex string) string {
	data, err := hex.DecodeString(coinbaseHex)
	if err != nil {
		return unknownTag
	}

	candidates := make([]string, 0)
	for _, pushed := range coinbasePushes(data) {
		if tag := printableTag(pushed); tag != "" {
			candidates = append(candidates, normalizeTag(tag))
		}
	}
	if hasBIP34HeightPrefix(data) {
		candidates = append(candidates, embeddedTags(data)...)
	}

	return bestTag(candidates)
}

// DecodeCoinbaseText is retained for diagnostics. It exposes all printable
// pushed strings, whereas ExtractCoinbaseTag selects the primary tag used for
// display and statistics.
func DecodeCoinbaseText(coinbaseHex string) string {
	data, err := hex.DecodeString(coinbaseHex)
	if err != nil {
		return ""
	}

	parts := make([]string, 0)
	for _, pushed := range coinbasePushes(data) {
		if tag := printableTag(pushed); tag != "" {
			parts = append(parts, tag)
		}
	}
	return strings.Join(parts, " ")
}

func coinbasePushes(script []byte) [][]byte {
	pushes := make([][]byte, 0)
	for offset := 0; offset < len(script); {
		opcode := script[offset]
		offset++

		length, ok := pushLength(script, &offset, opcode)
		if !ok || length < 0 || length > len(script)-offset {
			// A malformed push means that treating its payload bytes as opcodes
			// would create random-looking false tags from binary extranonce data.
			break
		}
		pushes = append(pushes, script[offset:offset+length])
		offset += length
	}
	return pushes
}

func pushLength(script []byte, offset *int, opcode byte) (int, bool) {
	switch {
	case opcode == 0x00: // OP_0 is a valid empty push.
		return 0, true
	case opcode >= 1 && opcode <= 75:
		return int(opcode), true
	case opcode == 0x4c: // OP_PUSHDATA1
		if *offset >= len(script) {
			return 0, false
		}
		length := int(script[*offset])
		*offset += 1
		return length, true
	case opcode == 0x4d: // OP_PUSHDATA2
		if len(script)-*offset < 2 {
			return 0, false
		}
		length := int(binary.LittleEndian.Uint16(script[*offset : *offset+2]))
		*offset += 2
		return length, true
	case opcode == 0x4e: // OP_PUSHDATA4
		if len(script)-*offset < 4 {
			return 0, false
		}
		length64 := uint64(binary.LittleEndian.Uint32(script[*offset : *offset+4]))
		*offset += 4
		if length64 > uint64(len(script)-*offset) {
			return 0, false
		}
		return int(length64), true
	default:
		return 0, false
	}
}

func printableTag(data []byte) string {
	return printableTagWithStart(data, isTagStart)
}

func printableTagOrContinuation(data []byte) string {
	return printableTagWithStart(data, func(b byte) bool {
		return isTagStart(b) || b == '/'
	})
}

func printableTagWithStart(data []byte, allowedStart func(byte) bool) string {
	if len(data) < minTagLength {
		return ""
	}
	if !allowedStart(data[0]) {
		return ""
	}
	letters := 0
	for _, b := range data {
		if !isTagByte(b) {
			return ""
		}
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
			letters++
		}
	}
	if letters < minTagLetters {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// embeddedTags handles miners which append a text tag inside or beside binary
// extranonce data instead of encoding it as a standalone script push. It is
// intentionally limited to scripts that begin with a valid BIP34 height, so a
// malformed arbitrary script cannot turn an embedded printable payload into a
// false miner tag.
func embeddedTags(script []byte) []string {
	start := 1 + int(script[0])
	runs := make([]textRun, 0)
	for index := start; index < len(script); {
		if !isTagByte(script[index]) {
			index++
			continue
		}

		runStart := index
		for index < len(script) && isTagByte(script[index]) {
			index++
		}
		if tag := printableTagOrContinuation(script[runStart:index]); tag != "" {
			runs = append(runs, textRun{start: runStart, end: index, tag: tag})
		}
	}

	tags := make([]string, 0, len(runs))
	for index, run := range runs {
		tags = append(tags, run.tag)
		if index+1 >= len(runs) {
			continue
		}
		next := runs[index+1]
		// Some coinbases split a tag with a one-byte push marker and continue
		// with a slash-delimited suffix.
		if next.start-run.end == 1 && script[run.end] <= 75 &&
			endsAlphaNumeric(run.tag) && strings.HasPrefix(next.tag, "/") {
			tags = append(tags, normalizeTag(run.tag+next.tag))
		}
	}
	return tags
}

type textRun struct {
	start int
	end   int
	tag   string
}

func hasBIP34HeightPrefix(script []byte) bool {
	return len(script) > 1 && script[0] >= 1 && script[0] <= 5 && int(script[0]) <= len(script)-1
}

func bestTag(candidates []string) string {
	best := ""
	for _, candidate := range candidates {
		candidate = normalizeTag(candidate)
		if candidate == "" {
			continue
		}
		if len(candidate) > len(best) {
			best = candidate
		}
	}
	if best == "" {
		return unknownTag
	}
	return best
}

func normalizeTag(tag string) string {
	tag = strings.Trim(strings.TrimSpace(tag), "/")
	tag = versionSuffix.ReplaceAllString(tag, "")
	return strings.TrimSpace(tag)
}

func isTagByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || strings.ContainsRune(" ._-/:[]<>", rune(b))
}

func isTagStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '['
}

func endsAlphaNumeric(tag string) bool {
	if tag == "" {
		return false
	}
	last := tag[len(tag)-1]
	return (last >= 'a' && last <= 'z') || (last >= 'A' && last <= 'Z') || (last >= '0' && last <= '9')
}
