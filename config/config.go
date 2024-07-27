package config

const (
	// TODO: Set a cache control header where useful
	CacheControl = "public, max-age=600" // 600 sec = 10 min
    ProductionPort = "5555"
    DevelopmentPort = "3000"
)

// Protocol fees
const (
	BaseFeeDivisor = 1000    // 0.1% base fee
	MinimumFee     = 100_000 // microalgos
)

// Number of characters to highlight displaying long strings, e.g. addresses
const (
	NumCharsToHighlight = 5
)
