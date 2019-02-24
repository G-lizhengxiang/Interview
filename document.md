第一题


第二题

```go
package transition

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/joho/godotenv/autoload"
	"fmt"
	"errors"
)
//连接数据库相关操作
var (
	connection *gorm.DB
)

func init() {
	connection = connect()
	//自动创建表
	connection.AutoMigrate(&UserBalance{},&UserTransitionDetail{},&UserInfo{})
}

func GetDB() *gorm.DB {
	return connection
}

const DATABASEURL  = "lzhx:xxxx@/transition?charset=utf8&parseTime=True&loc=Local"
func connect() *gorm.DB {
	conn, err := gorm.Open("mysql", DATABASEURL)
	if err != nil {
		panic(err)
	}
	conn.SingularTable(true)
	conn.LogMode(true)
	return conn
}




/**
mysql>
mysql> SHOW CREATE TABLE user_info \G;
*************************** 1. row ***************************
       Table: user_info
Create Table: CREATE TABLE `user_info` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_info_deleted_at` (`deleted_at`),
  KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1
1 row in set (0.00 sec)

mysql>
mysql> SHOW CREATE TABLE user_balance \G;
*************************** 1. row ***************************
       Table: user_balance
Create Table: CREATE TABLE `user_balance` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL,
  `balance` double DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_balance_deleted_at` (`deleted_at`),
  KEY `user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1
1 row in set (0.00 sec)


mysql> SHOW CREATE TABLE user_transition_detail \G;
*************************** 1. row ***************************
       Table: user_transition_detail
Create Table: CREATE TABLE `user_transition_detail` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `from_user_id` int(11) DEFAULT NULL,
  `to_user_id` int(11) DEFAULT NULL,
  `value` double DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_transition_detail_deleted_at` (`deleted_at`),
  KEY `from_user` (`from_user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=latin1
1 row in set (0.00 sec)


 */


//用户信息表
type UserInfo struct {
	gorm.Model
	//添加索引，高频查询
	Name string		`gorm:"index:name"`
}


//用户余额表
type UserBalance struct {
	gorm.Model
	//添加索引，高频查询
	UserId int		`gorm:"index:user"`
	Balance float64
}

//用户转账明细表
type UserTransitionDetail struct {
	gorm.Model
	//添加索引，高频查询
	FromUserId int	 `gorm:"index:from_user"`
	ToUserId int
	Value float64
}



//根据用户获取余额
func GetUserInfoByUserId(userId int) (uint, error) {
	var userInfo UserInfo
	db := GetDB()
	err := db.Where("id = ?",userId).Last(&userInfo).Error
	if err != nil {
		return 0,errors.New("get balance error:"+fmt.Sprintln(err))
	}
	return userInfo.ID,nil
}



//初始化用户余额
func InsterUser(user *UserInfo) error {
	db := GetDB()
	err := db.Create(user).Error
	if err != nil {
		return errors.New("inster user: "+fmt.Sprintln(err))
	}
	return nil
}

//根据用户获取余额
func GetBalanceByUserId(userId int) (float64, error) {
	var userBalance UserBalance
	db := GetDB()
	err := db.Where("user_id = ?",userId).Last(&userBalance).Error
	if err != nil {
		return 0,errors.New("get balance error:"+fmt.Sprintln(err))
	}
	return userBalance.Balance,nil
}

//初始化用户余额
func InsterUserBalance(userBalance *UserBalance) error {
	userId,err := GetUserInfoByUserId(userBalance.UserId)
	if err != nil || 0 == userId {
		return errors.New("get user info error")
	}
	db := GetDB()
	err = db.Create(userBalance).Error
	if err != nil {
		return errors.New("inster user balance error: "+fmt.Sprintln(err))
	}
	return nil
}

//转账
func Transition(fromUserId,toUserId int,value float64 ) error {
	//先判断账户是否存在等信息
	userId,err := GetUserInfoByUserId(fromUserId)
	if err != nil || 0 == userId {
		return errors.New("get fromUser info error")
	}
	userId,err = GetUserInfoByUserId(toUserId)
	if err != nil || 0 == userId {
		return errors.New("get toUser info error")
	}
	if value < 0 {
		return errors.New("The transfer amount is wrong")
	}
	//根据账户地址获取余额
	fromUserBalance,err := GetBalanceByUserId(fromUserId)
	if err != nil {
		return errors.New("get from user balance error: "+fmt.Sprintln(err))
	}
	toUserBalance,err := GetBalanceByUserId(toUserId)
	if err != nil {
		return errors.New("get to user balance error: "+fmt.Sprintln(err))
	}

	//判断余额是否够
	balanceDiff :=  fromUserBalance - value
	if balanceDiff < 0 {
		return errors.New("Insufficient balance")
	}

	//转账操作
	tx := GetDB()
	//开启实物
	db := tx.Begin()
	//更新余额表
	err = db.Model(&UserBalance{}).Where("user_id = ?", fromUserId).Updates(
		map[string]interface{}{"balance": balanceDiff}).Error
	if err != nil {
		db.Rollback()
		return errors.New("update user from balance error: "+fmt.Sprintln(err))
	}

	err = db.Model(&UserBalance{}).Where("user_id = ?", toUserId).Updates(
		map[string]interface{}{"balance": toUserBalance+value}).Error
	if err != nil {
		db.Rollback()
		return errors.New("update user to balance error: "+fmt.Sprintln(err))
	}

	//更新明细表
	err = db.Create(&UserTransitionDetail{
		FromUserId:fromUserId,
		ToUserId:toUserId,
		Value:value,
	}).Error
	if err != nil {
		db.Rollback()
		return errors.New("inster User Transition Detail error: "+fmt.Sprintln(err))
	}

	db.Commit()
	return nil
}

```


第三题 
1 先ping下服务器地址

2 如果公司有服务监控，则先看公司服务器监控 cpu、内存、网卡、磁盘、url请求的响应时间请求等

3 查看nginx请求日志，如果nginx响应很快说明不是，应用系统服务器本身，说明是网络环境有问题，联系运维解决网络

4 如果nginx 响应比较慢，看服务器负载，mysql负载，redis负载

1> 务器负载比较高，看下是否有异常流量，如果发现是恶意流量则限频。如果mysql负载，redis负载度比较底则
查看业务代码是不是在哪里阻塞住了。

2>如果mysql负载高，查询慢，先看看是不是可以优化下索引等操作

5 动态扩容



第四题
```go
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
```

第五题

用户中心项目：分析需求并参与功能规划的讨论，参与实现功能,数据库优化，索引
优化,对 text 等大字段进行业务拆分.该项目每天处理上亿请求


上线流程：

1 测试环境测试通过

2 demo 环境测试通过

3 灰度环境（所有配置和生成环境一致），现从nginx转发机拦截接口流量到灰度环境，观察nginx日志，mysql负载等情况。
如果灰度环境跑一天没有问题，则找个合适的时间点（系统负载比较底，联系各部门同事比较方便的时候，用户中心属于公共服务）防止服务抖动

4 上线生成环境，观察nginx日志，mysql负载等情况

5 取消灰度环境流量