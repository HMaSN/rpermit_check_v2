/*
Description: Code to check if the resident permit is issued.
Author: Hayssam Noweir - 05/2017
*/
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hoisie/redis"
)

const (
	// CheckURL is the URL used to check the issuing of the Permit status
	CheckURL string = "https://www17.muenchen.de/EATWebSearch/Auskunft"
	// RedisURL is Redis instance URL
	RedisURL string = "<REDIS_URL>"
	// RedisUser is Redis instance user
	RedisUser string = "<REDIS_USER>"
	// RedisPassword is Redis instance password
	RedisPassword string = "<REDIS_PASSWORD>"
	// zapNummer is the Permit request ID
	zapNummer string = "<PERMIT_ID>"
	// MailGunURL is the API URL for mailgun
	MailGunURL string = "<MailGunURL>"
	// MailGunAPIToken is the api token for Mailgun
	MailGunAPIToken string = "<MailGunAPIToken>"
)

func getResponse() (response string, statusCode int) {
	client := &http.Client{}
	data := []byte("zapnummer=" + zapNummer + "&pbAbfragen=Auskunft")
	req, _ := http.NewRequest("POST", CheckURL, bytes.NewBuffer(data))
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Add("Referer", "https://www17.muenchen.de/EATWebSearch/")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Origin", "https://www17.muenchen.de")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, _ := client.Do(req)
	defer resp.Body.Close()
	response2, _ := ioutil.ReadAll(resp.Body)
	response = string(response2)
	return response, resp.StatusCode
}

func getRedisClient() *redis.Client {
	var client redis.Client
	client.Addr = RedisURL
	client.Password = RedisPassword
	client.Auth(RedisUser + ":" + RedisPassword)
	return &client
}
func getRedisKey(key string) []byte {
	val, _ := getRedisClient().Get(key)
	return val
}
func setRedisKey(key string, value []byte) {
	getRedisClient().Set(key, value)
}
func main() {
	log.SetFlags(log.Ldate | log.Ltime)
	log.Println(" - INFO - Application Started")
	g := getRedisKey("IS_ISSUED")
	if string(g) == "YES" {
		log.Println(" - INFO - Resident card already issued. Quitting")
		return
	}
	log.Println(" - INFO - Value of IS_ISSUED key is " + string(g))

	resp, status := getResponse()
	if status != 200 {
		panic("Status code is : " + string(status))
	}
	if strings.Contains(resp, "liegt noch nicht zur Abholung bereit") {
		log.Println(" - INFO - Resident Permit Card is not yet issued")
		setRedisKey("IS_ISSUED", []byte("NO"))
	} else {
		setRedisKey("IS_ISSUED", []byte("YES"))
		subject := "Update found regaring your resident Permit"
		message := `Hello,
	 <br><br>An update has been found regarding your Resident permit, please check in the site
	 <br><br><a href="https://www17.muenchen.de/EATWebSearch/">https://www17.muenchen.de/EATWebSearch/</a>
	 <br><br> Your ID is ` + zapNummer + `
	 <br><br>
	 This is an automated email sent by rpermit_check_v2`
		res, status := sendEmail("<SENDER>",
			"<RECIEPIENT>", "", subject, message)
		log.Println(" - INFO - Email Status: " + status + "\nResponse " + res)
	}

}

func sendEmail(sender, reciepient, cc, subject, message string) (response, status string) {
	client := &http.Client{}
	f := url.Values{}
	f.Add("from", sender)
	f.Add("to", reciepient)
	if cc != "" {
		f.Add("cc", cc)
	}
	f.Add("subject", subject)
	f.Add("html", message)

	req, _ := http.NewRequest("POST", MailGunURL, strings.NewReader(f.Encode()))
	req.SetBasicAuth("api", MailGunAPIToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	defer resp.Body.Close()
	resp2, _ := ioutil.ReadAll(resp.Body)
	return string(resp2), resp.Status
}
