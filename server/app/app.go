package app

import (
	"guilhermec-costa/go-web-crawler/server/infra"
	"guilhermec-costa/go-web-crawler/server/services"
)

type DIContainer struct {
	UserStore              infra.UserDAO
	JobMonitor             *infra.JobMonitor
	CrawlerExtractionStore infra.CrawlerExtractionDAO
	UserService            *services.AuthService
	CrawlerService         *services.CrawlerService
}