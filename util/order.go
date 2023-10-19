package util

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	orderIDCounter int
	mutex          sync.Mutex
)

func NewOrderID() int64 {
	mutex.Lock()
	defer mutex.Unlock()

	currentTime := time.Now().UTC()
	year, month, day := currentTime.Date()
	hour := currentTime.Hour()

	orderIDCounter++
	uniqueID := orderIDCounter

	// Generate the order ID in the format: YYMMDDHHXXXX
	orderIDStr := fmt.Sprintf("%02d%02d%02d%02d%04d", year%100, month, day, hour, uniqueID)

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		// Handle error if conversion fails
		fmt.Println("Error converting orderID to int64:", err)
		return 0
	}

	return orderID
}
