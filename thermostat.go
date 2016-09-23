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
		return characteristic.TargetHeatingCoolingStateOff
	case Saving:
		return characteristic.TargetHeatingCoolingStateOff
	case Comfort:
		return characteristic.TargetHeatingCoolingStateAuto
	default:
		panic("Unexpected mode")
	}
}

func NewThermostat(username string, password string) (*accessory.Thermostat, error) {
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
		Name:         "e-thermostaat",
		Manufacturer: "ICY",
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
		switch value {
		case characteristic.TargetHeatingCoolingStateOff:
			portal.SetMode(Fixed)
		case characteristic.TargetHeatingCoolingStateCool:
			portal.SetMode(Saving)
		case characteristic.TargetHeatingCoolingStateHeat:
			fallthrough
		case characteristic.TargetHeatingCoolingStateAuto:
			portal.SetMode(Comfort)
		}
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
