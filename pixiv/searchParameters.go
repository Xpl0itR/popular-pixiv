package pixiv

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type SearchParameters struct {
	Word      string
	Offset    int
	Match     string
	Sort      string
	Duration  string
	StartDate string
	EndDate   string
	Filter    string
}

func (params *SearchParameters) toURLEncodedParams() string {
	return url.Values{
		"word":          {params.Word},
		"offset":        {strconv.Itoa(params.Offset)},
		"search_target": {params.Match},
		"sort":          {params.Sort},
		"duration":      {params.Duration},
		"start_date":    {params.StartDate},
		"end_date":      {params.EndDate},
		"filter":        {params.Filter},
	}.Encode()
}

func (params *SearchParameters) Validate() error {
	if params.Word == "" {
		return parameterError("word", params.Word)
	}

	if params.Match != "" && params.Match != "exact_match_for_tags" && params.Match != "partial_match_for_tags" && params.Match != "title_and_caption" {
		return parameterError("search_target", params.Match)
	}

	if params.Sort != "" && params.Sort != "date_asc" && params.Sort != "date_desc" && params.Sort != "popular_desc" {
		return parameterError("sort", params.Sort)
	}

	if params.Duration != "" && params.Duration != "within_last_month" && params.Duration != "within_last_week" && params.Duration != "within_last_day" {
		return parameterError("duration", params.Duration)
	}

	if params.StartDate != "" {
		if _, err := time.Parse("2000-01-01", params.StartDate); err != nil {
			return parameterError("start_date", params.StartDate)
		}
	}

	if params.EndDate != "" {
		if _, err := time.Parse("2000-01-01", params.EndDate); err != nil {
			return parameterError("end_date", params.EndDate)
		}
	}

	return nil
}

func parameterError(parameterName string, parameterValue string) error {
	return fmt.Errorf("\"%s\" is not a valid value for the \"%s\" parameter", parameterValue, parameterName)
}
