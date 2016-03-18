package icy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Portal struct {
	Session
	Status
}

type Session struct {
	Serial string `json:"serialthermostat1"`
	Token  string `json:"token"`
}

type Status struct {
	Current       float64 `json:"temperature2"`
	Target        float64 `json:"temperature1"`
	Configuration []int   `json:"configuration"`
}

func (portal *Portal) Login(username string, password string) error {
	res, err := http.PostForm("https://portal.icy.nl/login",
		url.Values{"username": {username}, "password": {password}, "remember": {"1"}})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&portal.Session)
	if err != nil {
		return err
	}
	return nil
}

func (portal *Portal) Read() error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://portal.icy.nl/data", nil)
	req.Header.Add("Session-Token", portal.Session.Token)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&portal.Status)
	if err != nil {
		return err
	}

	return nil
}

func (portal Portal) Write() error {
	client := &http.Client{}

	data := url.Values{
		"uid":          {portal.Session.Serial},
		"temperature1": {fmt.Sprintf("%.1f", portal.Status.Target)},
	}
	for _, c := range portal.Status.Configuration {
		data.Add("configuration[]", strconv.Itoa(c))
	}

	fmt.Println(data.Encode())

	req, err := http.NewRequest("POST", "https://portal.icy.nl/data", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Session-Token", portal.Session.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Status code %s", res.Status))
	}

	return nil
}

func (portal Portal) Temperature() float64 {
	return portal.Status.Current
}

func (portal Portal) TargetTemperature() float64 {
	return portal.Status.Target
}

func (portal *Portal) SetTargetTemperature(value float64) {
	portal.Status.Target = value
}

type Mode byte

const (
	Comfort Mode = 0x00
	Saving  Mode = 0x01
	Away    Mode = 0x02
	Fixed   Mode = 0x03
)

func (portal Portal) Mode() Mode {
	if portal.Status.Configuration[0]&128 == 128 {
		return Fixed
	} else if portal.Status.Configuration[0]&64 == 64 {
		return Saving
	} else if portal.Status.Configuration[0]&32 == 32 {
		return Comfort
	} else {
		return Away
	}
}

func (portal *Portal) SetMode(value Mode) {
	switch value {
	case Away:
		portal.Status.Configuration[0] = 0
	case Comfort:
		portal.Status.Configuration[0] = 32
	case Saving:
		portal.Status.Configuration[0] = 64
	case Fixed:
		portal.Status.Configuration[0] = 160 // @todo this doesn't work?
	}
}
