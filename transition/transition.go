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

const DATABASEURL  = "lzhx:lzx@/genaroNetworkMonitor?charset=utf8&parseTime=True&loc=Local"
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


//用户余额表
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