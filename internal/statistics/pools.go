package statistics

import (
	"sort"

	"github.com/grigata/chta-network-stats/internal/models"
)

func AnalyzeTags(blocks []models.NetworkBlock) []models.TagStats {

	tags := make(map[string]*models.TagStats)

	for _, block := range blocks {

		ps, ok := tags[block.MinerTag]
		if !ok {
			ps = &models.TagStats{
				Name: block.MinerTag,
			}
			tags[block.MinerTag] = ps
		}

		ps.TotalBlocks++

		switch block.Type {
		case "NORMAL":
			ps.NormalBlocks++

		case "LOW-DIFF":
			ps.LowDiffBlocks++
		}
	}

	result := make([]models.TagStats, 0, len(tags))

	for _, tag := range tags {
		tag.Percent =
			float64(tag.TotalBlocks) * 100 /
				float64(len(blocks))

		result = append(result, *tag)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalBlocks > result[j].TotalBlocks
	})

	return result
}
