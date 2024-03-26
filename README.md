# Current Weather Server (Using OpenWeather API)
A simple server that provides one API and a select few web pages for getting a summary of the current weather at a specific longitude and latitude.

### API 
Only one API is provided:

```script
api/currentweather
```

It should be issued as a GET command and takes the following options:

```script
longitude:  A floating point value between -180 and 180 (inclusive).  REQUIRED.
latitude: A floating point value between -90 and 90 (inclusive).  REQUIRED.
units: imperial, metric, or standard.  OPTIONAL.

"imperial" with return values in Fahrenheit.
"metric" will return values in Celsius.
"standard" will return values in Kelvin.
The default for "units" is "metric"
```
### Example API usage
curl http://localhost:8000/api/currentweather\?longitude=80\&latitude=30\&units=imperial 

```json
  {
    "units": "F",
    "dataCollectionTime": "2024-03-26 19:44:08 +0000 UTC",
    "latitude": 30,
    "longitude": 80,
    "cloudinessPercent": 97,
    "humidityPercent": 40,
    "temp": 51.76,
    "tempHigh": 51.76,
    "tempLow": 51.76,
    "tempFeelsLike": 48.52,
    "expectedWeather": "clouds",
    "weatherDescription": "overcast clouds",
    "subjectiveTemp": "cool",
    "summary": "The weather will be cool.  Expect clouds with a high of 51.76 \u00b0F, a low of 51.76 \u00b0F, and an average temperature of 51.76 \u00b0F.  It'll fell like 48.52 \u00b0F with a humidity of 40% and a cloud cover of 97%."
 }
```

### To run
You'll need to acquire an API key from https://openweathermap.org/.  
This API key must be passed to the server when it's started like this:

```script
./weatherserver -apiKey=XXXXXXXXXXXX
```

### Other available options

```shell
  -logFilePrefix string
        The prefix for log files (default "weatherserver")
  -apiKey string
        The key to use for API calls to Open Weather
  -coldCoolWarmF string
        Comma separated list of cold/cool/warm temperatures in Fahrenheit (default "40,60,77")
  -logDir string
        Log directory (default ".")
  -maxProcessors int
        Maximum number of processors to use (0=ALL)
  -port string
        The port on which to run the server (default "8000")
  -version
        Print version and exit
```


### Available web pages

```shell
http://localhost:8000/
http://localhost:8000/version.html
http://localhost:8000/getcurrentweather.html
http://localhost:8000/displaycurrentweather.html (used by getcurrentweather.html to display the results)
```

