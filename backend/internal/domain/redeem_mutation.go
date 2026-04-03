package domain

type GenerateRedeemCodesInput struct {
	Count        int
	Type         string
	Value        float64
	GroupID      *int64
	ValidityDays int
}

type GenerateCodesRequest struct {
	Count int
	Value float64
	Type  string
}
