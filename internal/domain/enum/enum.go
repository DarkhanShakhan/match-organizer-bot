package enum

type SportType string

const (
	SportTypeFootbal    SportType = "football"
	SportTypeVolleyball SportType = "volleyball"
	SportTypeBasketball SportType = "basketball"
)

type MatchDay string

const (
	MatchDayToday    MatchDay = "today"
	MatchDayTomorrow MatchDay = "tomorrow"
)
