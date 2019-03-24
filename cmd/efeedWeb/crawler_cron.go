package main

import (
	"efeed"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	err := efeed.RunCrawlerFanatics()
	if err != nil {
		return err
	}

	return nil
}
