package data

import (
	"fmt"
	"strconv"
	"strings"
)

// This map holds the cold, cool, warm temparatures for
// "C" (celsius), "F" (fahrenheit), and "K" (Kelvin)
var subjectiveTempMap map[string][]float64

// SetColdCoolWarmCelsius sets what temperatures will be
// used to determine subjective weather (hot, cold, etc.)
// Anything above warm is considered hot.
// The subjective weather is returned in SimplifiedWeather
// as subjectiveTemp
func SetColdCoolWarmCelsius(cold, cool, warm float64) error {
	if !(cold < cool && cool < warm) {
		return fmt.Errorf("Cold temp (%v) must be less than cool (%v).  Cool temp must be less than warm (%v).",
			cold, cool, warm)
	}

	subjectiveTempMap = map[string][]float64{}

	subjectiveTempMap["C"] = []float64{cold, cool, warm}
	subjectiveTempMap["F"] = []float64{CelsiusToFahrenheit(cold), CelsiusToFahrenheit(cool), CelsiusToFahrenheit(warm)}
	subjectiveTempMap["K"] = []float64{CelsiusToKelvin(cold), CelsiusToKelvin(cool), CelsiusToKelvin(warm)}
	return nil
}

func CelsiusToFahrenheit(c float64) float64 {
	return c*9.0/5.0 + 32
}

func CelsiusToKelvin(c float64) float64 {
	return c + 273.15
}

func FahrenheitToCelsius(f float64) float64 {
	return (f - 32) * 5.0 / 9.0
}

type CurrentWeatherData struct {
	// not part of the json return structure
	// added to the structure after the call to Open Weather
	Units              string
	DataCollectionTime string

	// These attributes are in the json return structure
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  float64 `json:"pressure"`
		Humidity  float64 `json:"humidity"`
		SeaLevel  float64 `json:"sea_level"`
		GrndLevel float64 `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Rain struct {
		H float64 `json:"1h"`
	} `json:"rain"`
	Clouds struct {
		All float64 `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		Id      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// SimplifiedWeather is the structure returned by
// calls to /api/currentweather
type SimplifiedWeather struct {
	Units              string  `json:"units"`
	DataCollectionTime string  `json:"dataCollectionTime"`
	Lat                float64 `json:"latitude"`
	Long               float64 `json:"longitude"`
	CloudinessPercent  float64 `json:"cloudinessPercent"`
	HumidityPercent    float64 `json:"humidityPercent"`
	Temp               float64 `json:"temp"`
	TempHigh           float64 `json:"tempHigh"`
	TempLow            float64 `json:"tempLow"`
	TempFeelsLike      float64 `json:"tempFeelsLike"`
	ExpectedWeather    string  `json:"expectedWeather"`
	WeatherDescription string  `json:"weatherDescription"`
	SubjectiveTemp     string  `json:"subjectiveTemp"`
	Summary            string  `json:"summary"`
}

// SimplifyCurrentWeatherData generates a SimplifiedWeather object from
// the CurrentWeatherData (returned by Open Weather) and the
// cold, cool, warm temperature set when the application started.
func SimplifyCurrentWeatherData(data *CurrentWeatherData) *SimplifiedWeather {
	if data == nil || data.Weather == nil || len(data.Weather) == 0 {
		return nil
	}

	simplified := &SimplifiedWeather{}
	simplified.DataCollectionTime = data.DataCollectionTime
	simplified.Lat = data.Coord.Lat
	simplified.Long = data.Coord.Lon
	simplified.CloudinessPercent = data.Clouds.All
	simplified.Temp = data.Main.Temp
	simplified.TempHigh = data.Main.TempMax
	simplified.TempLow = data.Main.TempMin
	simplified.TempFeelsLike = data.Main.FeelsLike
	simplified.HumidityPercent = data.Main.Humidity
	mainDesc := make([]string, len(data.Weather))
	mainSubDesc := make([]string, len(data.Weather))
	for inx, weather := range data.Weather {
		mainDesc[inx] = weather.Main
		mainSubDesc[inx] = weather.Description
	}

	simplified.ExpectedWeather = strings.ToLower(strings.Join(mainDesc, ","))
	simplified.WeatherDescription = strings.ToLower(strings.Join(mainSubDesc, ","))

	switch data.Units {
	case "standard":
		simplified.Units = "K" // kelvin
	case "metric":
		simplified.Units = "C" // celsius
	case "imperial":
		simplified.Units = "F" // fahrenheit
	}

	m := subjectiveTempMap[simplified.Units]
	cold, cool, warm := m[0], m[1], m[2]

	if simplified.Temp <= cold {
		simplified.SubjectiveTemp = "cold"
	} else if simplified.Temp <= cool {
		simplified.SubjectiveTemp = "cool"
	} else if simplified.Temp <= warm {
		simplified.SubjectiveTemp = "warm"
	} else {
		simplified.SubjectiveTemp = "hot"
	}

	simplified.Summary = fmt.Sprintf("The weather will be %v.  Expect %v with a high of %v °%v, a low of %v °%v, "+
		"and an average temperature of %v °%v.  It'll fell like %v °%v with a humidity of %v%% and a cloud cover of %v%%.",
		simplified.SubjectiveTemp, simplified.ExpectedWeather,
		simplified.TempHigh, simplified.Units,
		simplified.TempLow, simplified.Units,
		simplified.Temp, simplified.Units,
		simplified.TempFeelsLike, simplified.Units,
		simplified.HumidityPercent, simplified.CloudinessPercent)

	return simplified
}

func ParseColdCoolWarmValues(str string) (float64, float64, float64, error) {
	parts := strings.Split(str, ",")

	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("Invalid number of temperatures for coldCoolWarmF: %v", len(parts))
	}

	cold, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("Error parsing cold value in coldCoolWarmF: %v", err)
	}

	cool, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("Error parsing cool value in coldCoolWarmF: %v", err)
	}

	warm, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("Error parsing warm value in coldCoolWarmF: %v", err)
	}

	return cold, cool, warm, nil
}
