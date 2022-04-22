package main

import (
	"crypto/md5"
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type pool struct {
	queue chan int
	wg    *sync.WaitGroup
}

func New(size int) *pool {
	if size <= 0 {
		size = 1
	}
	return &pool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

func (p *pool) Add(delta int) {
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
	for i := 0; i > delta; i-- {
		<-p.queue
	}
	p.wg.Add(delta)
}

func (p *pool) Done() {
	<-p.queue
	p.wg.Done()
}

func (p *pool) Wait() {
	p.wg.Wait()
}

//获取指定目录及所有子目录下的所有文件
func WalkDir(dirPth string) (files []string, err error) {
	files = make([]string, 0, 30)

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录

		if fi.IsDir() { // 忽略目录
			return nil
		}
		files = append(files, filename)
		return nil
	})

	return files, err
}

// 获取文件的SHA校验
func cryptoSum(filepath string, cryptoType string, res *string) {
	f, err := os.Open(filepath)

	if err != nil {
		fmt.Println(err)
		*res = "Cannot find file " + filepath + ". make sure the path is correct."
		return
	}
	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		*res = "The format of the file is incorrect."
		return
	}
	if cryptoType == "SHA1" {
		*res = fmt.Sprintf("%x", sha1.Sum(body))
	} else {
		*res = fmt.Sprintf("%x", md5.Sum(body))
	}
	runtime.GC()
}
func main() {
	// 定义命令行参数
	var (
		path       string
		cryptoType string
		thread     int
	)
	flag.StringVar(&path, "path", "./", "set Path.")
	flag.StringVar(&cryptoType, "type", "MD5", "SHA1 or MD5, default is MD5.")
	flag.IntVar(&thread, "thread", 10, "Number of concurrent threads.")
	flag.Parse()

	fmt.Println(thread)
	files, err := WalkDir(path)
	// 队列
	pool := New(thread)
	// 结果
	res := make([]string, len(files))
	if err != nil {
		fmt.Println(err)
	}
	for i, f := range files {
		pool.Add(1)
		go func(f string, cryptoType string, t *string) {

			defer pool.Done()
			cryptoSum(f, cryptoType, t)
		}(f, cryptoType, &res[i])
	}
	// 等待
	pool.Wait()
	// 输出
	for i, f := range files {
		fmt.Println(res[i], " * ", f)
	}
}
