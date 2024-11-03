package repository

import "time"

func parseTime(t *string) (*time.Time, error) {
	if t == nil || *t == "" {
		return nil, nil
	}
	ti, err := time.Parse("2006-01-02 15:04:05-07:00", *t)
	if err != nil {
		ti, err = time.Parse("2006-01-02T15:04:05.999999-07:00", *t)
	}
	return &ti, err
}
