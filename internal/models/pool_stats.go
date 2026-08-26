package models

type TagStats struct {
	Name          string
	TotalBlocks   int
	NormalBlocks  int
	LowDiffBlocks int
	Percent       float64
}
