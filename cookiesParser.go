package main

import (
	"encoding/json"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"
)

const Domain = "https://www.dianping.com"

type RawCookies struct {
	RawCookies []*RawCookie
}

type RawCookie struct {
	Domain		string		`json:"domain"`
	Expires		float64		`json:"expirationDate"`
	HostOnly	bool		`json:"hostOnly"`
	HttpOnly	bool		`json:"httpOnly"`
	Name		string		`json:"name"`
	Path		string		`json:"path"`
	SameSite	string		`json:"sameSite"`
	Secure		bool		`json:"secure"`
	Session		bool		`json:"session"`
	StoreId		string		`json:"storeId"`
	Value		string		`json:"value"`
	Id			int			`json:"id"`
}

// parse raw cookies in json form into http.Cookie
// raw cookies are outputted using Chrome extension EditThisCookie @ http://www.editthiscookie.com/
func ParseCookie(domain string) *cookiejar.Jar {
	u, err := url.Parse(domain)
	if err != nil {
		log.Fatalf("cannot parse url: %v", err)
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("rawCookies.json")
	if err != nil {
		log.Fatalln(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	var rawCookies []RawCookie
	err = json.Unmarshal(b, &rawCookies)
	if err != nil {
		log.Fatalln(err)
	}

	var cookies []*http.Cookie
	for _, rc := range rawCookies {
		cookies = append(cookies, rc.cookieMarshaller())
	}

	jar.SetCookies(u, cookies)

	return jar
}

// parse expiration data from float to int
func (rc *RawCookie) getTime() time.Time {
	sec, dec := math.Modf(rc.Expires)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}

func (rc *RawCookie) parseSameSite() int {
	if len(rc.SameSite) > 1 {
		return 0
	} else {
		i, err := strconv.Atoi(rc.SameSite)
		if err != nil {
			return 0
		} else {
			return i
		}
	}
}

func (rc *RawCookie) cookieMarshaller() *http.Cookie {
	return &http.Cookie{
		Name:       rc.Name,
		Value:      rc.Value,
		Path:       rc.Path,
		Domain:     rc.Domain,
		Expires:    rc.getTime(),
		Secure:     rc.Secure,
		HttpOnly:   rc.HttpOnly,
		SameSite:   http.SameSite(rc.parseSameSite()),
	}
}