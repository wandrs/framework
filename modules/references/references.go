// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package references

import (
	"regexp"

	"code.gitea.io/gitea/modules/markup/mdstripper"
)

var (
	// NOTE: All below regex matching do not perform any extra validation.
	// Thus a link is produced even if the linked entity does not exist.
	// While fast, this is also incorrect and lead to false positives.
	// TODO: fix invalid linking issue

	// mentionPattern matches all mentions in the form of "@user" or "@org/team"
	mentionPattern = regexp.MustCompile(`(?:\s|^|\(|\[)(@[0-9a-zA-Z-_]+|@[0-9a-zA-Z-_]+\/?[0-9a-zA-Z-_]+|@[0-9a-zA-Z-_][0-9a-zA-Z-_.]+\/?[0-9a-zA-Z-_.]+[0-9a-zA-Z-_])(?:\s|[:,;.?!]\s|[:,;.?!]?$|\)|\])`)
	// spaceTrimmedPattern let's us find the trailing space
	spaceTrimmedPattern = regexp.MustCompile(`(?:.*[0-9a-zA-Z-_])\s`)
)

// RefSpan is the position where the reference was found within the parsed text
type RefSpan struct {
	Start int
	End   int
}

// FindAllMentionsMarkdown matches mention patterns in given content and
// returns a list of found unvalidated user names **not including** the @ prefix.
func FindAllMentionsMarkdown(content string) []string {
	bcontent, _ := mdstripper.StripMarkdownBytes([]byte(content))
	locations := FindAllMentionsBytes(bcontent)
	mentions := make([]string, len(locations))
	for i, val := range locations {
		mentions[i] = string(bcontent[val.Start+1 : val.End])
	}
	return mentions
}

// FindAllMentionsBytes matches mention patterns in given content
// and returns a list of locations for the unvalidated user names, including the @ prefix.
func FindAllMentionsBytes(content []byte) []RefSpan {
	// Sadly we can't use FindAllSubmatchIndex because our pattern checks for starting and
	// trailing spaces (\s@mention,\s), so if we get two consecutive references, the space
	// from the second reference will be "eaten" by the first one:
	// ...\s@mention1\s@mention2\s...	--> ...`\s@mention1\s`, (not) `@mention2,\s...`
	ret := make([]RefSpan, 0, 5)
	pos := 0
	for {
		match := mentionPattern.FindSubmatchIndex(content[pos:])
		if match == nil {
			break
		}
		ret = append(ret, RefSpan{Start: match[2] + pos, End: match[3] + pos})
		notrail := spaceTrimmedPattern.FindSubmatchIndex(content[match[2]+pos : match[3]+pos])
		if notrail == nil {
			pos = match[3] + pos
		} else {
			pos = match[3] + pos + notrail[1] - notrail[3]
		}
	}
	return ret
}

// FindFirstMentionBytes matches the first mention in then given content
// and returns the location of the unvalidated user name, including the @ prefix.
func FindFirstMentionBytes(content []byte) (bool, RefSpan) {
	mention := mentionPattern.FindSubmatchIndex(content)
	if mention == nil {
		return false, RefSpan{}
	}
	return true, RefSpan{Start: mention[2], End: mention[3]}
}
