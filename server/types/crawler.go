package types

import "guilhermec-costa/go-web-crawler/crawler/validation"

type Job struct {
	Params validation.CrawlerParams
	UserId string
}

type JobProcessor func(Job) error
