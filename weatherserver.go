package main

import (
	"current-weather-server/data"
	"current-weather-server/logging"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const VERSION = "1.0.0"

var openWeatherApiKey string

// Mutex used when increment the request number which is used in logging
var requestNumberMutex sync.Mutex
var requestNumber uint64 = 0

func getNextRequestNumber() uint64 {
	requestNumberMutex.Lock()
	defer requestNumberMutex.Unlock()
	requestNumber++
	return requestNumber
}

// The templates used to serve files
var templates *template.Template

func init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error parsing template files: %v\n", r)
			os.Exit(1)
		}
	}()

	templates = template.Must(template.New("templateFiles").Parse(data.TEMPLATE_FILES))
}

func unixEpochTimeToString(epochTime int64) string {
	return time.Unix(epochTime, 0).UTC().String()
}

func versionHandler(requestNum uint64, writer http.ResponseWriter, request *http.Request) {
	templates.ExecuteTemplate(writer, "version_page", VERSION)
}

func getCurrentWeatherForm(requestNum uint64, writer http.ResponseWriter, request *http.Request) {
	templates.ExecuteTemplate(writer, "get_longitude_latitude", "")
}

func apiGetCurrentWeather(requestNum uint64, writer http.ResponseWriter, request *http.Request) {
	_, simplifiedData, err, statusCode := getCurrentWeather(request)

	if err != nil {
		if statusCode == 200 {
			// This should never happen, but in case it does, I'm logging and overriding it
			statusCode = http.StatusInternalServerError
			logging.LogError(requestNum, fmt.Sprintf("Overriding status code 200.  Setting to %v", statusCode))
		}

		logging.LogHTTPError(requestNum, err.Error(), statusCode)
		http.Error(writer, err.Error(), statusCode)
		return
	}

	jsonBytes, err := json.Marshal(simplifiedData)

	if err != nil {
		msg := fmt.Sprintf("Error marshing response: %v", err)
		logging.LogError(requestNum, msg)
		http.Error(writer, msg, statusCode)
		return
	}

	writer.Write(jsonBytes)
}

func getCurrentWeather(request *http.Request) (*data.CurrentWeatherData, *data.SimplifiedWeather, error, int) {
	queryValues := request.URL.Query()
	longitudeStr := queryValues.Get("longitude")
	latitudeStr := queryValues.Get("latitude")
	units := queryValues.Get("units")

	switch units {
	case "metric": // celsius, meters/sec
	case "imperial": // fahrenheit, miles/hour
	case "standard": // kelvin, meters/sec
	case "":
		units = "metric" // celsius, meters/sec
	default:
		return nil, nil, fmt.Errorf("Invalid units value: %v", units), http.StatusBadRequest
	}

	if longitudeStr == "" {
		return nil, nil, errors.New("missing longitude"), http.StatusBadRequest
	}

	if latitudeStr == "" {
		return nil, nil, errors.New("missing latitude"), http.StatusBadRequest
	}

	longitude, err := strconv.ParseFloat(longitudeStr, 64)

	if err != nil || longitude < -180 || longitude > 180 {
		return nil, nil, fmt.Errorf("Invalid longitude value: %v", longitudeStr), http.StatusBadRequest
	}

	latitude, err := strconv.ParseFloat(latitudeStr, 64)

	if err != nil || latitude < -90 || latitude > 90 {
		return nil, nil, fmt.Errorf("Invalid latitude value: %v", latitudeStr), http.StatusBadRequest
	}

	requestStr := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%v&lon=%v&appid=%v&units=%v",
		latitude, longitude, openWeatherApiKey, units)

	response, err := http.Get(requestStr)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return nil, nil, fmt.Errorf("Error calling Open Weather API: %v", err), http.StatusInternalServerError
	}

	// Should never happen, but just in case...
	if response == nil {
		return nil, nil, errors.New(fmt.Sprintf("Null response from Open Weather API call")), http.StatusInternalServerError
	}

	if response.StatusCode != 200 {
		return nil, nil,
			fmt.Errorf(fmt.Sprintf("Bad status code calling Open Weather API: %v (%v)",
				response.StatusCode, response.Status)),
			http.StatusInternalServerError
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, nil, fmt.Errorf("Error reading response body: %v", err), http.StatusInternalServerError
	}

	var currentWeatherDate data.CurrentWeatherData

	err = json.Unmarshal(body, &currentWeatherDate)

	if err != nil {
		return nil, nil, fmt.Errorf("Error unmarshalling json response body"), http.StatusInternalServerError
	}

	currentWeatherDate.Units = units
	currentWeatherDate.DataCollectionTime = unixEpochTimeToString(int64(currentWeatherDate.Dt))
	simplifiedData := data.SimplifyCurrentWeatherData(&currentWeatherDate)

	return &currentWeatherDate, simplifiedData, nil, http.StatusOK
}

func displayCurrentWeatherForm(requestNum uint64, writer http.ResponseWriter, request *http.Request) {

	_, simplifiedData, err, statusCode := getCurrentWeather(request)

	if err != nil {
		logging.LogHTTPError(requestNum, err.Error(), statusCode)
		templates.ExecuteTemplate(writer, "display_current_weather_error", err.Error())
		return
	}

	templates.ExecuteTemplate(writer, "display_current_weather", simplifiedData)
}

func logRequest(h func(requestNum uint64, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Client: %v, URL: %v", r.RemoteAddr, r.RequestURI)
		requestNum := getNextRequestNumber()
		logging.LogInfo(requestNum, msg)
		h(requestNum, w, r)
	}
}

func main() {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	var (
		version       = flag.Bool("version", false, "Print version and exit")
		maxProcessors = flag.Int("maxProcessors", 0, "Maximum number of processors to use (0=ALL)")
		logDir        = flag.String("logDir", ".", "Log directory")
		logFilePrefix = flag.String("logFilePrefix", "weatherserver", "The prefix for log files")
		port          = flag.String("port", "8000", "The port on which to run the server")
		apiKey        = flag.String("apiKey", "", "The key to use for API calls to Open Weather")
		//coldCoolWarmC = flag.String("coldCoolWarmC", "4.5,15.5,25", "Comma separated list of cold/cool/warm temperatures in Celsius")
		coldCoolWarmF = flag.String("coldCoolWarmF", "40,60,77", "Comma separated list of cold/cool/warm temperatures in Fahrenheit")
	)

	flag.Parse()

	if *version {
		fmt.Printf("Current Weather Server, version: %v\n", VERSION)
		return
	}

	if *logDir == "." {
		*logDir = currentWorkingDir
	}

	err = logging.Initialize(*logFilePrefix, *logDir)
	if err != nil {
		fmt.Printf("Error initializing logging:%v\n", err)
		os.Exit(1)
	}

	if *coldCoolWarmF == "" {
		logging.LogError(0, "No values specified for cold, cool, warm")
		os.Exit(1)
	}

	cold, cool, warm, err := data.ParseColdCoolWarmValues(*coldCoolWarmF)
	if err != nil {
		logging.LogError(0, err.Error())
		os.Exit(1)
	}

	err = data.SetColdCoolWarmCelsius(data.FahrenheitToCelsius(cold),
		data.FahrenheitToCelsius(cool),
		data.FahrenheitToCelsius(warm))

	if err != nil {
		logging.LogError(0, err.Error())
		os.Exit(1)
	}

	if *apiKey == "" {
		logging.LogError(0, "apiKey is request")
		os.Exit(1)
	}

	openWeatherApiKey = *apiKey

	if *maxProcessors == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
		logging.LogInfo(0, fmt.Sprintf("MAX_PROCS=%v", runtime.NumCPU()))
	} else {
		runtime.GOMAXPROCS(*maxProcessors)
		logging.LogInfo(0, fmt.Sprintf("MAX_PROCS=%v", *maxProcessors))
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", logRequest(versionHandler))
	mux.HandleFunc("/version", logRequest((versionHandler)))
	mux.HandleFunc("/getcurrentweather.html", logRequest((getCurrentWeatherForm)))
	mux.HandleFunc("/displaycurrentweather.html", logRequest((displayCurrentWeatherForm)))
	mux.HandleFunc("/api/currentweather", logRequest((apiGetCurrentWeather)))

	startMsg := fmt.Sprintf("Starting server on port %v", *port)
	logging.LogInfo(0, startMsg)
	fmt.Println(startMsg)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *port), mux)

	if err != nil {
		errMsg := fmt.Sprintf("Error starting server: %v", err)
		logging.LogFatal(0, errMsg)
		fmt.Println(errMsg)
		os.Exit(1)
	}
}
