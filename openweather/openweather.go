package openweather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const OpenWeatherAPI = "23c6fd16f52d2bbd5edc293bfab7681a"

type scheduleStruct struct {
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

func WeatherAnswer(lat float64, lon float64) (string, float64, error) {
	client := http.Client{}
	WeatherApi := fmt.Sprint("http://api.openweathermap.org/data/2.5/weather?lat=", lat, "&lon=", lon, "&units=metric&APPID=", OpenWeatherAPI)

	WeatherResponse, err := client.Get(WeatherApi)
	if err != nil {
		//log.Println(err)
		return "Не удалось считать данные", 0, err
	}

	DateWeather, err := ioutil.ReadAll(WeatherResponse.Body)
	if err != nil {
		log.Println(err)
	}

	type OWMResponse struct {
		Main struct {
			Temp float64
		}
		Name string
	}
	var WeatherMessageGet OWMResponse

	err = json.Unmarshal(DateWeather, &WeatherMessageGet)
	if err != nil {
		log.Println(err)
	}

	if WeatherMessageGet.Name == "" {
		WeatherMessageGet.Name = "Unknown"
	}

	return WeatherMessageGet.Name, WeatherMessageGet.Main.Temp, nil
}

func WeatherScheule(lat float64, lon float64) (scheduleStruct, error) {

	client := http.Client{}

	WeatherApi := fmt.Sprint("http://api.openweathermap.org/data/2.5/forecast?lat=", lat, "&lon=", lon, "&units=metric&APPID=", OpenWeatherAPI)
	WeatherResponse, err := client.Get(WeatherApi)
	if err != nil {
		//log.Println(err)
		return scheduleStruct{}, err
	}

	DateWeather, err := ioutil.ReadAll(WeatherResponse.Body)
	if err != nil {
		log.Println(err)
	}

	var WeatherMessageGet scheduleStruct

	err = json.Unmarshal(DateWeather, &WeatherMessageGet)
	if err != nil {
		log.Println(err)
	}

	if WeatherMessageGet.City.Name == "" {
		WeatherMessageGet.City.Name = "Unknown"
	}

	return WeatherMessageGet, nil
}
