package domain

import "math"

// PriceMinorUnits converts DB float price to proto int64 kopecks.
// PriceMinorUnits переводит float цены из БД в int64 копеек для proto.
func PriceMinorUnits(price float64) int64 {
	return int64(math.Round(price * 100))
}
