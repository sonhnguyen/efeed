package main

import (
	"efeed"
	"fmt"
)

// RunCrawlerFanaticsAndSave RunCrawlerFanaticsAndSave
func (a *App) RunCrawlerFanaticsAndSave() error {
	values, err := efeed.RunCrawlerFanatics()
	if err != nil {
		return err
	}

	fmt.Println(values)
	return nil
}
