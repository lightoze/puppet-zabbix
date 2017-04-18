package main

import (
	"strings"
	"regexp"
	"strconv"
	"sort"
	"os"
	"fmt"
)

func NonEmpty(args []string) (ret []string) {
	for _, arg := range args {
		if len(arg) > 0 {
			ret = append(ret, arg)
		}
	}
	return
}

func SplitTwo(str string, delim string) (string, string) {
	parts := strings.SplitN(str, delim, 2)
	if len(parts) < 2 {
		return str, ""
	} else {
		return parts[0], parts[1]
	}
}

func NormalizePath(path string) string {
	domain, path := SplitTwo(path, ":")
	path, attribute := SplitTwo(path, "/")

	tags := strings.Split(path, ",")
	sort.Strings(tags)
	path = strings.Join(tags, ",")

	path = domain + ":" + path
	if len(attribute) > 0 {
		path += "/" + attribute
	}
	return path
}

func ZabbixEscape(str string) string {
	return strings.NewReplacer(
		`%`, `%%`,
		`+`, `%+`,
		`,`, `+`,
		"`", "%60",
		`\`, `%5C`,
		`'`, `%27`,
		`"`, `%22`,
		`*`, `%2A`,
		`?`, `%3F`,
		`[`, `%5B`,
		`]`, `%5D`,
		`{`, `%7B`,
		`}`, `%7D`,
		`~`, `%7E`,
		`$`, `%24`,
		`!`, `%21`,
		`&`, `%26`,
		`;`, `%3B`,
		`(`, `%28`,
		`)`, `%29`,
		`<`, `%3C`,
		`>`, `%3E`,
		`|`, `%7C`,
		`#`, `%23`,
		`@`, `%40`,
		"\n", `%0D`,
	).Replace(str)
}

func ZabbixUnescape(str string) string {
	return regexp.MustCompile(`\+|%([%+]|[0-9A-Fa-f]{2})?`).ReplaceAllStringFunc(str, func(s string) string {
		if s == `+` {
			return `,`
		} else if s == `%` {
			return `*`
		} else {
			s = s[1:]
			if len(s) == 1 {
				return s
			} else {
				ascii, err := strconv.ParseInt(s, 16, 0)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return string(byte(ascii))
			}
		}
	})
}
