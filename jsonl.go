//go:build linux || darwin || windows

package main

import (
	"bytes"
	"strconv"
	"time"
)

var (
	tokenEventSignature = []byte(`,"type":"event_msg","payload":{"type":"token_count"`)
	lastUsageKey        = []byte(`"last_token_usage":`)
	outputTokensKey     = []byte(`"output_tokens":`)
	reasoningTokensKey  = []byte(`"reasoning_output_tokens":`)
	timestampPrefix     = []byte(`{"timestamp":"`)
)

type tokenCount struct {
	timestamp time.Time
	output    uint64
	reasoning uint64
}

// extractTokenCount deliberately avoids unmarshalling arbitrary JSONL records. Rollout lines can
// contain very large tool outputs, while token_count records have a stable compact signature and a
// flat last_token_usage object.
func extractTokenCount(line []byte) (tokenCount, bool) {
	signatureAt := bytes.Index(line, tokenEventSignature)
	if signatureAt < 0 || signatureAt > 128 {
		return tokenCount{}, false
	}

	usageAt := bytes.Index(line[signatureAt:], lastUsageKey)
	if usageAt < 0 {
		return tokenCount{}, false
	}
	usage := line[signatureAt+usageAt+len(lastUsageKey):]
	usage = bytes.TrimLeft(usage, " \t")
	if len(usage) == 0 || usage[0] != '{' {
		return tokenCount{}, false
	}
	objectEnd := bytes.IndexByte(usage, '}')
	if objectEnd < 0 {
		return tokenCount{}, false
	}
	usage = usage[:objectEnd+1]

	output, ok := unsignedValue(usage, outputTokensKey)
	if !ok {
		return tokenCount{}, false
	}
	reasoning, ok := unsignedValue(usage, reasoningTokensKey)
	if !ok {
		return tokenCount{}, false
	}

	return tokenCount{
		timestamp: extractTimestamp(line),
		output:    output,
		reasoning: reasoning,
	}, true
}

func unsignedValue(object []byte, key []byte) (uint64, bool) {
	keyAt := bytes.Index(object, key)
	if keyAt < 0 {
		return 0, false
	}
	value := object[keyAt+len(key):]
	value = bytes.TrimLeft(value, " \t")
	end := 0
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	parsed, err := strconv.ParseUint(string(value[:end]), 10, 64)
	return parsed, err == nil
}

func extractTimestamp(line []byte) time.Time {
	if !bytes.HasPrefix(line, timestampPrefix) {
		return time.Time{}
	}
	remainder := line[len(timestampPrefix):]
	end := bytes.IndexByte(remainder, '"')
	if end < 0 {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, string(remainder[:end]))
	if err != nil {
		return time.Time{}
	}
	return parsed
}
