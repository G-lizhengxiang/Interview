package bitmap

import (
	"github.com/garyburd/redigo/redis"
	"errors"
	"time"
	"fmt"
)

const STARTDATE  = "2019-02-21 00:00:00"

//用户签到
func Signed(userId string) error {
	conn,err := redis.Dial("tcp","127.0.0.1:6379")
	if err != nil {
		return errors.New("connect redis error:"+fmt.Sprintln(err))
	}
	defer conn.Close()
	_, err = conn.Do("SETBIT", "userSigned:"+userId, TimeDiff(),1)
	if err != nil {
		return errors.New("redis set error: "+fmt.Sprintln(err))
	}

	return nil
}


//获取用户签到天数
func GetSignedDayNumber(userId string) (int, error) {
	conn,err := redis.Dial("tcp","127.0.0.1:6379")

	if err != nil {
		return 0,errors.New("connect redis error: "+fmt.Sprintln(err))
	}
	defer conn.Close()
	num := 0
	for i := TimeDiff(); i >= 0; i-- {
		vale, err := conn.Do("GETBIT", "userSigned:"+userId,i)
		if err != nil {
			return 0,errors.New("redis get error: "+fmt.Sprintln(err))
		}
		if 1 == vale.(int64) {
			num++
		} else {
			return num,nil
		}
	}
	return num,nil
}


//计算签到时间差
func TimeDiff() int64 {

	//开始签到时间
	formatStartDate,_:=time.Parse("2006-01-02 15:04:05",STARTDATE)
	startDateUnix := formatStartDate.Unix()

	//获取签到时间
	currentDateUnix := time.Now().Unix()
	return (currentDateUnix-startDateUnix)/86400
}