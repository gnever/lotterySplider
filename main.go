package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/util/gutil"
)

type twoColorBall struct {
	No       int    `orm:"no"`
	Red1     int    `orm:"red1"`
	Red2     int    `orm:"red2"`
	Red3     int    `orm:"red3"`
	Red4     int    `orm:"red4"`
	Red5     int    `orm:"red5"`
	Red6     int    `orm:"red6"`
	Blue1    int    `orm:"blue1"`
	Date     string `orm:"date"`
	CreateTS int64  `orm:"create_ts"`
}

var (
	l         *glog.Logger
	currentNo int
	curlStop  bool //数据库内最大的彩票期号 >= 抓取的彩票期号 则不用再抓取
	outFile   string
)

func init() {
	flag.StringVar(&outFile, "o", "", "输出目标文件")
	flag.Parse()
}

func main() {

	l = glog.New().Line(true)

	db := g.DB()

	one, err := db.GetOne("select no from two_color_ball order by no desc limit 1")

	if err == nil {
		currentNo = one.Map()["no"].(int)
		l.Infof("数据库内最新数据为 %d 期", currentNo)
	}

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

		l.Infof("正在抓取第 %d 页", i)

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
				l.Infof("抓取 %s 失败", url)
			}
			ct := response.ReadAllString()
			response.Close()
			content = ct
		}
		parseHTML(content, treeMap)
	}

	if treeMap.Size() == 0 {
		l.Info("没有需要处理的数据")
		return
	}

	for _, v := range treeMap.Values() {
		vv := v.(*twoColorBall)

		_line := fmt.Sprintf("%d,%s,%d,%d,%d,%d,%d,%d,%d\n", vv.No, vv.Date, vv.Red1, vv.Red2, vv.Red3, vv.Red4, vv.Red5, vv.Red6, vv.Blue1)

		l.Debugf("写入 ： %s", _line)

		if outFile != "" {
			gfile.PutContentsAppend(outFile, _line)
		}

		vv.CreateTS = time.Now().Unix()
		_, err := db.Table("two_color_ball").Data(vv).Insert()
		if err != nil {
			l.Errorf("插入失败 err: %s", err)
		}

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

		//l.Debugf("抓取到 ： %s--%d %d %d %d %d %d %d  %d\n", _twoColorBall.Date, _twoColorBall.No, _twoColorBall.Red1, _twoColorBall.Red2, _twoColorBall.Red3, _twoColorBall.Red4, _twoColorBall.Red5, _twoColorBall.Red6, _twoColorBall.Blue1)

		treeMap.Set(arr[2], _twoColorBall)
	}
	return nil
}
