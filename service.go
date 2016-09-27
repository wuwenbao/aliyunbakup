package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wuwenbao/aliyunbakup/uitl"
)

//Config 配置信息
type Config struct {
	localDir string
}

const template = `阿里云虚拟主机备份（请输入以下数字）：
1、MYSQL + FTP 全备份
2、Mysql 单备份
3、FTP 单备份
4、结束退出`

func main() {
Exit:
	for {
		fmt.Print(template, "\n:")
		var num int
		fmt.Scanf("%d\n", &num)
		switch num {
		case 1:
			zdyMysqlFtp()
			zdyZip()
		case 2:
			zdyMysql()
			zdyZip()
		case 3:
			zdyFtp()
			zdyZip()
		case 4:
			break Exit
		}
	}
	return
}

func zdyMysqlFtp() {
	var localDir string
	var err error
	if localDir, err = os.Getwd(); err != nil {
		log.Panicf("根目录获取失败：%s\n", err)
	}
	//Mysql参数
	mysql := &uitl.MysqlConf{}
	var mysqlHost, mysqlUser, mysqlPass, mysqlDb string
	fmt.Print("请输入Mysql服务器连接:")
	fmt.Scanf("%s\n", &mysqlHost)
	mysql.Host = mysqlHost + ":3306"
	fmt.Print("请输入数据库用户名:")
	fmt.Scanf("%s\n", &mysqlUser)
	mysql.User = mysqlUser
	fmt.Print("请输入数据库密码:")
	fmt.Scanf("%s\n", &mysqlPass)
	mysql.Pass = mysqlPass
	fmt.Print("请输入选择数据库:")
	fmt.Scanf("%s\n", &mysqlDb)
	mysql.Db = mysqlDb
	mysql.File = "database.sql"
	mysql.LocalDir = localDir + "/bakup"
	//FTP参数
	ftp := &uitl.FtpConf{}
	var ftpHost, ftpUser, ftpPass, ftpServerDir string
	fmt.Print("请输入FTP服务器连接:")
	fmt.Scanf("%s\n", &ftpHost)
	ftp.Host = ftpHost + ":21"
	fmt.Print("请输入FTP用户名:")
	fmt.Scanf("%s\n", &ftpUser)
	ftp.User = ftpUser
	fmt.Print("请输入FTP密码:")
	fmt.Scanf("%s\n", &ftpPass)
	ftp.Pass = ftpPass
	fmt.Print("请输入FTP服务器目录（默认：/htdocs）:")
	fmt.Scanf("%s\n", &ftpServerDir)
	if ftpServerDir == "" {
		ftp.ServerDir = "/htdocs"
	} else {
		ftp.ServerDir = ftpServerDir
	}
	ftp.LocalDir = localDir + "/bakup"
	// 参数获取完成
	log.Println("Mysql备份中...")
	err = uitl.Mysql(mysql)
	if err != nil {
		log.Printf("MYSQL备份失败：%s\n", err)
	}
	log.Println("Mysql备份成功")
	log.Println("FTP下载中...")
	err = uitl.Ftp(ftp)
	if err != nil {
		fmt.Printf("FTP下载失败:%s\n", err)
	}
	log.Println("FTP下载成功")
}
func zdyMysql() {
	var localDir string
	var err error
	if localDir, err = os.Getwd(); err != nil {
		log.Panicf("根目录获取失败：%s\n", err)
	}
	//Mysql数据库备份
	mysql := &uitl.MysqlConf{}
	var host, user, pass, db string
	fmt.Print("请输入Mysql服务器连接:")
	fmt.Scanf("%s\n", &host)
	mysql.Host = host
	fmt.Print("请输入数据库用户名:")
	fmt.Scanf("%s\n", &user)
	mysql.User = user
	fmt.Print("请输入数据库密码:")
	fmt.Scanf("%s\n", &pass)
	mysql.Pass = pass
	fmt.Print("请输入选择数据库:")
	fmt.Scanf("%s\n", &db)
	mysql.Db = db
	// mysql.Host = "qdm166535389.my3w.com:3306"
	// mysql.User = "qdm166535389"
	// mysql.Pass = "1qaz2wsx"
	// mysql.Db = "qdm166535389_db"
	mysql.File = "database.sql"
	mysql.LocalDir = localDir + "/bakup"
	log.Println("Mysql备份中...")
	err = uitl.Mysql(mysql)
	if err != nil {
		log.Printf("MYSQL备份失败：%s\n", err)
	}
	log.Println("Mysql备份成功")
}

func zdyFtp() {
	var localDir string
	var err error
	if localDir, err = os.Getwd(); err != nil {
		log.Printf("根目录获取失败：%s\n", err)
	}
	//FTP源码备份
	ftp := &uitl.FtpConf{}
	var host, user, pass, serverDir string
	fmt.Print("请输入FTP服务器连接:")
	fmt.Scanf("%s\n", &host)
	ftp.Host = host
	fmt.Print("请输入FTP用户名:")
	fmt.Scanf("%s\n", &user)
	ftp.User = user
	fmt.Print("请输入FTP密码:")
	fmt.Scanf("%s\n", &pass)
	ftp.Pass = pass
	fmt.Print("请输入FTP服务器目录（默认：/htdocs）:")
	fmt.Scanf("%s\n", &serverDir)
	if serverDir == "" {
		ftp.ServerDir = "/htdocs"
	} else {
		ftp.ServerDir = serverDir
	}
	// ftp.Host = "121.42.113.195:21"
	// ftp.User = "qxu1588640052"
	// ftp.Pass = "1qaz2wsx"
	ftp.LocalDir = localDir + "/bakup"
	log.Println("FTP下载中...")
	err = uitl.Ftp(ftp)
	if err != nil {
		fmt.Printf("FTP下载失败:%s\n", err)
	}
	log.Println("FTP下载成功")
}

func zdyZip() {
	var localDir string
	var err error
	if localDir, err = os.Getwd(); err != nil {
		log.Panicf("根目录获取失败：%s\n", err)
	}
	//Zip打包
	zipConfing := &uitl.ZipConf{}
	zipConfing.LocalDir = localDir + "/bakup"
	zipConfing.File = "web.zip"
	log.Println("Zip打包中...")
	err = uitl.Zip(zipConfing)
	if err != nil {
		fmt.Println(err)
		log.Println("Zip打包失败")
	}
	log.Println("Zip打包成功")
}
