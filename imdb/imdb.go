// Takes an IMDB export in the csv format and answers pre-defined questions
// based on that data.
package imdb

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"

	"github.com/chelmertz/movie-summary/movie"

	_ "github.com/mattn/go-sqlite3"
)

// TODO handle errors
func todo(err error) {
	if err != nil {
		panic(err)
	}
}

// imdbQueries is an attempt at avoiding repeating if-err checks. I'm trying to
// not go down the Result/Either/... rabbit hole.
type imdbQueries struct {
	db      *sql.DB
	err     error
	summary *movie.MoviesSummary
}

// NewFromCsv creates an in-memory sqlite session to go over the data.
func NewFromCsv(f io.Reader) (*movie.MoviesSummary, error) {
	db, err := setupDb()
	todo(err)
	defer db.Close()

	r := csv.NewReader(f)
	err = insertToDbFromCsv(db, r)
	todo(err)

	is := &imdbQueries{
		db:      db,
		err:     nil,
		summary: &movie.MoviesSummary{},
	}
	is.totalCount()
	is.topPerYear()
	is.topAllYears()
	is.leastPopularYouLikedTheMost()
	is.mostPopularYouLikedTheLeast()

	return is.summary, is.err
}

func (is *imdbQueries) totalCount() *imdbQueries {
	if is.err != nil {
		return is
	}

	var totalCount int
	row := is.db.QueryRow("select count(*) from imdb")
	if err := row.Err(); err != nil {
		// TODO assign errors like this to is.err
		panic(err)
	}
	if err := row.Scan(&totalCount); err != nil {
		panic(err)
	}
	is.summary.TotalCount = totalCount
	return is
}

func (is *imdbQueries) topPerYear() *imdbQueries {
	if is.err != nil {
		return is
	}

	topPerYear := make(map[int]movie.RankedList)
	// TODO verify that row_number() is what I want
	rows, err := is.db.Query(`select
	imdb_id, your_rating, imdb_rating, title, year
	from (
		select imdb_id, your_rating, imdb_rating, title, year,
		row_number() over(partition by year order by your_rating desc) row_number
		from imdb
		where title_type = 'movie'
	) t
	where
	t.row_number <= 10`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		// TODO this cannot be the cheapest way of doing things
		m := movie.Movie{}
		if err := rows.Scan(&m.ImdbId, &m.YourRating, &m.ImdbRating, &m.Title, &m.Year); err != nil {
			panic(err)
		}

		topPerYear[m.Year] = append(topPerYear[m.Year], m)
	}
	is.summary.TopPerYear = topPerYear
	return is
}

func (is *imdbQueries) topAllYears() *imdbQueries {
	if is.err != nil {
		return is
	}

	topAllYears := make(movie.RankedList, 0)
	rows, err := is.db.Query(`select
	imdb_id, your_rating, imdb_rating, title, year
	from imdb
	where title_type = 'movie'
	order by your_rating desc
	limit 10`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		m := movie.Movie{}
		if err := rows.Scan(&m.ImdbId, &m.YourRating, &m.ImdbRating, &m.Title, &m.Year); err != nil {
			panic(err)
		}

		topAllYears = append(topAllYears, m)
	}
	is.summary.TopAllYears = topAllYears
	return is
}

func (is *imdbQueries) leastPopularYouLikedTheMost() *imdbQueries {
	if is.err != nil {
		return is
	}

	leastPopularYouLikedTheMost := make(movie.RankedList, 0)
	rows, err := is.db.Query(`select
	imdb_id, your_rating, imdb_rating, title, year
	from imdb
	where title_type = 'movie'
	order by (your_rating - imdb_rating) desc
	limit 10`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		m := movie.Movie{}
		if err := rows.Scan(&m.ImdbId, &m.YourRating, &m.ImdbRating, &m.Title, &m.Year); err != nil {
			panic(err)
		}

		leastPopularYouLikedTheMost = append(leastPopularYouLikedTheMost, m)
	}
	is.summary.LeastPopularYouLikedTheMost = leastPopularYouLikedTheMost
	return is
}

func (is *imdbQueries) mostPopularYouLikedTheLeast() *imdbQueries {
	if is.err != nil {
		return is
	}

	mostPopularYouLikedTheLeast := make(movie.RankedList, 0)
	rows, err := is.db.Query(`select
	imdb_id, your_rating, imdb_rating, title, year
	from imdb
	where title_type = 'movie'
	order by (imdb_rating - your_rating) desc
	limit 10`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		m := movie.Movie{}
		if err := rows.Scan(&m.ImdbId, &m.YourRating, &m.ImdbRating, &m.Title, &m.Year); err != nil {
			panic(err)
		}

		mostPopularYouLikedTheLeast = append(mostPopularYouLikedTheLeast, m)
	}
	is.summary.MostPopularYouLikedTheLeast = mostPopularYouLikedTheLeast
	return is
}

func insertToDbFromCsv(db *sql.DB, reader *csv.Reader) error {
	// discard header row
	_, err := reader.Read()
	todo(err)

	addMovie, err := db.Prepare(`insert into imdb
(imdb_id,
your_rating,
date_rated,
title,
url,
title_type,
imdb_rating,
runtime_mins,
year,
genres,
num_votes,
release_date,
directors
) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	todo(err)

	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			todo(err)
		}
		_, err = addMovie.Exec(row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7], row[8], row[9], row[10], row[11], row[12])
		todo(err)
	}
	return nil
}

func setupDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	todo(err)

	err = db.Ping()
	todo(err)

	_, err = db.Exec(`create table if not exists imdb
	(imdb_id text,
 your_rating integer,
 date_rated text,
 title text,
 url text,
 title_type text,
 imdb_rating real,
 runtime_mins integer,
 year integer,
 genres string,
 num_votes integer,
 release_date text,
 directors text
)
	`)
	todo(err)

	return db, err
}
