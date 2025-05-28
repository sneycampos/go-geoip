# 🌍 go-geoip — Free GeoIP API Service

A simple, fast, open-source and free IP geolocation lookup service powered by MaxMind's GeoLite2 database.

## 🔍 Features

- ✅ Free
- ✅ No authentication required
- ✅ No rate limits
- ✅ No API key needed
- ✅ Open-source
- ✅ JSON response format
- ✅ Supports IPv4 addresses
- ✅ Provides country, city, coordinates, and timezone data
- ✅ Fast and reliable

## 🔗 How to Use

Just send a GET request to: https://geoip.sney.eu/ip/{IP_ADDRESS}

### Example Request:
```bash
curl https://geoip.sney.eu/ip/172.68.100.190
```

### Example Response:
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

### 💻 JavaScript Example 
```javascript
fetch('https://geoip.sney.eu/ip/172.68.100.190 ')
  .then(response => response.json())
  .then(data => console.log(data));
```
