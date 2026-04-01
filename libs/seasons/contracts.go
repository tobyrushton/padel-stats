package seasons

type Season struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate" format:"date"`
	EndDate   string `json:"endDate" format:"date"`
}

type CreateSeasonInput struct {
	Name      string `json:"name" validate:"required"`
	StartDate string `json:"startDate" format:"date" validate:"required"`
}
