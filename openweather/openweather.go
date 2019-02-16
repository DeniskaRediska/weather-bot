package openweather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const OpenWeatherAPI = "23c6fd16f52d2bbd5edc293bfab7681a"

type ForecastResponse struct {
	List []struct {
		Main struct {
			Temp float64
		}
		Dt_txt string
	}
	City struct {
		Name string
	}
}

func GetCurrentWeather(lat float64, lon float64) (string, float64, error) {
	client := http.Client{}
	weatherApi := fmt.Sprint("http://api.openweathermap.org/data/2.5/weather?lat=", lat, "&lon=", lon, "&units=metric&APPID=", OpenWeatherAPI)

	weatherResponse, err := client.Get(weatherApi)
	if err != nil {
		//log.Println(err)
		return "Не удалось считать данные", 0, err
	}
	defer weatherResponse.Body.Close()

	dateWeather, err := ioutil.ReadAll(weatherResponse.Body)
	if err != nil {
		return "", 0, err
	}

	type oWMResponse struct {
		Main struct {
			Temp float64
		}
		Name string
	}
	var weatherMessageGet oWMResponse

	err = json.Unmarshal(dateWeather, &weatherMessageGet)
	if err != nil {
		return "", 0, err
	}

	if weatherMessageGet.Name == "" {
		weatherMessageGet.Name = "Unknown"
	}

	return weatherMessageGet.Name, weatherMessageGet.Main.Temp, nil
}

func GetDailyForecast(lat float64, lon float64) (*ForecastResponse, error) {

	client := http.Client{}

	weatherApi := fmt.Sprint("http://api.openweathermap.org/data/2.5/forecast?lat=", lat, "&lon=", lon, "&units=metric&APPID=", OpenWeatherAPI)
	weatherResponse, err := client.Get(weatherApi)
	if err != nil {
		return nil, err
	}
	defer weatherResponse.Body.Close()

	dateWeather, err := ioutil.ReadAll(weatherResponse.Body)
	if err != nil {
		return nil, err
	}

	var weatherMessageGet ForecastResponse

	err = json.Unmarshal(dateWeather, &weatherMessageGet)
	if err != nil {
		return nil, err
	}

	if weatherMessageGet.City.Name == "" {
		weatherMessageGet.City.Name = "Unknown"
	}

	return &weatherMessageGet, nil
}
