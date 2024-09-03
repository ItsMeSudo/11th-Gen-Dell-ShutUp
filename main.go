package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	IPMIHOST     = getEnv("IPMIHOST", "192.168.1.100")
	IPMIUSER     = getEnv("IPMIUSER", "admin")
	IPMIPW       = getEnv("IPMIPW", "password")
	SENSOR       = getEnv("SENSOR", "Ambient Temp")
	MAXTEMP      = getEnvInt("MAXTEMP", 30)
	WARNTEMP     = getEnvInt("WARNTEMP", 27)
	POLLINTERVAL = getEnvInt("POLLINTERVAL", 60) // Polling interval in seconds
	FANSPEEDS    = map[int]string{
		1920: "0x07",
		1800: "0x06",
		1680: "0x05",
		1560: "0x04", // Example: 1560 RPM
	}
	LOGFILE = getEnv("LOGFILE", "/var/log/ipmi-temp-monitor.log")
)

func main() {
	setupLogging()
	defer recoverFromPanic()

	handleShutdown()

	for {
		temp, err := getTemperature()
		if err != nil {
			log.Printf("Error getting temperature: %v", err)
			time.Sleep(time.Duration(POLLINTERVAL) * time.Second)
			continue
		}

		log.Printf("Current temperature (%s): %d°C", SENSOR, temp)

		if temp > MAXTEMP {
			log.Printf("Temperature is BAD (%d°C). Setting fans to auto.", temp)
			runIPMICommandWithRetry([]string{"raw", "0x30", "0x30", "0x01", "0x01"}, 3, 2*time.Second)
		} else if temp > WARNTEMP {
			log.Printf("Temperature is WARN (%d°C). Setting fans to manual mode.", temp)
			runIPMICommandWithRetry([]string{"raw", "0x30", "0x30", "0x01", "0x00"}, 3, 2*time.Second)
			runIPMICommandWithRetry([]string{"raw", "0x30", "0x30", "0x02", "0xff", FANSPEEDS[1920]}, 3, 2*time.Second)
		} else {
			log.Printf("Temperature is OK (%d°C). Setting fans to lower speed.", temp)
			runIPMICommandWithRetry([]string{"raw", "0x30", "0x30", "0x01", "0x00"}, 3, 2*time.Second)
			runIPMICommandWithRetry([]string{"raw", "0x30", "0x30", "0x02", "0xff", FANSPEEDS[1560]}, 3, 2*time.Second)
		}

		time.Sleep(time.Duration(POLLINTERVAL) * time.Second)
	}
}

func setupLogging() {
	logFile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// recovers from panic and restarts the main loop.
func recoverFromPanic() {
	if r := recover(); r != nil {
		log.Printf("Program crashed with error: %v. Restarting...", r)
		main() // Restart main loop after crash
	}
}

// gracefully shuts down the program on interrupt signals.
func handleShutdown() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopChan
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()
}

// executes an ipmitool
func runIPMICommand(args []string) error {
	cmd := exec.Command("ipmitool", append([]string{"-I", "lanplus", "-H", IPMIHOST, "-U", IPMIUSER, "-P", IPMIPW}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		if strings.Contains(out.String(), "rsp=0xcc") && strings.Contains(out.String(), "Invalid data field in request") {
			log.Printf("IPMI command %v executed with a minor error (rsp=0xcc), continuing. Trust me or not IDC", args)
			return nil
		}
		log.Printf("Error running IPMI command %v: %v. Output: %s", args, err, out.String())
		return fmt.Errorf("critical error running IPMI command %v: %v", args, err)
	}
	log.Printf("IPMI command %v successful. Output: %s", args, out.String())
	return nil
}

// retries an IPMI command on failure with exponential backoff.
func runIPMICommandWithRetry(args []string, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		err := runIPMICommand(args)
		if err == nil {
			return nil
		}
		log.Printf("Attempt %d: Error running IPMI command %v: %v", i+1, args, err)
		time.Sleep(delay)
		delay *= 2 // Exponential backoff
	}
	return fmt.Errorf("failed to run IPMI command %v after %d retries", args, retries)
}

// fetches and parses the temperature from the IPMI sensor.
func getTemperature() (int, error) {
	cmd := exec.Command("ipmitool", "-I", "lanplus", "-H", IPMIHOST, "-U", IPMIUSER, "-P", IPMIPW, "sdr", "get", SENSOR)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error running IPMI temperature command: %v", err)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Sensor Reading") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				temp, err := strconv.Atoi(parts[3])
				if err != nil {
					return 0, fmt.Errorf("error parsing temperature: %v", err)
				}
				return temp, nil
			}
		}
	}
	return 0, fmt.Errorf("temperature reading not found in output: %s", output)
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
