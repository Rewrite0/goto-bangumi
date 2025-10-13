package model

// ShowInfo represents a TV show search result from TMDB
type ShowInfo struct {
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	ID               int      `json:"id"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	FirstAirDate     string   `json:"first_air_date"`
	Name             string   `json:"name"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

// Genre represents a TV show genre
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Network represents a TV network
type Network struct {
	ID            int     `json:"id"`
	LogoPath      *string `json:"logo_path"`
	Name          string  `json:"name"`
	OriginCountry string  `json:"origin_country"`
}

// LastEpisodeToAir represents the last episode that aired
type LastEpisodeToAir struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	EpisodeType    string  `json:"episode_type"`
	ProductionCode string  `json:"production_code"`
	Runtime        int     `json:"runtime"`
	SeasonNumber   int     `json:"season_number"`
	ShowID         int     `json:"show_id"`
	StillPath      string  `json:"still_path"`
}

// ProductionCompany represents a production company
type ProductionCompany struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
	LogoPath      string `json:"logo_path"`
}

// SeasonTmdb represents a TV show season
type SeasonTmdb struct {
	AirDate      string `json:"air_date"`
	EpisodeCount int     `json:"episode_count"`
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	SeasonNumber int     `json:"season_number"`
	VoteAverage  float64 `json:"vote_average"`
}

// TVShow represents detailed TV show information from TMDB
type TVShow struct {
	Adult               bool                         `json:"adult"`
	BackdropPath        string                       `json:"backdrop_path"`
	CreatedBy           []any                `json:"created_by"`
	EpisodeRunTime      []int                        `json:"episode_run_time"`
	FirstAirDate        string                       `json:"first_air_date"`
	Genres              []Genre                      `json:"genres"`
	Homepage            string                       `json:"homepage"`
	ID                  int                          `json:"id"`
	InProduction        bool                         `json:"in_production"`
	Languages           []string                     `json:"languages"`
	LastAirDate         string                       `json:"last_air_date"`
	LastEpisodeToAir    any                  `json:"last_episode_to_air"` // Can be string or object
	Name                string                       `json:"name"`
	Networks            []Network                    `json:"networks"`
	NumberOfEpisodes    int                          `json:"number_of_episodes"`
	NumberOfSeasons     int                          `json:"number_of_seasons"`
	OriginCountry       []string                     `json:"origin_country"`
	OriginalLanguage    string                       `json:"original_language"`
	OriginalName        string                       `json:"original_name"`
	Overview            string                       `json:"overview"`
	Popularity          float64                      `json:"popularity"`
	PosterPath          string                       `json:"poster_path"`
	ProductionCompanies []ProductionCompany          `json:"production_companies"`
	ProductionCountries []map[string]string          `json:"production_countries"`
	Seasons             []SeasonTmdb                     `json:"seasons"`
	NextEpisodeToAir    any                  `json:"next_episode_to_air"` // Can be string or null
}

// TMDBInfo represents processed TMDB information for a bangumi
type TMDBInfo struct {
	ID            int              `json:"id"`
	Title         string           `json:"title"`
	OriginalTitle string           `json:"original_title"`
	Seasons       []SeasonTmdb   `json:"seasons"` // Changed from int to []SeasonSimple
	LastSeason    int              `json:"last_season"`
	Year          string           `json:"year"`
	PosterLink    string           `json:"poster_link"`
}


// SearchResult represents the search result from TMDB API
type SearchResult struct {
	Page         int        `json:"page"`
	Results      []ShowInfo `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}
