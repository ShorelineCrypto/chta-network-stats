package parser

import "testing"

func TestExtractCoinbaseTagMasterMine(t *testing.T) {
	const coinbase = "03002b4d000492e88e6a042994cf160c59e78e6a02000000000000000a4d61737465724d696e6505506f6f6c31"

	if got := ExtractCoinbaseTag(coinbase); got != "MasterMine" {
		t.Fatalf("tag=%q, want MasterMine", got)
	}
	if got := DecodeCoinbaseText(coinbase); got != "MasterMine Pool1" {
		t.Fatalf("text=%q, want MasterMine Pool1", got)
	}
}

func TestExtractCoinbaseTagDoesNotInferUnknown(t *testing.T) {
	if got := ExtractCoinbaseTag("03002b4d04aabbccdd"); got != "Unknown" {
		t.Fatalf("tag=%q, want Unknown", got)
	}
}

func TestExtractCoinbaseTagRejectsShortOrMalformedNoise(t *testing.T) {
	for _, coinbase := range []string{
		"036d284d",                     // three printable extranonce bytes: m(M
		"4dffff0a4d61737465724d696e65", // malformed OP_PUSHDATA2 must not scan its payload
		"073a74755a57537a",             // random printable extranonce bytes: :tuZWSz
	} {
		if got := ExtractCoinbaseTag(coinbase); got != "Unknown" {
			t.Fatalf("coinbase %s: tag=%q, want Unknown", coinbase, got)
		}
	}
}

func TestExtractCoinbaseTagAcceptsNewUnmappedTag(t *testing.T) {
	if got := ExtractCoinbaseTag("084272756d4272756d"); got != "BrumBrum" {
		t.Fatalf("tag=%q, want BrumBrum", got)
	}
}

func TestExtractCoinbaseTagRealWorldFormats(t *testing.T) {
	tests := []struct {
		name     string
		coinbase string
		want     string
	}{
		{
			name:     "ckpool Kraskus split tag",
			coinbase: "03952a4d00045dc58e6a040aa1e01c0c7d458a6a16000000000000000a636b706f6f6c132f4b7261736b7573204348544120536f6c6f2f",
			want:     "ckpool/Kraskus CHTA Solo",
		},
		{
			name:     "yiimp binary suffix",
			coinbase: "03d1274d0486e78d6a08810003c90200000000000000067969696d7000fabe6d6d07203eda4f7891ea57546dc2cdd23f0d2fed7965b6257be0c908c8dc7166f6c10100000000000000",
			want:     "yiimp",
		},
		{
			name:     "cgminer42 standalone push",
			coinbase: "03922a4d00fe79b72200fe9f0c0c000963676d696e6572343208390000000000000000",
			want:     "cgminer42",
		},
		{
			name:     "TinyWiny version suffix",
			coinbase: "03292b4d2f434854412d506f6f6c2d56332e372e31352f010000007e00000000000000",
			want:     "CHTA-Pool",
		},
		{
			name:     "Cheetahcoin solo miner",
			coinbase: "0302264d43686565746168636f696e20736f6c6f206d696e6572",
			want:     "Cheetahcoin solo miner",
		},
		{
			name:     "RT pool",
			coinbase: "03c8274d0488e68d6a08810003c902000000000000000a52542d504f4f4c2e4343",
			want:     "RT-POOL.CC",
		},
		{
			name:     "bowserlab binary suffix",
			coinbase: "03b7224d045e698d6a08810040d4ee2c0000626f777365726c616200fabe6d6db9bbe278261512f72d216afbe51915eca5fa5166abc3ed3ee3f42020f5fa00952000000000000000",
			want:     "bowserlab",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ExtractCoinbaseTag(test.coinbase); got != test.want {
				t.Fatalf("tag=%q, want %q", got, test.want)
			}
		})
	}
}
