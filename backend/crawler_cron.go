package main

import (
	"github.com/sonhnguyenn/cmd/crawler"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	err := crawler.RunCrawlerFanatics()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) RunCrawlerSoccerProAndSave() error {
	err := crawler.RunCrawlerSoccerPro()
	if err != nil {
		return err
	}

	return nil
}
