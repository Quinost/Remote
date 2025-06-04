# Remote Chrome Controller

Web-based remote control for Chrome browser with real-time screenshot streaming and system controls. Built with Go backend and Angular frontend.

## Features
- Real-time screenshot streaming via WebSocket
- Click-to-interact with browser content
- Volume controls and system shutdown
- Cross-platform support (Linux/Windows)

## Setup
Chrome profile is stored in `ChromeProfile/Profile` directory relative to the executable. The profile folder must be named exactly "Profile" to work properly.

## Usage
1. Run the Go backend - it will start Chrome and a web server on port 8080
2. Open the web interface in any browser to control the Chrome instance
3. Use scroll buttons, click on screenshots, or system controls like volume/shutdown

Server will display all available network interfaces on startup for easy access from other devices.