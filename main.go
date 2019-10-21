package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/util/gutil"
)

type twoColorBall struct {
	No    int
	Red1  int
	Red2  int
	Red3  int
	Red4  int
	Red5  int
	Red6  int
	Blue1 int
	Date  string
}

var (
	l         *glog.Logger
	currentNo int
	curlStop  bool //数据库内最大的彩票期号 >= 抓取的彩票期号 则不用再抓取
)

func main() {

	l = glog.New().Line(true)

	client := ghttp.NewClient()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.120 Safari/537.36")
	client.SetTimeOut(time.Second * 10)

	response, err := client.Get("http://kaijiang.zhcw.com/zhcw/html/ssq/list.html")
	if err != nil {
		l.Panicf("curl fail %s", err)
	}

	ct := response.ReadAllString()
	response.Close()

	//fmt.Println(string(ct))
	match, _ := gregex.MatchString(`共<strong>(\d.+)</strong> 页`, ct)
	if len(match[1]) < 2 {
		l.Panicf("获取总页数失败 %v", match)
	}
	pages, serr := strconv.Atoi(match[1])
	if serr != nil {
		l.Panicf("页数转码失败 %s", match[1])
	}

	treeMap := gmap.NewTreeMap(gutil.ComparatorInt, true)

	for i := 1; i <= pages; i++ {
		var content string

		//第一页已经取了，不用再处理
		if i == 1 {
			content = ct
		} else {
			if curlStop {
				l.Infof("当前数据库内期号为：%d, 已成功抓取到该期号之前的数据", currentNo)
				break
			}

			time.Sleep(time.Second*time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10)) + 5)
			url := fmt.Sprintf("http://kaijiang.zhcw.com/zhcw/html/ssq/list_%d.html", i)
			response, err := client.Get(url)
			if err != nil {
				l.Errorf("抓取 %s 失败", url)
			}
			ct := response.ReadAllString()
			response.Close()
			content = ct
		}
		parseHTML(content, treeMap)
	}

	if treeMap.Size() == 0 {
		l.Error("未抓取到数据")
		return
	}

	for _, v := range treeMap.Values() {
		vv := v.(*twoColorBall)
		l.Infof("写入 ： %s--%d %d %d %d %d %d %d  %d\n", vv.Date, vv.No, vv.Red1, vv.Red2, vv.Red3, vv.Red4, vv.Red5, vv.Red6, vv.Blue1)
	}
}

func parseHTML(content string, treeMap *gmap.TreeMap) (err error) {
	match, err := gregex.MatchAllString(`<tr>\s*?<td align="center">([0-9,\-]*?)<\/td>\s*?<td align="center">([0-9]*?)<\/td>\s*?<td align="center" style="padding-left:10px;">\s*?<em class="rr">(\d{2})<\/em>\s*?<em class="rr">(\d{2})<\/em>\s*?<em class="rr">(\d{2})<\/em>\s*?<em class="rr">(\d{2})<\/em>\s*?<em class="rr">(\d{2})<\/em>\s*?<em class="rr">(\d{2})<\/em>\s*?<em>(\d{2})<\/em><\/td>(?s).*?<\/tr>`, content)

	if err != nil {
		return
	}

	for _, arr := range match {
		_twoColorBall := new(twoColorBall)
		_twoColorBall.Date = arr[1]
		_twoColorBall.No, _ = strconv.Atoi(arr[2])
		_twoColorBall.Red1, _ = strconv.Atoi(arr[3])
		_twoColorBall.Red2, _ = strconv.Atoi(arr[4])
		_twoColorBall.Red3, _ = strconv.Atoi(arr[5])
		_twoColorBall.Red4, _ = strconv.Atoi(arr[6])
		_twoColorBall.Red5, _ = strconv.Atoi(arr[7])
		_twoColorBall.Red6, _ = strconv.Atoi(arr[8])
		_twoColorBall.Blue1, _ = strconv.Atoi(arr[9])

		if currentNo >= _twoColorBall.No {
			curlStop = true
			return nil
		}

		l.Infof("抓取到 ： %s--%d %d %d %d %d %d %d  %d\n", _twoColorBall.Date, _twoColorBall.No, _twoColorBall.Red1, _twoColorBall.Red2, _twoColorBall.Red3, _twoColorBall.Red4, _twoColorBall.Red5, _twoColorBall.Red6, _twoColorBall.Blue1)

		treeMap.Set(arr[2], _twoColorBall)
	}
	return nil
}
