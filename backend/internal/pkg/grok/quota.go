package grok

const (
	QuotaSourceDefault = "default"

	QuotaWindowAuto   = "auto"
	QuotaWindowFast   = "fast"
	QuotaWindowExpert = "expert"
	QuotaWindowHeavy  = "heavy"
)

type QuotaWindow struct {
	Remaining     int
	Total         int
	WindowSeconds int
	Source        string
}

type QuotaSet struct {
	Auto   QuotaWindow
	Fast   QuotaWindow
	Expert QuotaWindow
	Heavy  *QuotaWindow
}

func quotaWindow(total int, windowSeconds int) QuotaWindow {
	return QuotaWindow{
		Remaining:     total,
		Total:         total,
		WindowSeconds: windowSeconds,
		Source:        QuotaSourceDefault,
	}
}

func DefaultQuotaSet(tier Tier) QuotaSet {
	switch tier {
	case TierSuper:
		return QuotaSet{
			Auto:   quotaWindow(50, 7200),
			Fast:   quotaWindow(140, 7200),
			Expert: quotaWindow(50, 7200),
		}
	case TierHeavy:
		heavy := quotaWindow(20, 7200)
		return QuotaSet{
			Auto:   quotaWindow(150, 7200),
			Fast:   quotaWindow(400, 7200),
			Expert: quotaWindow(150, 7200),
			Heavy:  &heavy,
		}
	default:
		return QuotaSet{
			Auto:   quotaWindow(20, 72000),
			Fast:   quotaWindow(60, 72000),
			Expert: quotaWindow(8, 36000),
		}
	}
}

func (set QuotaSet) ToMap() map[string]any {
	result := map[string]any{
		QuotaWindowAuto:   set.Auto.toMap(),
		QuotaWindowFast:   set.Fast.toMap(),
		QuotaWindowExpert: set.Expert.toMap(),
	}
	if set.Heavy != nil {
		result[QuotaWindowHeavy] = set.Heavy.toMap()
	}
	return result
}

func (window QuotaWindow) toMap() map[string]any {
	return map[string]any{
		"remaining":      window.Remaining,
		"total":          window.Total,
		"window_seconds": window.WindowSeconds,
		"source":         window.Source,
	}
}

func InferTierFromAutoTotal(total int) Tier {
	switch total {
	case 20:
		return TierBasic
	case 50:
		return TierSuper
	case 150:
		return TierHeavy
	default:
		return TierUnknown
	}
}
