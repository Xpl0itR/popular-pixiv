package pixiv

import (
    "fmt"
    "net/url"
    "strconv"
)

const (
    wordParam      = "word"
    offsetParam    = "offset"
    matchParam     = "search_target"
    sortParam      = "sort"
    durationParam  = "duration"
    startDateParam = "start_date"
    endDateParam   = "end_date"
    filterParam    = "filter"
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

func (params *SearchParameters) toString() (string, error) {
    if params.Word == "" {
        return "", parameterError(wordParam, params.Word)
    }

    values := url.Values {
        wordParam:   { params.Word },
        offsetParam: { strconv.Itoa(params.Offset) },
    }

    if params.Match != "" {
        if err := params.validateMatchOption(); err != nil {
            return "", err
        }

        values.Set(matchParam, params.Match)
    }

    if params.Sort != "" {
        if err := params.validateSortOption(); err != nil {
            return "", err
        }

        values.Set(sortParam, params.Sort)
    }

    if params.Duration != "" {
        if err := params.validateDurationOption(); err != nil {
            return "", err
        }

        values.Set(durationParam, params.Duration)
    }

    if params.StartDate != "" {
        if err := params.validateStartDateOption(); err != nil {
            return "", err
        }

        values.Set(startDateParam, params.StartDate)
    }

    if params.EndDate != "" {
        if err := params.validateEndDateOption(); err != nil {
            return "", err
        }

        values.Set(endDateParam, params.EndDate)
    }

    if params.Filter != "" {
        values.Set(filterParam, params.Filter)
    }

    return values.Encode(), nil
}

func (params *SearchParameters) validateMatchOption() error {
    if params.Match == "exact_match_for_tags" || params.Match == "partial_match_for_tags" || params.Match == "title_and_caption" {
        return nil
    }

    return parameterError(matchParam, params.Match)
}

func (params *SearchParameters) validateSortOption() error {
    if params.Sort == "date_asc" || params.Sort == "date_desc" || params.Sort == "popular_desc" {
        return nil
    }

    return parameterError(sortParam, params.Sort)
}

func (params *SearchParameters) validateDurationOption() error {
    if params.Duration == "within_last_month" || params.Duration == "within_last_week" || params.Duration == "within_last_day" {
        return nil
    }

    return parameterError(durationParam, params.Duration)
}

func (params *SearchParameters) validateStartDateOption() error {
    if dateOptionIsValid(params.StartDate) {
        return nil
    }

    return parameterError(startDateParam, params.StartDate)
}

func (params *SearchParameters) validateEndDateOption() error {
    if dateOptionIsValid(params.EndDate) {
        return nil
    }

    return parameterError(endDateParam, params.EndDate)
}

func parameterError(parameterName string, parameterValue string) error {
    return fmt.Errorf("\"%s\" is not a valid value for the \"%s\" parameter", parameterValue, parameterName)
}

func dateOptionIsValid(str string) bool {
    return true // TODO: implement this
}