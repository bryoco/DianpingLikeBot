package main

import (
	"bytes"
	// colly was edited to emulate browser activity, to circumvent bot-preventing captcha.
	// original repo @ https://github.com/gocolly/colly
	"DianpingLikeBot/colly"
	mapset "github.com/deckarep/golang-set"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const ReviewPage = "https://www.dianping.com/member/{any_given_user_id}/reviews"
const Referer = "https://www.dianping.com/member/{any_given_user_id}/reviews?pg=1&reviewCityId=0&reviewShopType=0&c=0&shopTypeIndex=0"

func main() {
	jar := ParseCookie(ReviewPage)
	s := mapset.NewSet()

	collectReview(jar, s)
	log.Println(s.String())

	itr := s.Iterator()
	for i := range itr.C {
		log.Printf("posting like at %v", i)
		postLike(i.(string), jar)
	}

	// test one id
	//postLike("564465801", jar)
}

func collectReview(jar *cookiejar.Jar, s mapset.Set) {
	c := colly.NewCollector()
	c.SetCookieJar(jar)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		dataID := e.Attr("data-id")
		// filter reviews that has already been liked
		// unliked = class="iheart heart-bg heart-s1"
		// liked = class="iheart heart-bg heart-s3"
		unliked := strings.Contains(e.ChildAttr("i", "class"), "heart-s1")
		if len(dataID) != 0 && unliked {
			s.Add(dataID)
		}

		// find next review page
		dataPg := e.Attr("data-pg")
		if len(dataPg) > 0 {
			nextPage := e.Attr("href")
			_ = e.Request.Visit(nextPage)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		// test only
		//log.Println("Visiting:", r.URL)
	})

	err := c.Visit(ReviewPage)
	if err != nil {
		log.Fatalf("cannot execute page [%v]: %v", ReviewPage, err)
	}
}

func postLike(i string, jar *cookiejar.Jar) {

	const api = "https://www.dianping.com/ajax/json/shop/reviewflower"
	client := http.Client{
		Jar:           jar,
	}

	// got to send two requests, "do=a" and "do=aa", respectively, apparently.
	payload1 := []byte("do=a&i=" + i + "&t=1&s=2")
	payload2 := []byte("do=aa&i=" + i + "&t=1&s=2")
	req1, _ := http.NewRequest(http.MethodPost, api, bytes.NewBuffer(payload1))
	setHeaders(req1)
	req2, _ := http.NewRequest(http.MethodPost, api, bytes.NewBuffer(payload2))
	setHeaders(req2)

	// send twice
	resp, err := client.Do(req1)
	if err != nil {
		log.Printf("cannot request id %v: %v", i, err)
	}
	//body, err := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))

	resp, err = client.Do(req2)
	if err != nil {
		log.Printf("cannot request id %v: %v", i, err)
	}
	//body, err = ioutil.ReadAll(resp.Body)
	//log.Println(string(body))

	log.Printf("like posted at %v, code %v", i, resp.StatusCode)
}

func setHeaders(req *http.Request) {
	// server doesnt seem to care headers that much
	// left for reference, not fully tested

	//req.Header.Set("Accept", "application/json, text/javascript")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	//req.Header.Set("Accept-Language", "en-US,en;q=0.9,ja;q=0.8,zh-CN;q=0.7,zh;q=0.6,zh-TW;q=0.5")
	//req.Header.Set("Cache-Control", "no-cache")
	//req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8;")
	//req.Header.Set("DNT", "1")
	req.Header.Set("Host", "www.dianping.com")
	req.Header.Set("Origin", "http://dianping.com")
	//req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", Referer)
	//req.Header.Set("Sec-Fetch-Mode", "cors")
	//req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	//req.Header.Set("X-Request", "JSON")
	//req.Header.Set("X-Requested-With", "XMLHttpRequest")
}
