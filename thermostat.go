package icy

import (
	"log"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
)

func (portal Portal) CurrentHeatingCoolingState() int {
	if portal.IsHeating() {
		return characteristic.CurrentHeatingCoolingStateHeat
	}
	if portal.Mode() == Comfort {
		return characteristic.CurrentHeatingCoolingStateCool
	}
	return characteristic.CurrentHeatingCoolingStateOff
}

func (portal Portal) TargetHeatingCoolingState() int {
	switch portal.Mode() {
	case Fixed:
		return characteristic.TargetHeatingCoolingStateOff
	case Away:
		fallthrough
	case Saving:
		fallthrough
	case Comfort:
		return characteristic.TargetHeatingCoolingStateAuto
	default:
		panic("Unexpected mode")
	}
}

func NewThermostat(name string, username string, password string) (*accessory.Thermostat, error) {
	portal := Portal{}
	err := portal.Login(username, password)
	if err != nil {
		log.Fatal(err)
	}

	err = portal.Read()
	if err != nil {
		return nil, err
	}

	thermostat := accessory.NewThermostat(accessory.Info{
		Name:         name,
		Manufacturer: "ICY",
		SerialNumber: portal.Session.SerialNumber,
	}, portal.TargetTemperature(), 0, 30, 0.5)

	thermostat.Thermostat.CurrentTemperature.SetValue(portal.Temperature())
	thermostat.Thermostat.TargetTemperature.SetValue(portal.TargetTemperature())
	thermostat.Thermostat.CurrentHeatingCoolingState.SetValue(portal.CurrentHeatingCoolingState())
	thermostat.Thermostat.TargetHeatingCoolingState.SetValue(portal.TargetHeatingCoolingState())

	thermostat.Thermostat.TargetTemperature.OnValueRemoteUpdate(func(value float64) {
		portal.SetTargetTemperature(value)
		portal.Write()
	})
	thermostat.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(func(value int) {
		var newTemperature float64
		switch value {
		case characteristic.TargetHeatingCoolingStateOff:
			newTemperature = float64(portal.Configuration[4]) / 2
			portal.SetMode(Fixed)
		case characteristic.TargetHeatingCoolingStateCool:
			newTemperature = float64(portal.Configuration[5]) / 2
			portal.SetMode(Saving)
		case characteristic.TargetHeatingCoolingStateHeat:
			fallthrough
		case characteristic.TargetHeatingCoolingStateAuto:
			newTemperature = float64(portal.Configuration[6]) / 2
			portal.SetMode(Comfort)
		}
		portal.SetTargetTemperature(newTemperature)
		thermostat.Thermostat.TargetTemperature.SetValue(portal.TargetTemperature())
		err := portal.Write()
		if err != nil {
			log.Fatal(err)
		}
	})

	// Refresh every 20 seconds
	ticker := time.NewTicker(time.Millisecond * 1000 * 20)
	go func() {
		for _ = range ticker.C {
			err := portal.Read()
			if err != nil {
				log.Fatal(err)
				return
			}
			thermostat.Thermostat.CurrentTemperature.SetValue(portal.Temperature())
			thermostat.Thermostat.TargetTemperature.SetValue(portal.TargetTemperature())

			thermostat.Thermostat.CurrentHeatingCoolingState.SetValue(portal.CurrentHeatingCoolingState())
			thermostat.Thermostat.TargetHeatingCoolingState.SetValue(portal.TargetHeatingCoolingState())
		}
	}()

	return thermostat, nil
}
