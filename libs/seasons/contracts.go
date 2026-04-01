package seasons

type Season struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate" format:"date"`
	EndDate   string `json:"endDate" format:"date"`
}
