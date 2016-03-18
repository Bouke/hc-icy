HomeControl - ICY Thermostat
============================

For usage with the ICY / Essent e-thermostaat. Example usage:

    package main

    import (
        "log"

        "github.com/bouke/hc-icy"
        "github.com/brutella/hc/hap"
    )

    func main() {
        thermostat, err := icy.NewThermostat("username", "password")
        if err != nil {
            log.Fatal(err)
        }

        t, err := hap.NewIPTransport(hap.Config{Pin: "00102003"}, thermostat.Accessory)
        if err != nil {
            log.Fatal(err)
        }

        hap.OnTermination(func() {
            t.Stop()
        })

        t.Start()
    }
