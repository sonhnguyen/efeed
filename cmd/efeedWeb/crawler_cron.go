package main

import (
	"efeed"
	"net/http"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	err := efeed.RunCrawlerFanatics()
	if err != nil {
		return err
	}

	return nil
}

// RunCrawlerSoccerProAndSave RunCrawlerSoccerProAndSave
func (a *App) RunCrawlerSoccerProAndSave() error {
	err := efeed.RunCrawlerSoccerPro()
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
