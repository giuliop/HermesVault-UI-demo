package config

const (
	// TODO: Set a cache control header where useful
	CacheControl = "public, max-age=600" // 600 sec = 10 min
)

// Protocol fees
const (
	BaseFeeDivisor = 1000    // 0.1% base fee
	MinimumFee     = 100_000 // microalgos
)
