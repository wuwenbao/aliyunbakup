package uitl

import (
	"container/list"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jlaffaye/ftp"
)

//Task 任务
type Task struct {
	Name    string    //任务名称
	Type    int       //文件类型
	To      int       //方向
	Local   string    //本地目录
	Server  string    //服务器目录
	Size    uint64    //文件大小
	Level   int       //优先级
	Time    time.Time //时间
	Status  int       //状态
	Message string    //原因
}

//FtpConf ftp配置参数
type FtpConf struct {
	Host      string
	User      string
	Pass      string
	LocalDir  string
	ServerDir string
}

//Quest 队列
type quest struct {
	confing      *FtpConf        //ftp参数
	items        *list.List      //队列
	itemsSuccess *list.List      //成功列表
	itemsError   *list.List      //失败列表
	con          *ftp.ServerConn //ftp连接
}

func init() {

}

//Ftp 初始化
func Ftp(conf *FtpConf) error {
	q := new(quest)
	q.confing = conf
	c, err := ftp.Connect(q.confing.Host)
	if err != nil {
		return err
	}
	err = c.Login(q.confing.User, q.confing.Pass)
	if err != nil {
		return err
	}
	err = c.ChangeDir(q.confing.ServerDir)
	if err != nil {
		return err
	}
	q.items = list.New()
	q.itemsSuccess = list.New()
	q.itemsError = list.New()
	q.con = c
	err = q.Run()
	if err != nil {
		return err
	}
	q.Quest()
	//打印记录
	log.Printf("任务队列：%10d 条\n", q.items.Len())
	log.Printf("成功队列：%10d 条\n", q.itemsSuccess.Len())
	log.Printf("失败队列：%10d 条\n", q.itemsError.Len())
	return nil
}

//Quest 退出
func (q *quest) Quest() {
	q.con.Logout()
	q.con.Quit()
}

//Run 运行
func (q *quest) Run() error {
	log.Print("FTP任务队列添加...")
	err := q.makeList(q.confing.ServerDir, q.confing.LocalDir)
	if err != nil {
		return err
	}
	for e := q.items.Front(); e != nil; e = e.Next() {
		task := e.Value.(Task)
		err := q.wget(task)
		if err != nil {
			return err
		}
	}
	// 第一次失败重新尝试
	for e := q.itemsError.Front(); e != nil; e = e.Next() {
		task := e.Value.(Task)
		err := q.wget(task)
		if err != nil {
			return err
		}
	}
	// 第二次失败重新尝试
	for e := q.itemsError.Front(); e != nil; e = e.Next() {
		task := e.Value.(Task)
		err := q.wget(task)
		if err != nil {
			return err
		}
	}
	return nil
}

//wget 从服务器中下载到本地
func (q *quest) wget(task Task) error {
	r, err := q.con.Retr(task.Server)
	if err != nil {
		q.itemsError.PushBack(task)
		return err
	}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		q.itemsError.PushBack(task)
		return err
	}
	err = ioutil.WriteFile(task.Local, buf, 0644)
	if err != nil {
		q.itemsError.PushBack(task)
		return err
	}
	r.Close()
	q.itemsSuccess.PushBack(task)
	log.Printf("下载文件：%s", task.Server)
	return nil
}

//makeList 任务文件压入队列
func (q *quest) makeList(serverDir, localDir string) error {
	_, err := os.Stat(localDir)
	if err != nil {
		err = os.MkdirAll(localDir, os.ModePerm) //生成多级目录
		if err != nil {
			return err
		}
	}
	lists, err := q.con.List(serverDir)
	if err != nil {
		return err
	}
	for _, val := range lists {
		serverTemp := serverDir + "/" + val.Name
		localTemp := localDir + "/" + val.Name
		if val.Type == 1 {
			err = q.makeList(serverTemp, localTemp)
			if err != nil {
				return err
			}
		} else {
			task := Task{}
			task.Name = val.Name
			task.Type = 1
			task.To = 1
			task.Local = localTemp
			task.Server = serverTemp
			task.Size = val.Size
			task.Level = 0
			task.Time = time.Now()
			task.Status = 1
			q.items.PushBack(task)
		}
	}
	return nil
}
