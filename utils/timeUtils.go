package utils

import (
	"github.com/jinzhu/now"
	"log"
	"time"
)

func getNowConfig() *now.Config {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatal("cannot LoadLocation UTC: ", err)
	}

	return &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
	}
}

func UTCNow() time.Time {
	return time.Now().UTC()
}

func UTCNowMilli() int64 {
	return time.Now().UTC().UnixMilli()
}

func UTCNowBeginningOfWeek() int64 {
	myConfig := getNowConfig()
	return myConfig.With(time.Now().UTC()).BeginningOfWeek().UnixMilli()
}

func UTCNowBeginningOfMonth() int64 {
	myConfig := getNowConfig()
	return myConfig.With(time.Now().UTC()).BeginningOfMonth().UnixMilli()
}
