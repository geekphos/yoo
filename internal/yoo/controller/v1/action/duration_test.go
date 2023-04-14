package action

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	duration := time.Duration(412) * time.Second // example duration of 2 days

	days := int(duration.Hours()) / 24      // calculate the number of days from hours
	hours := int(duration.Hours()) % 24     // calculate the remaining hours
	minutes := int(duration.Minutes()) % 60 // calculate the remaining minutes
	seconds := int(duration.Seconds()) % 60 // calculate the remaining seconds

	dataString := fmt.Sprintf("%02d:%02d:%02d:%02d", days, hours, minutes, seconds) // f7325ormat the output string

	var output []string

	zeroRex, err := regexp.Compile(`^0+$`)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(dataString)

	for _, section := range strings.Split(dataString, ":") {
		if !zeroRex.MatchString(section) {
			output = append(output, section)
		} else if len(output) > 0 {
			output = append(output, section)
		}
	}

	units := []string{"秒", "分钟", "小时", "天"}

	fmt.Println(units[1:2])

	// reverse the output
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}

	length := len(output)

	for i := 0; i < length; i++ {
		output = append(output[:(0+i*2)], append([]string{units[i]}, output[(i*2):]...)...)
	}

	// reverse the output array
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}

	fmt.Println(strings.Join(output, ""))
}
