package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/grigata/chta-network-stats/internal/chain"
	"github.com/grigata/chta-network-stats/internal/models"
	"github.com/grigata/chta-network-stats/internal/parser"
)

type Scanner struct {
	client chain.Client
}
type ProgressFunc func(current, total int)

func New(client chain.Client) *Scanner {
	return &Scanner{
		client: client,
	}
}

func (s *Scanner) ReadLastBlocks(
	ctx context.Context,
	count int,
	progress ProgressFunc,
) ([]models.NetworkBlock, error) {
	if count <= 0 {
		return nil, fmt.Errorf("block count must be greater than zero")
	}

	latestHeight, err := s.client.GetBlockCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest block height: %w", err)
	}

	if int64(count) > latestHeight {
		count = int(latestHeight)
	}

	latestHash, err := s.client.GetBlockHash(ctx, latestHeight)
	if err != nil {
		return nil, fmt.Errorf(
			"get latest block hash at height %d: %w",
			latestHeight,
			err,
		)
	}

	currentBlock, err := s.client.GetBlock(ctx, latestHash)
	if err != nil {
		return nil, fmt.Errorf(
			"get latest block at height %d: %w",
			latestHeight,
			err,
		)
	}

	blocks := make([]models.NetworkBlock, 0, count)

	for index := 0; index < count; index++ {
		height := latestHeight - int64(index)

		var (
			gap           int64
			previousBlock = currentBlock
		)

		if height > 0 {
			if currentBlock.PreviousBlockHash == "" {
				return nil, fmt.Errorf(
					"block at height %d has no previous block hash",
					height,
				)
			}

			previousBlock, err = s.client.GetBlock(
				ctx,
				currentBlock.PreviousBlockHash,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"get previous block for height %d: %w",
					height,
					err,
				)
			}

			gap = currentBlock.Time - previousBlock.Time
		}

		coinbaseTxID := ""
		coinbaseHex := ""
		minerTag := "Unknown"

		if len(currentBlock.Tx) > 0 {
			coinbaseTxID = currentBlock.Tx[0]

			tx, err := s.client.GetRawTransaction(ctx, coinbaseTxID)
			if err != nil {
				return nil, fmt.Errorf(
					"get coinbase transaction at height %d: %w",
					height,
					err,
				)
			}

			if len(tx.Vin) > 0 {
				coinbaseHex = tx.Vin[0].Coinbase
				minerTag = parser.ExtractCoinbaseTag(coinbaseHex)
			}
		}

		blocks = append(blocks, models.NetworkBlock{
			Height:       currentBlock.Height,
			Hash:         currentBlock.Hash,
			Difficulty:   currentBlock.Difficulty,
			Bits:         currentBlock.Bits,
			Time:         time.Unix(currentBlock.Time, 0).Local(),
			Gap:          gap,
			TxCount:      len(currentBlock.Tx),
			Type:         blockType(currentBlock.Difficulty),
			CoinbaseTxID: coinbaseTxID,
			CoinbaseHex:  coinbaseHex,
			MinerTag:     minerTag,
		})
		if progress != nil {
			progress(index+1, count)
		}

		currentBlock = previousBlock
	}

	return blocks, nil
}

func blockType(difficulty float64) string {
	if difficulty < 1 {
		return "LOW-DIFF"
	}

	return "NORMAL"
}
