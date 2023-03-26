package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestS(t *testing.T) {
	//now := time.Date(2023, 3, 26, 1, 26, 15, 0, spbLocation)
	//newDate := now.AddDate(0, 0, 1)
	//fmt.Println(newDate.Format(time.RFC3339Nano))
	//fmt.Println(newDate.In(time.UTC).Format(time.RFC3339Nano))

	//

	now := time.Date(2023, 3, 25, 22, 26, 15, 0, time.UTC)
	fmt.Println(now.In(spbLocation).Format(time.RFC3339Nano))
	year, month, day := now.In(spbLocation).AddDate(0, 0, 1).Date()
	nextTry := time.Date(year, month, day, 1, 0, 0, 0, spbLocation)
	fmt.Println(nextTry.Format(time.RFC3339Nano))

}
