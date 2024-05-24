package main

import (
	"cmp"
	"fmt"
	"slices"
	"testing"
)

func TestMatchEquals(t *testing.T) {
	filter := Filter{
		Key:      "test",
		Operator: "=",
		Value:    "works",
	}

	if !filter.Match(map[string]string{
		"test": "works",
	}) {
		t.Error("Failed the simple positive match!")
	}

	if !filter.Match(map[string]string{
		"test":   "works",
		"second": "example",
	}) {
		t.Error("Failed positive match with two keys!")
	}

	if filter.Match(map[string]string{
		"unrelated": "keys",
		"don't":     "match",
	}) {
		t.Error("Failed simple negative case")
	}

	if filter.Match(map[string]string{
		"test": "fails",
	}) {
		t.Error("Failed key exists negative case")
	}

	if filter.Match(map[string]string{
		"test": "works (not)",
	}) {
		t.Error("Failed exact match check")
	}
}

func TestMatchNotEquals(t *testing.T) {
	filter := Filter{
		Key:      "test",
		Operator: "!=",
		Value:    "works",
	}

	if filter.Match(map[string]string{
		"test": "works",
	}) {
		t.Error("Failed the simple positive match!")
	}

	if filter.Match(map[string]string{
		"test":   "works",
		"second": "example",
	}) {
		t.Error("Failed positive match with two keys!")
	}

	if filter.Match(map[string]string{
		"unrelated": "keys",
		"don't":     "match",
	}) {
		t.Error("Missing key does not match")
	}

	if !filter.Match(map[string]string{
		"test": "fails",
	}) {
		t.Error("Failed key exists negative case")
	}

	if !filter.Match(map[string]string{
		"test": "works (not)",
	}) {
		t.Error("Failed exact match check")
	}
}

func TestMatchLike(t *testing.T) {
	filter := Filter{
		Key:      "test",
		Operator: "~",
		Value:    "ork",
	}

	if !filter.Match(map[string]string{
		"test": "works",
	}) {
		t.Error("Failed the simple positive match!")
	}

	if !filter.Match(map[string]string{
		"test":   "works",
		"second": "example",
	}) {
		t.Error("Failed positive match with two keys!")
	}

	if filter.Match(map[string]string{
		"unrelated": "keys",
		"don't":     "match",
	}) {
		t.Error("Missing key does not match")
	}

	if filter.Match(map[string]string{
		"test": "fails",
	}) {
		t.Error("Failed key exists negative case")
	}

	if !filter.Match(map[string]string{
		"test": "works (not)",
	}) {
		t.Error("Failed exact match check")
	}
}

func TestMatchNotLike(t *testing.T) {
	filter := Filter{
		Key:      "test",
		Operator: "!~",
		Value:    "ork",
	}

	if filter.Match(map[string]string{
		"test": "works",
	}) {
		t.Error("Failed the simple positive match!")
	}

	if filter.Match(map[string]string{
		"test":   "works",
		"second": "example",
	}) {
		t.Error("Failed positive match with two keys!")
	}

	if filter.Match(map[string]string{
		"unrelated": "keys",
		"don't":     "match",
	}) {
		t.Error("Missing key does not match")
	}

	if !filter.Match(map[string]string{
		"test": "fails",
	}) {
		t.Error("Failed key exists negative case")
	}

	if filter.Match(map[string]string{
		"test": "works (not)",
	}) {
		t.Error("Failed exact match check")
	}
}

func compareFilters(t *testing.T, expected, actual []Filter) {
	if len(expected) != len(actual) {
		t.Errorf("Got %v filters, expected %v", len(actual), len(expected))
	}

	// sort both as order doesn't matter
	slices.SortFunc(actual, func(a, b Filter) int {
		return cmp.Compare(a.Key, b.Key)
	})

	slices.SortFunc(expected, func(a, b Filter) int {
		return cmp.Compare(a.Key, b.Key)
	})

	for i, _ := range actual {
		if actual[i] != expected[i] {
			t.Errorf("Expected %+v, got %+v", expected[i], actual[i])
		}
	}
}

func TestReadFilter(t *testing.T) {
	// simple parsing
	for _, op := range []string{"=", "!=", "~", "!~"} {
		compareFilters(
			t,
			readFilter(fmt.Sprintf(`key%svalue`, op)),
			[]Filter{
				{
					Key:      "key",
					Operator: op,
					Value:    "value",
				},
			},
		)
	}

	// no-op parsing
	for _, op := range []string{"+", "-"} {
		compareFilters(
			t,
			readFilter(fmt.Sprintf("key%s", op)),
			[]Filter{
				{
					Key:      "key",
					Operator: op,
					Value:    "",
				},
			},
		)
	}

	// multiple filters
	compareFilters(
		t,
		readFilter(`key=value other~ex absent- present+`),
		[]Filter{
			{
				Key:      "key",
				Operator: "=",
				Value:    "value",
			},
			{
				Key:      "other",
				Operator: "~",
				Value:    "ex",
			},
			{
				Key:      "absent",
				Operator: "-",
				Value:    "",
			},
			{
				Key:      "present",
				Operator: "+",
				Value:    "",
			},
		},
	)

	// quoted values
	compareFilters(
		t,
		readFilter(`key=value other="longer value" last~value`),
		[]Filter{
			{
				Key:      "key",
				Operator: "=",
				Value:    "value",
			},
			{
				Key:      "other",
				Operator: "=",
				Value:    "longer value",
			},
			{
				Key:      "last",
				Operator: "~",
				Value:    "value",
			},
		},
	)
}

func TestMatchesFilter(t *testing.T) {
	// postiive match
	if !matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "=",
				Value:    "value",
			},
			{
				Key:      "other",
				Operator: "~",
				Value:    "al",
			},
		},
		map[string]string{
			"key":        "value",
			"other":      "value",
			"irrelevant": "value",
		},
	) {
		t.Error("Failed positive case!")
	}

	// negative match
	if matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "=",
				Value:    "value",
			},
			{
				Key:      "other",
				Operator: "~",
				Value:    "al",
			},
		},
		map[string]string{
			"missing": "keys",
		},
	) {
		t.Error("Failed negative case!")
	}

	// present match
	if !matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "+",
			},
		},
		map[string]string{
			"key":        "present",
			"irrelevant": "key",
		},
	) {
		t.Error("Presence filter doesn't match as expected")
	}

	if matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "+",
			},
		},
		map[string]string{
			"other-key":  "present",
			"irrelevant": "key",
		},
	) {
		t.Error("Presence filter doesn't not match as expected")
	}

	// absent match
	if !matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "-",
			},
		},
		map[string]string{
			"other-key":  "present",
			"irrelevant": "key",
		},
	) {
		t.Error("Absence filter doesn't match as expected")
	}

	if matchesFilter(
		[]Filter{
			{
				Key:      "key",
				Operator: "-",
			},
		},
		map[string]string{
			"key":        "present",
			"irrelevant": "key",
		},
	) {
		t.Error("Absence filter doesn't not match as expected")
	}
}
