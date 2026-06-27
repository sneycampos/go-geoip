# 🌍 GeoIP API Service

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/docker-ready-brightgreen.svg)](https://www.docker.com/)

A **lightweight**, **fast**, and **free** IP geolocation lookup API service built with Go and powered by [MaxMind's GeoLite2 database](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data).

## ✨ Features

- 🚀 **High Performance** - Built with Go for optimal speed and low memory footprint
- 📍 **Detailed Geolocation** - Returns country, city, coordinates, postal code, and timezone
- 🛑 **Graceful Shutdown** - Finishes active requests before stopping

## 🚀 Quick Start

**Prerequisites:**
- Docker and Docker Compose installed on your machine
- Download the latest free version of the [GeoLite2 City database](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) from MaxMind. You will need to create a free account to access the download link.
- Extract the downloaded file and place the `GeoLite2-City.mmdb` file in the project root.

- **Start the service**
```bash
docker compose up -d
```

**Health Check:**
```bash
curl -i http://localhost:8888/healthz
```

**Example Request:**
```bash
curl http://localhost:8888/172.68.100.190
```

**Example Response:**
```json
{
  "ip": "172.68.100.190",
  "country": "Portugal",
  "country_code": "PT",
  "city": "Lisbon",
  "postal_code": "1000-005",
  "latitude": 38.7219,
  "longitude": -9.1398,
  "timezone": "Europe/Lisbon"
}
```

## 🏗️ Architecture

- **Web Framework**: Go standard library (`net/http`)
- **Database**: MaxMind GeoLite2-City (MMDB format)
- **Database Reader**: [maxminddb-golang](https://github.com/oschwald/maxminddb-golang)

## 📝 License

This project is licensed under the MIT License.
