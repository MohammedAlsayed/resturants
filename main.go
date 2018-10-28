package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type PlaceSearch struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Results          []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		Icon         string `json:"icon"`
		ID           string `json:"id"`
		Name         string `json:"name"`
		OpeningHours struct {
			OpenNow bool `json:"open_now"`
		} `json:"opening_hours,omitempty"`
		Photos []struct {
			Height           int           `json:"height"`
			HTMLAttributions []interface{} `json:"shtml_attributions"`
			PhotoReference   string        `json:"photo_reference"`
			Width            int           `json:"width"`
		} `json:"photos"`
		PlaceID string `json:"place_id"`
		Scope   string `json:"scope"`
		AltIds  []struct {
			PlaceID string `json:"place_id"`
			Scope   string `json:"scope"`
		} `json:"alt_ids,omitempty"`
		Reference string   `json:"reference"`
		Types     []string `json:"types"`
		Vicinity  string   `json:"vicinity"`
	} `json:"results"`
	Status string `json:"status"`
}

type PlaceDetails struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Result           struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		AdrAddress           string `json:"adr_address"`
		FormattedAddress     string `json:"formatted_address"`
		FormattedPhoneNumber string `json:"formatted_phone_number"`
		Geometry             struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		Icon                     string  `json:"icon"`
		ID                       string  `json:"id"`
		InternationalPhoneNumber string  `json:"international_phone_number"`
		Name                     string  `json:"name"`
		PlaceID                  string  `json:"place_id"`
		Rating                   float64 `json:"rating"`
		Reference                string  `json:"reference"`
		Reviews                  []struct {
			AuthorName              string `json:"author_name"`
			AuthorURL               string `json:"author_url"`
			Language                string `json:"language"`
			ProfilePhotoURL         string `json:"profile_photo_url"`
			Rating                  int    `json:"rating"`
			RelativeTimeDescription string `json:"relative_time_description"`
			Text                    string `json:"text"`
			Time                    int    `json:"time"`
		} `json:"reviews"`
		Scope     string   `json:"scope"`
		Types     []string `json:"types"`
		URL       string   `json:"url"`
		UtcOffset int      `json:"utc_offset"`
		Vicinity  string   `json:"vicinity"`
		Website   string   `json:"website"`
	} `json:"result"`
	Status string `json:"status"`
}

type Place struct {
	Name       string
	Rating     float64
	NumReviews string
	Latitude   float64
	Longitude  float64
}

func main() {
	key := flag.String("key", "", "your google maps api key")
	location := flag.String("location", "24.796074,46.669509", "the central location point you want to look around it")
	radius := flag.String("radius", "3000", "the radius you want to search around the central location point in meters")
	name := flag.String("name", "", "the place keywork you are looking for")

	if *key == "" {
		log.Fatal("No key provided")
	}
	if *name == "" {
		log.Fatal("No place name provided")
	}

	placeSearch, err := findPlaces(*name, *key, *location, *radius)
	if err != nil {
		log.Fatal(err.Error())
	}
	if len(placeSearch.Results) == 0 {
		fmt.Println("no results")
		os.Exit(0)
	}

	file, err := os.OpenFile("data.csv", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal("faild opening data.csv file")
	}
	defer file.Close()
	w := csv.NewWriter(file)

	results := placeSearch.Results
	for _, result := range results {
		placeDetails, err := placeDetails(result.PlaceID, *key)
		if err != nil {
			fmt.Errorf(err.Error())
		}
		resturant := Place{Name: result.Name,
			Rating:     placeDetails.Result.Rating,
			NumReviews: "1",
			Latitude:   result.Geometry.Location.Lat,
			Longitude:  result.Geometry.Location.Lng,
		}
		if err := w.Write(resturant.ToString()); err != nil {
			fmt.Errorf("Error while writing to file (err=%s)", err.Error())
		}
		w.Flush()
	}
}

func findPlaces(keyword string, key string, location string, radius string) (*PlaceSearch, error) {
	var api *url.URL
	api, err := url.Parse("https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=" + location + "&radius=" + radius + "&type=restaurant&keyword=" + keyword + "&key=" + key)
	if err != nil {
		return nil, err
	}
	res, err := http.Get(api.String())
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var place PlaceSearch
	err = json.Unmarshal(body, &place)
	if err != nil {
		return nil, err
	}
	return &place, nil
}

func placeDetails(placeId string, key string) (*PlaceDetails, error) {
	var api *url.URL
	api, err := url.Parse("https://maps.googleapis.com/maps/api/place/details/json?placeid=" + placeId + "&key=" + key)
	if err != nil {
		return nil, err
	}
	res, err := http.Get(api.String())
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var placeDetails PlaceDetails
	err = json.Unmarshal(body, &placeDetails)
	if err != nil {
		return nil, err
	}
	return &placeDetails, nil
}

func (this *Place) ToString() []string {
	array := []string{this.Name, this.NumReviews, strconv.FormatFloat(this.Rating, 'f', 6, 64), strconv.FormatFloat(this.Latitude, 'f', 6, 64), strconv.FormatFloat(this.Longitude, 'f', 6, 64)}
	return array
}
