package main

import (
	"bufio"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type E struct {
	hash map[string]bool
	lock sync.RWMutex // 加锁
}

func (m *E) Set(key string, value bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.hash[key] = value
}

func (m *E) Get(key string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, res := m.hash[key]
	return res
}

func main() {
	done := make(chan int)
	var events = new(E)
	events.hash = make(map[string]bool)

	const (
		seleniumPath = `/chromedriver.exe`
		port         = 9515
	)
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(seleniumPath, port, opts...)
	if nil != err {
		fmt.Println("start a chromedriver service falid", err.Error())
		return
	}
	defer service.Stop()
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	imagCaps := map[string]interface{}{
		"profile.managed_default_content_settings.images": 2,
	}

	chromeCaps := chrome.Capabilities{
		Prefs: imagCaps,
		Path:  "",
		Args: []string{
			//"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			//"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36",
		},
	}

	caps.AddChrome(chromeCaps)
	wb, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		fmt.Println("connect to the webDriver faild", err.Error())
		return
	}
	defer wb.Quit()
	defer wb.Close()
	err = wb.Get("https://web.telegram.org/")

	if err != nil {
		fmt.Println("get page faild", err.Error())
	}

	for {
		codeElem, err := wb.FindElement(selenium.ByXPATH, "//*[@id='ng-app']/body/div[1]/div/div[2]/div[2]/form/div[2]/div[1]/input")
		if err != nil {
			continue
		}
		codeElem.Clear()
		codeElem.SendKeys("+63")
		break
	}
	phoneElem, _ := wb.FindElement(selenium.ByXPATH, "//*[@id='ng-app']/body/div[1]/div/div[2]/div[2]/form/div[2]/div[1]/input")
	fmt.Println("输入手机号：")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	phoneElem.SendKeys(input.Text())

	nextElem, _ := wb.FindElement(selenium.ByClassName, "login_head_submit_btn")
	nextElem.Click()

	for {
		subElem, err := wb.FindElement(selenium.ByClassName, "btn-md-primary")
		if err != nil {
			fmt.Println("提交号码卡住")
			continue
		}
		subElem.Click()
		break
	}

	fmt.Println("输入验证码：")
	input = bufio.NewScanner(os.Stdin)
	input.Scan()

	verifyElem, err := wb.FindElement(selenium.ByXPATH, "//*[@id='ng-app']/body/div[1]/div/div[2]/div[2]/form/div[4]/input")
	verifyElem.SendKeys(input.Text())

	time.Sleep(3 * time.Second)

	gointo(wb,"gz技术部考勤群")

	go test(wb)
	go sign(wb)
	go punch(wb, events, 13)
	<-done
}

func gointo(wb selenium.WebDriver,name string)  {
I:
	for {
		elements, err := wb.FindElements(selenium.ByXPATH, "//li[@class='im_dialog_wrap']")
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		for _, element := range elements {
			e, _ := element.FindElement(selenium.ByXPATH, ".//span[@my-peer-link='dialogMessage.peerID']")
			result, _ := e.Text()
			//fmt.Println(result)
			if result == name {
				element.Click()
				fmt.Println("进入监测区 "+name)
				break I
			}
		}
	}
}

func test(wb selenium.WebDriver)  {
	for {
		t := time.Now()
		h := t.Hour()
		m := t.Minute()
		if h > 21 || h < 10 {
			if h%2 == 0 && m > 30 {
				gointo(wb, "背多分 背多分")
				sendElem, _ := wb.FindElement(selenium.ByClassName, "composer_rich_textarea")
				sendElem.SendKeys("hikki机器人还在线上")
				sendElem.SendKeys(selenium.EnterKey)
				gointo(wb, "gz技术部考勤群")
			}
			time.Sleep(50 * time.Minute)
		}else {
			time.Sleep(30 * time.Minute)
		}
	}
}

func sign(wb selenium.WebDriver) {
	for {
		t := time.Now()
		h := t.Hour()
		m := t.Minute()
		if h == 10 && m > 45 && m < 48 {
			fmt.Println("开工打卡")
			mess := "上班打卡：早11:00-晚8:00"
			if t.Weekday() == time.Sunday {
				mess = "hikki 休假"
			}
			sendElem, erro := wb.FindElement(selenium.ByClassName, "composer_rich_textarea")
			if erro != nil {
				fmt.Println(erro)
				continue
			}
			sendElem.SendKeys(mess)
			sendElem.SendKeys(selenium.EnterKey)

			time.Sleep(5 * time.Hour)
		}
		if t.Weekday() == time.Sunday {
			continue
		}
		if h == 20 && m > 5 && m < 10 {
			fmt.Println("下班打卡")
			mess := "下班打卡：早11:00-晚8:00"
			sendElem, erro := wb.FindElement(selenium.ByClassName, "composer_rich_textarea")
			if erro != nil {
				continue
			}
			sendElem.SendKeys(mess)
			sendElem.SendKeys(selenium.EnterKey)
			time.Sleep(8 * time.Hour)
		}
	}
}

func punch(wb selenium.WebDriver, events *E, num int) {
	temp := num
	for {
		xpath := "//div[@class='im_history_message_wrap'][last() - " + strconv.Itoa(num) + "]"
		num--
		if num<0 {
			num = temp
		}
		lastElem, err := wb.FindElement(selenium.ByXPATH, xpath)
		if err != nil {
			continue
		}
		textElem, err := lastElem.FindElement(selenium.ByXPATH, ".//div[@class='im_message_text']")
		if err != nil {
			continue
		}
		mess, _ := textElem.Text()
		authorElem, err := lastElem.FindElement(selenium.ByXPATH, ".//span[@class='im_message_author_wrap']/a")
		if err != nil {
			continue
		}
		author, _ := authorElem.Text()
		reg, _ := regexp.MatchString("(.*)报(.*)数(.*)", mess)
		if author == "GZ Shanna" && reg {
			t := time.Now()
			event := fmt.Sprintf("%d-%02d-%02dT%02d",
				t.Year(), t.Month(), t.Day(),
				t.Hour())

			ok := events.Get(event)
			if ok {
				//已经创建
				continue
			}
			fmt.Println("任务执行")

			events.Set(event, true)
			//报数
			sendElem, erro := wb.FindElement(selenium.ByClassName, "composer_rich_textarea")
			if erro != nil {
				continue
			}
			sendElem.SendKeys("1")
			time.Sleep(1*time.Second)
			sendElem.SendKeys(selenium.EnterKey)
			time.Sleep(1 * time.Hour)
		}
		time.Sleep(500*time.Millisecond)
	}
}
