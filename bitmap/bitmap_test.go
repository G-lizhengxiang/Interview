package bitmap

import (
	"testing"
	"fmt"
)

func TestSigned(t *testing.T) {
	err := Signed("1099")
	if nil != err {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
}

func TestGetSignedDayNumber(t *testing.T) {
	num, err := GetSignedDayNumber("1099")
	if nil != err {
		fmt.Println(err)
	}else {
		fmt.Println(num,err)
	}
}

func TestTimeDiff(t *testing.T) {
	fmt.Println(TimeDiff())
}