package movie

// Movie is the common grounds for sharing data between the libraries &
// programs.
// TODO benchmark + try data oriented design
type Movie struct {
	// TMDB also supports IMDB ID:
	// https://developers.themoviedb.org/3/find/find-by-id
	// I'm treating the IMDB ID as ISBN. Silly, right?
	ImdbId     string
	Title      string
	Year       int
	YourRating int
	ImdbRating float64
}

type MoviesSummary struct {
	// TODO total per genre
	TotalCount                  int
	TopPerYear                  map[int]RankedList
	TopAllYears                 RankedList
	LeastPopularYouLikedTheMost RankedList
	MostPopularYouLikedTheLeast RankedList
}

type RankedList []Movie
