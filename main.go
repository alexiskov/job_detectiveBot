package main

import (
	"fmt"
	"os"

	"vacancydealer/hh"
	"vacancydealer/logger"
)

func main() {
	logger.InitInfoTextlog(os.Stdout)
	logger.Info("logger status is Run...")

	logger.InitErrorJSONlog(os.Stdout)
	logger.Info("debug log stream status is run!")

	res, err := hh.SentRequest("golang", hh.REMOTE_JOB, hh.NO_EXPERIENCE)
	if err != nil {
		logger.Error(err.Error())
	}

	fmt.Printf("\n%+v\n", res)
}
