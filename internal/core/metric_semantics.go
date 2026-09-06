package core

func MetricUsedPercent(key string, m Metric) float64 {
	if m.Unit == "%" && m.Used != nil {
		return *m.Used
	}
	if m.Unit == "%" && m.Remaining != nil {
		return 100 - *m.Remaining
	}
	if m.Limit != nil && m.Remaining != nil && *m.Limit > 0 {
		return (*m.Limit - *m.Remaining) / *m.Limit * 100
	}
	if m.Limit != nil && m.Used != nil && *m.Limit > 0 {
		return *m.Used / *m.Limit * 100
	}
	return -1
}
