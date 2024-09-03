# IPMI Temperature-Based Fan Control

## Description

**Fan control for 11th Gen Dell servers**: This program dynamically controls the fan speed based on temperature readings from IPMI sensors. It helps manage fan noise, especially since Dell servers are notoriously loud by default.

## Features

- Automatically adjusts fan speeds based on ambient or CPU temperature.
- Logs temperature readings and actions to both a log file and the console.
- Gracefully handles errors and retries IPMI commands if necessary.
- Configurable through environment variables for easy customization.
- Recovers from crashes and restarts itself to maintain continuous monitoring.
- Supports user-defined polling intervals and temperature thresholds for fan speed adjustments.

## Prerequisites

- Go 1.22 or later installed on your system.
- IPMI tool installed and accessible from your server.

## Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/ItsMeSudo/11th-Gen-Dell-ShutUp.git
   cd 11th-Gen-Dell-ShutUp
   ```
2. **Build the Go program:**
   ```bash
   go build
    ```
3.**Set environment variables:** You can set the necessary environment variables in your shell or use a .env file or just in the code.
```bash
   export IPMIHOST="192.168.1.100"
   export IPMIUSER="admin"
   export IPMIPW="password"
   export SENSOR="Ambient Temp"
   export MAXTEMP=30
   export WARNTEMP=27
   export POLLINTERVAL=60
   export LOGFILE="/var/log/ipmi-temp-monitor.log"
```
4. **Run the program: (systemd config also provided)**
   ```bash
   ./11thgen-shutup
   ```
   
## Environment Variables
- IPMIHOST: IP address of the IPMI interface (default: 192.168.1.100).
- IPMIUSER: Username for IPMI login (default: admin).
- IPMIPW: Password for IPMI login (default: password).
- SENSOR: Sensor to monitor (default: "Ambient Temp").
- MAXTEMP: Maximum temperature threshold to set fans to auto mode (default: 30).
- WARNTEMP: Warning temperature threshold to set fans to manual high speed (default: 27).
- POLLINTERVAL: Time in seconds between each temperature check (default: 60).
- LOGFILE: Path to the log file (default: /var/log/ipmi-temp-monitor.log).

## Contributing
Contributions are welcome! If you'd like to fork this repository and modify the code, feel free to do so. However, please note that this is my original work. If you modify or build upon this code, kindly do not claim it as your own. Acknowledge the original author and provide a link back to this repository.

## License
This project is licensed under the MIT License - see the LICENSE file for details. While you're free to use and modify this code, please give credit to the original author and do not redistribute modified versions as your own work.
Feel free to adjust any specific details to better match your project's requirements or personal preferences.

