package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"encoding/json"

	"fmt"

	"github.com/inimei/backup/log"
	"github.com/inimei/ddns/web/signature"
)

var (
	ipServer = "http://pv.sohu.com/cityjson?ie=utf-8"
)

func main() {

	log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	defer log.Close()

	if len(Data.Key) == 0 {
		log.Error("key is emptry...")
		return
	}

	if len(Data.ServerHost) == 0 {
		log.Error("server_host is emptry...")
		return
	}

	if Data.TickTime <= 0 {
		Data.TickTime = 60
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	tick := time.NewTicker(time.Duration(Data.TickTime) * time.Second)

forever:
	for {
		select {
		case <-sig:
			log.Debug("signal received, stopping")
			break forever
		case <-tick.C:
			go update()
		}
	}
}

func update() {
	log.Info("update ip")

	ip, err := getIp()
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = postIp(ip)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func getIp() (string, error) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequest("GET", ipServer, strings.NewReader(""))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Encoding", "")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	datas := strings.Split(string(body), "=")
	if len(datas) != 2 {
		return "", errors.New("get data failed: " + string(body))
	}

	var IpData struct {
		Cip   string `json:"cip"`
		Cid   string `json:"cid"`
		Cname string `json:"cname"`
	}

	rowData := datas[1]
	log.Debug(rowData)
	rowData = strings.Trim(rowData, ";")

	err = json.Unmarshal([]byte(rowData), &IpData)
	if err != nil {
		err = fmt.Errorf("Unmarshal data failed: %v", err)
		log.Error(err.Error())
		return "", err
	}

	return IpData.Cip, nil
}

func postIp(ip string) error {

	values := make(url.Values)
	values.Set("key", Data.Key)
	values.Set("ip", ip)

	trans := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: trans}

	data := values.Encode()
	url := "/api/update"

	req, err := http.NewRequest("POST", Data.ServerHost+url, strings.NewReader(data))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	signature.SignRequest(req, data, Data.Accesskey, Data.SecretKey, "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("update failed http code(%v)", resp.StatusCode)
		log.Error(err.Error())
		return errors.New(err.Error())
	}

	var rsp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = fmt.Errorf("Unmarshal rsp data [%v] failed: %v", string(body), err)
		log.Error(err.Error())
		return err
	}

	if rsp.Code != "ok" {
		err = fmt.Errorf("update ip failed, code=%v, msg=%v", rsp.Code, rsp.Msg)
		log.Error(err.Error())
		return err
	}

	return nil
}
