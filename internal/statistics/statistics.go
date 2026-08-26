package statistics

import (
	"sort"

	"github.com/grigata/chta-network-stats/internal/models"
)

type Statistics struct {
	TotalBlocks int

	NormalBlocks  int
	LowDiffBlocks int

	AverageGap float64
	MinGap     int64
	MaxGap     int64

	Tags map[string]int
}

func Calculate(blocks []models.NetworkBlock) Statistics {

	stats := Statistics{
		TotalBlocks: len(blocks),
		Tags:        make(map[string]int),
	}

	if len(blocks) == 0 {
		return stats
	}

	stats.MinGap = blocks[0].Gap

	var totalGap int64

	for _, block := range blocks {

		stats.Tags[block.MinerTag]++

		if block.Type == "LOW-DIFF" {
			stats.LowDiffBlocks++
		} else {
			stats.NormalBlocks++
		}

		totalGap += block.Gap

		if block.Gap < stats.MinGap {
			stats.MinGap = block.Gap
		}

		if block.Gap > stats.MaxGap {
			stats.MaxGap = block.Gap
		}
	}

	stats.AverageGap = float64(totalGap) / float64(len(blocks))

	return stats

}
func SortedTags(stats Statistics) []PoolStat {

	pools := make([]PoolStat, 0, len(stats.Tags))

	for name, blocks := range stats.Tags {

		pools = append(pools, PoolStat{
			Name:    name,
			Blocks:  blocks,
			Percent: float64(blocks) * 100 / float64(stats.TotalBlocks),
		})
	}

	sort.Slice(pools, func(i, j int) bool {

		if pools[i].Blocks == pools[j].Blocks {
			return pools[i].Name < pools[j].Name
		}

		return pools[i].Blocks > pools[j].Blocks
	})

	return pools
}
