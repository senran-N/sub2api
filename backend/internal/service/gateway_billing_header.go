package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var ccVersionInBillingRe = regexp.MustCompile(`cc_version=\d+\.\d+\.\d+`)
var cchPlaceholderRe = regexp.MustCompile(`(x-anthropic-billing-header:[^"]*?\bcch=)(00000)(;)`)

const cchSeed uint64 = 0x6E52736AC806831E

func syncBillingHeaderVersion(body []byte, userAgent string) []byte {
	version := ExtractCLIVersion(userAgent)
	if version == "" {
		return body
	}

	system := gjson.GetBytes(body, "system")
	if !system.Exists() || !system.IsArray() {
		return body
	}

	replacement := "cc_version=" + version
	index := 0
	system.ForEach(func(_, item gjson.Result) bool {
		text := item.Get("text")
		if text.Exists() && text.Type == gjson.String && strings.HasPrefix(text.String(), "x-anthropic-billing-header") {
			updatedText := ccVersionInBillingRe.ReplaceAllString(text.String(), replacement)
			if updatedText != text.String() {
				if updatedBody, err := sjson.SetBytes(body, fmt.Sprintf("system.%d.text", index), updatedText); err == nil {
					body = updatedBody
				}
			}
		}
		index++
		return true
	})

	return body
}

func signBillingHeaderCCH(body []byte) []byte {
	if !cchPlaceholderRe.Match(body) {
		return body
	}
	cch := fmt.Sprintf("%05x", xxHash64Seeded(body, cchSeed)&0xFFFFF)
	return cchPlaceholderRe.ReplaceAll(body, []byte("${1}"+cch+"${3}"))
}

func xxHash64Seeded(data []byte, seed uint64) uint64 {
	hasher := xxhash.NewWithSeed(seed)
	_, _ = hasher.Write(data)
	return hasher.Sum64()
}
