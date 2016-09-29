package main

import (
	"container/list"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/wuwenbao/aliyunbakup/uitl"
)

//Config 配置信息
type Config struct {
	localDir string
}

// bakTask 备份参数
type bakTask struct {
	name      string
	mysqlHost string
	mysqlUser string
	mysqlPass string
	mysqlPort string
	mysqlDb   string
	ftpHost   string
	ftpUser   string
	ftpPass   string
	ftpPort   string
	LocalDir  string
	ServerDir string
	ctime     time.Time
	status    int
}

//Quest 队列
type quest struct {
	items        *list.List //队列
	itemsSuccess *list.List //成功列表
	itemsError   *list.List //失败列表
}

func main() {
	q := new(quest)
	q.items = list.New()
	q.itemsSuccess = list.New()
	q.itemsError = list.New()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		q.httpx()
		log.Println("http server end")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		q.bakup()
		log.Println("bakup end")
	}()

	wg.Wait()
}

//UserInfo xx
type UserInfo struct {
	Name string
}

//D x
type D struct {
	Title        string
	Items        int
	ItemsSuccess int
	ItemsError   int
}

func (q *quest) httpx() {
	// 设置静态目录
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 规则1
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		switch r.Method {
		case "GET":
			t, err := template.ParseFiles("view/index.html")
			if err != nil {
				log.Println(err, "11")
			}
			err = t.Execute(w, "任务提交页面")
			if err != nil {
				log.Println(err, "22")
			}
			// w.Write([]byte("OK"))
		case "POST":
			name := r.FormValue("name")
			mysqlHost := r.FormValue("mysqlHost")
			mysqlUser := r.FormValue("mysqlUser")
			mysqlPass := r.FormValue("mysqlPass")
			mysqlPort := r.FormValue("mysqlPort")
			mysqlDb := r.FormValue("mysqlDb")

			ftpHost := r.FormValue("ftpHost")
			ftpUser := r.FormValue("ftpUser")
			ftpPass := r.FormValue("ftpPass")
			ftpPort := r.FormValue("ftpPort")

			LocalDir := r.FormValue("LocalDir")
			ServerDir := r.FormValue("ServerDir")

			if name == "" {
				w.Write([]byte("<script type=\"text/javascript\">alert(\"备份名称不能为空\"); history.go(-1);</script>"))
				return
			}

			//添加任务队列
			task := bakTask{
				name:      name,
				mysqlHost: mysqlHost,
				mysqlUser: mysqlUser,
				mysqlPass: mysqlPass,
				mysqlPort: mysqlPort,
				mysqlDb:   mysqlDb,
				ftpHost:   ftpHost,
				ftpUser:   ftpUser,
				ftpPass:   ftpPass,
				ftpPort:   ftpPort,
				LocalDir:  LocalDir,
				ServerDir: ServerDir,
				ctime:     time.Now(),
				status:    1,
			}
			q.items.PushBack(task)
			w.Write([]byte("<script type=\"text/javascript\">alert(\"任务提交成功！\"); history.go(-1);</script>"))
		}
	})

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte(fmt.Sprintf("队列：%d\n成功：%d\n失败：%d\n", q.items.Len(), q.itemsSuccess.Len(), q.itemsError.Len())))
		t, err := template.ParseFiles("view/items.html")
		if err != nil {
			log.Println(err)
		}
		data := D{
			Title:        "队列页面",
			Items:        q.items.Len(),
			ItemsSuccess: q.itemsSuccess.Len(),
			ItemsError:   q.itemsError.Len()}
		err = t.Execute(w, data)
		if err != nil {
			log.Println(err, "22")
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (q *quest) bakup() {
	for {
		for e := q.items.Front(); e != nil; e = e.Next() {
			q.items.Remove(e)
			task := e.Value.(bakTask)
			err := zdyMysqlFtp(task)
			if err != nil {
				log.Println(err)
				q.itemsError.PushBack(task)
			} else {
				q.itemsSuccess.PushBack(task)
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func zdyMysqlFtp(task bakTask) error {
	var localDir string
	var err error
	if localDir, err = os.Getwd(); err != nil {
		log.Panicf("根目录获取失败：%s\n", err)
		return err
	}
	localDir += "/bakup"
	//Mysql参数
	mysql := &uitl.MysqlConf{}

	mysql.Host = task.mysqlHost + ":3306"
	mysql.User = task.mysqlUser
	mysql.Pass = task.mysqlPass
	mysql.Db = task.mysqlDb
	mysql.File = "database.sql"
	mysql.LocalDir = localDir + "/" + task.name
	//FTP参数
	ftp := &uitl.FtpConf{}
	ftp.Host = task.ftpHost + ":21"
	ftp.User = task.ftpUser
	ftp.Pass = task.ftpPass
	ftp.ServerDir = task.ServerDir
	ftp.LocalDir = localDir + "/" + task.name
	// 参数获取完成
	log.Println("Mysql备份中...")
	err = uitl.Mysql(mysql)
	if err != nil {
		log.Printf("MYSQL备份失败：%s\n", err)
		return err
	}
	log.Println("Mysql备份成功")
	log.Println("FTP下载中...")
	err = uitl.Ftp(ftp)
	if err != nil {
		fmt.Printf("FTP下载失败:%s\n", err)
		return err
	}
	log.Println("FTP下载成功")
	//Zip打包
	zipConfing := &uitl.ZipConf{}
	zipConfing.LocalDir = localDir + "/" + task.name
	zipConfing.File = "web.zip"
	log.Println("Zip打包中...")
	err = uitl.Zip(zipConfing)
	if err != nil {
		fmt.Println(err)
		log.Println("Zip打包失败")
		return err
	}
	log.Println("Zip打包成功")
	return nil
}

func zdyMysql(task bakTask) {
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

func zdyFtp(task bakTask) {
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

func zdyZip(task bakTask) {
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
