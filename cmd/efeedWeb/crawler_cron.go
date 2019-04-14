package main

import (
	"efeed"
	"net/http"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	err := efeed.RunCrawlerFanatics(a.config.DoSpaceURL, a.svc)
	if err != nil {
		return err
	}

	return nil
}

// RunCrawlerSoccerProAndSave RunCrawlerSoccerProAndSave
func (a *App) RunCrawlerSoccerProAndSave() error {
	err := efeed.RunCrawlerSoccerPro(a.config.DoSpaceURL, a.svc)
	if err != nil {
		return err
	}

	return nil
}

// RunPingHeroku RunPingHeroku
func (a *App) RunPingHeroku() error {
	_, err := http.Get("http://efeed.herokuapp.com")
	if err != nil {
		return err
	}
	return nil
}
