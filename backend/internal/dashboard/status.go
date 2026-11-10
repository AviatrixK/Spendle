package dashboard

// CalculateStatus evaluates net paise and returns "profit", "loss", or "break_even".
// Pure function: deterministic calculation based solely on net financial balance.
func CalculateStatus(netPaise int64) string {
	if netPaise > 0 {
		return "profit"
	}
	if netPaise < 0 {
		return "loss"
	}
	return "break_even"
}

