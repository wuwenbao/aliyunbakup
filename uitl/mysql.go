package uitl

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//MysqlConf ftp配置参数
type MysqlConf struct {
	Host     string
	User     string
	Pass     string
	Db       string
	LocalDir string
	File     string
}

//DataBase 数据库备份
type dataBase struct {
	db   *sql.DB
	conf *MysqlConf
	file *os.File
}

//Mysql 初始化
func Mysql(conf *MysqlConf) error {
	f := new(dataBase)
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?allowOldPasswords=1", conf.User, conf.Pass, conf.Host, conf.Db)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Open database error: %s\n", err)
		return err
	}
	f.db = db
	f.conf = conf
	defer f.db.Close()
	err = f.db.Ping()
	if err != nil {
		log.Printf("Ping database error: %s\n", err)
		return err
	}
	_, err = os.Stat(conf.LocalDir)
	if err != nil {
		err = os.MkdirAll(conf.LocalDir, os.ModePerm) //生成多级目录
		if err != nil {
			return err
		}
	}
	filename := conf.LocalDir + "/" + conf.File
	os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("File error: %s\n", err)
		return err
	}
	defer file.Close()
	f.file = file
	err = f.listTables()
	if err != nil {
		return err
	}
	return nil
}

func (f *dataBase) listTables() error {
	rows, err := f.db.Query("show tables")
	if err != nil {
		return err
	}
	defer rows.Close()
	f.file.WriteString("/*\n")
	f.file.WriteString("金方时代（技术部） Database Transfer\n")
	f.file.WriteString("Source File           : " + f.conf.File + "\n")
	f.file.WriteString("Source Host           : " + f.conf.Host + "\n")
	f.file.WriteString("Source Database       : " + f.conf.Db + "\n")
	f.file.WriteString("Date: " + time.Now().Format("2006-01-02 15:04:05"))
	f.file.WriteString("\n*/\n\n")
	f.file.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
	var table string
	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			log.Fatal(err)
		}
		err = f.xcreate(table)
		if err != nil {
			log.Fatal(err)
		}
		err = f.xselect(table)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (f *dataBase) xcreate(table string) error {
	rows, err := f.db.Query("show create table `" + table + "`")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer rows.Close()

	var tab, createTable string
	for r := 0; rows.Next(); r++ {
		err = rows.Scan(&tab, &createTable)
		if err != nil {
			log.Fatal(err)
			return err
		}
		f.file.WriteString("\n-- ----------------------------\n")
		f.file.WriteString("-- Table structure for `" + tab + "`\n")
		f.file.WriteString("-- ----------------------------\n")
		f.file.WriteString("DROP TABLE IF EXISTS `" + tab + "`;\n")
		f.file.WriteString(createTable + ";\n")
	}
	return nil
}

func (f *dataBase) xselect(table string) error {
	rows, err := f.db.Query("select * from `" + table + "`")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	f.file.WriteString("\n-- ----------------------------\n")
	f.file.WriteString("-- Records of " + table + "\n")
	f.file.WriteString("-- ----------------------------\n")
	for r := 0; rows.Next(); r++ {
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
			return err
		}
		strarr := make([]string, len(columns))
		for i, col := range values {
			if col != nil {
				strarr[i] = strings.Replace(string(col), "'", "''", -1)
			} else {
				strarr[i] = ""
			}
		}
		f.file.WriteString("INSERT INTO `" + table + "` VALUES ('" + strings.Join(strarr, "', '") + "');\n")
	}
	return nil
}
