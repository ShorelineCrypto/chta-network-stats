package main

import (
	"testing"

	"github.com/grigata/chta-network-stats/internal/models"
)

func TestReadOptions(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		web   bool
		count int
	}{
		{"defaults", nil, false, 100}, {"console count", []string{"500"}, false, 500},
		{"web defaults", []string{"web"}, true, 100}, {"web count", []string{"WEB", "250"}, true, 250},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := readOptions(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if got.Web != test.web || got.BlockCount != test.count {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestReadOptionsRejectsInvalidInput(t *testing.T) {
	for _, args := range [][]string{{"0"}, {"-1"}, {"web", "nope"}, {"web", "1", "extra"}} {
		if _, err := readOptions(args); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}

func TestTagColumnWidthUsesLongestMinerTag(t *testing.T) {
	tags := []models.TagStats{
		{Name: "Unknown"},
		{Name: "ckpool-lhr-chta/[::H3L10SP00L::]"},
	}
	if got, want := tagColumnWidth(tags), len("ckpool-lhr-chta/[::H3L10SP00L::]"); got != want {
		t.Fatalf("width=%d, want %d", got, want)
	}
}

func TestBlockTagColumnWidthUsesLongestMinerTag(t *testing.T) {
	blocks := []models.NetworkBlock{
		{MinerTag: "Unknown"},
		{MinerTag: "ckpool/Kraskus CHTA Solo"},
	}
	if got, want := blockTagColumnWidth(blocks), len("ckpool/Kraskus CHTA Solo"); got != want {
		t.Fatalf("width=%d, want %d", got, want)
	}
}
