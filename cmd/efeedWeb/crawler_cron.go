package main

import (
	"efeed"
	"fmt"
)

func (a *App) RunCrawlerFanaticsAndSave() error {
	values, err := efeed.RunCrawlerFanatics()
	if err != nil {
		return err
	}
	var animal = efeed.User{Name: "hello"}
	a.db.Create(&animal)

	fmt.Println(values)
	return nil
}
