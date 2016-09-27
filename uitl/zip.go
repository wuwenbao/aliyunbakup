package uitl

import (
	"archive/zip"
	"io/ioutil"
	"os"
)

//ZipConf 外调配置
type ZipConf struct {
	LocalDir string
	File     string
}

//DaBao 内部配置
type DaBao struct {
	zip  *zip.Writer
	conf *ZipConf
}

//Zip 打包压缩
func Zip(conf *ZipConf) error {
	db := new(DaBao)
	db.conf = conf
	os.Remove(conf.LocalDir + "/" + conf.File)
	fw, err := os.Create(conf.LocalDir + "/" + conf.File)
	if err != nil {
		return err
	}
	defer fw.Close()
	db.zip = zip.NewWriter(fw)
	defer db.zip.Close()

	err = db.cdir(conf.LocalDir, "")
	if err != nil {
		return err
	}
	return nil
}

//cdir 遍历打包
func (db *DaBao) cdir(local, zipdir string) error {
	// 打开文件夹
	dir, err := os.Open(local)
	if err != nil {
		return err
	}
	defer dir.Close()
	// 读取文件列表
	fielLists, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	var zipTemp string
	for _, file := range fielLists {
		if file.Name() == "web.zip" {
			continue
		}
		if zipdir == "" {
			zipTemp = file.Name()
		} else {
			zipTemp = zipdir + "/" + file.Name()
		}
		if file.IsDir() {
			_, err := db.zip.Create(zipTemp + "/")
			if err != nil {
				return err
			}
			err = db.cdir(dir.Name()+"/"+file.Name(), zipTemp)
			if err != nil {
				return err
			}
		} else {
			fr, err := os.Open(dir.Name() + "/" + file.Name())
			if err != nil {
				return err
			}
			defer fr.Close()
			fd, err := ioutil.ReadAll(fr)
			f, err := db.zip.Create(zipTemp)
			if err != nil {
				return err
			}
			_, err = f.Write([]byte(fd))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
