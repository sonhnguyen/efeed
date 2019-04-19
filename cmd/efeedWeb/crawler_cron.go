package main

import (
	"efeed"
	"net/http"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	efeedConfig := efeed.Config{DoSpaceURL: a.config.DoSpaceURL, EnableProxy: a.config.EnableProxy, ProxyURL: a.config.ProxyURL}

	err := efeed.RunCrawlerFanatics(efeedConfig, a.svc)
	if err != nil {
		return err
	}

	return nil
}

// RunCrawlerSoccerProAndSave RunCrawlerSoccerProAndSave
func (a *App) RunCrawlerSoccerProAndSave() error {
	efeedConfig := efeed.Config{DoSpaceURL: a.config.DoSpaceURL, EnableProxy: a.config.EnableProxy, ProxyURL: a.config.ProxyURL}

	err := efeed.RunCrawlerSoccerPro(efeedConfig, a.svc)
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
