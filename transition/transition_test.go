package transition

import (
	"testing"
	"fmt"
)

func TestInsterUser(t *testing.T) {
	err := InsterUser(&UserInfo{
		Name:"userB",
	})

	if nil != err {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
}


func TestInsterUserBalance(t *testing.T) {
	err := InsterUserBalance(&UserBalance{
		UserId:3,
		Balance:100,
	})

	if nil != err {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
}


func TestGetBalanceByUserId(t *testing.T) {
	balance,err := GetBalanceByUserId(2)
	if nil !=  err {
		fmt.Println(err)
	}else{
		fmt.Println(balance)
	}
}

func TestTransition(t *testing.T) {
	err := Transition(1,2,13)
	if nil != err {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
}
