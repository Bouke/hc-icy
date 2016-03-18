package icy

import (
	"log"
	"time"

	"github.com/brutella/hc/model"
	"github.com/brutella/hc/model/accessory"
)

func (portal Portal) HeatCoolMode() model.HeatCoolModeType {
	switch portal.Mode() {
	case Fixed:
		return model.HeatCoolModeOff
	case Away:
		return model.HeatCoolModeCool
	case Saving:
		return model.HeatCoolModeCool
	case Comfort:
		return model.HeatCoolModeHeat
	default:
		panic("Unexpected mode")
	}
}

func (portal Portal) TargetHeatCoolMode() model.HeatCoolModeType {
	switch portal.Mode() {
	case Fixed:
		return model.HeatCoolModeOff
	case Away:
		return model.HeatCoolModeAuto
	case Saving:
		return model.HeatCoolModeAuto
	case Comfort:
		return model.HeatCoolModeAuto
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

	thermostat := accessory.NewThermostat(model.Info{
		Name:         "e-thermostaat",
		Manufacturer: "ICY",
	}, portal.TargetTemperature(), 0, 30, 0.5)
	thermostat.SetTemperature(portal.Temperature())
	thermostat.SetMode(portal.HeatCoolMode())
	thermostat.SetTargetMode(portal.TargetHeatCoolMode())

	thermostat.OnTargetTempChange(func(value float64) {
		portal.SetTargetTemperature(value)
		portal.Write()
	})
	thermostat.OnTargetModeChange(func(value model.HeatCoolModeType) {
		switch value {
		case model.HeatCoolModeOff:
			portal.SetMode(Fixed)
			temperature := float64(portal.Status.Configuration[5]) / 2
			portal.SetTargetTemperature(temperature)
			thermostat.SetTargetTemperature(temperature)
		case model.HeatCoolModeCool:
			portal.SetMode(Saving)
			temperature := float64(portal.Status.Configuration[5]) / 2
			portal.SetTargetTemperature(temperature)
			thermostat.SetTargetTemperature(temperature)
		case model.HeatCoolModeHeat:
			fallthrough
		case model.HeatCoolModeAuto:
			portal.SetMode(Comfort)
			temperature := float64(portal.Status.Configuration[6]) / 2
			portal.SetTargetTemperature(temperature)
			thermostat.SetTargetTemperature(temperature)
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
			thermostat.SetTemperature(portal.Temperature())
			thermostat.SetTargetTemperature(portal.TargetTemperature())

			thermostat.SetMode(portal.HeatCoolMode())
			thermostat.SetTargetMode(portal.TargetHeatCoolMode())
		}
	}()

	return thermostat, nil
}
