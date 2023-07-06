// Very, very quick-and-dirty® program to control fan on Orange Pi.
// The OPi tends to run quite coolly, so I want the fan to only switch on when
// we get spikes of CPU temperature.
// Everything is hard-coded at the moment because ...... I can't be bothered.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// TODO - make these command-line / env.
var (
	// The temperature sensors available.
	temperatureSensors = []string{"/sys/devices/virtual/thermal/thermal_zone0/temp", "/sys/devices/virtual/thermal/thermal_zone1/temp"}
	// Temperature will be an integer, scaled by 1000.
	scale = 1000
	// Temperature at which we will turn on the fan.
	onTemp = 65
	// Temperature below which we will turn off the fan.
	offTemp = 45
	// How long the temperature has to stay above onTemp before we turn on.
	onDelay = 4000 * time.Millisecond
	// How long the temperature has to stay below offTemp before we turn off.
	offDelay = 30 * time.Second
)

// Read the temperature from the given sensor.
func readSensor(name string) (int, error) {
	f, err := os.Open(name)
	if err != nil {
		return 0, err
	}

	var t int
	_, err = fmt.Fscan(f, &t)
	if err != nil {
		return 0, err
	}

	return t / scale, nil
}

// Return the board temperature in Celsius. The hottest sensor value is returned.
func temperature() int {
	max := 0
	for _, sensor := range temperatureSensors {
		t, err := readSensor(sensor)
		if err != nil {
			// TODO - what can we do, except ignore it?
			continue
		}

		if t > max {
			max = t
		}
	}

	return max
}

// Run the gpio (wiringpi) command (which does *not* require root access).
func gpio(args ...string) error {
	cmd := exec.Command("gpio", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return cmd.Wait()
}

func gpioSetup() {
	gpio("mode", "3", "out")
}

func turnOff() {
	fmt.Println(time.Now().String()[:19], "Off", temperature())
	gpio("write", "3", "0")
}

func turnOn() {
	fmt.Println(time.Now().String()[:19], "On ", temperature())
	gpio("write", "3", "1")
}

func main() {
	gpioSetup()

	for {
		turnOff()
		lastCoolTimestamp := time.Now()
		for {
			// The hardware CPU temp is refreshed approx twice per second.
			time.Sleep(550 * time.Millisecond)
			if temperature() < onTemp {
				lastCoolTimestamp = time.Now()
			} else if lastCoolTimestamp.Add(onDelay).Before(time.Now()) {
				break
			}
		}

		turnOn()
		lastHotTimestamp := time.Now()
		for {
			// As a reminder: the workload causes higher CPU temperature with occasional
			// dips down to normal. Make sure the temp has *stayed* low for 30 seconds.
			time.Sleep(time.Second)
			if temperature() >= offTemp {
				lastHotTimestamp = time.Now()
			} else if lastHotTimestamp.Add(offDelay).Before(time.Now()) {
				break
			}
		}
	}
}
