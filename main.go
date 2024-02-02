package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

type ClientCredentialsResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type Statistics struct {
	PP         float32 `json:"pp"`
	GlobalRank int32   `json:"global_rank"`
}

type User struct {
	ID             int32  `json:"id"`
	Username       string `json:"username"`
	AvatarUrl      string `json:"avatar_url"`
	CountryCode    string `json:"country_code"`
	CountrySVG     string
	UserStatistics Statistics `json:"statistics"`
}

func getBearer(client *http.Client) string {
	client_id, _ := strconv.Atoi(os.Getenv("OSU_CLIENT_ID"))
	client_secret := os.Getenv("OSU_CLIENT_SECRET")

	data := strings.NewReader(fmt.Sprintf("client_id=%d&client_secret=%s&grant_type=client_credentials&scope=public", client_id, client_secret))
	req, err := http.NewRequest("POST", "https://osu.ppy.sh/oauth/token", data)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response ClientCredentialsResponse

	err = json.Unmarshal(bodyText, &response)
	if err != nil {
		log.Fatal(err)
	}

	return response.AccessToken

}

func test(w http.ResponseWriter, r *http.Request) {
	return_template := template.Must(template.ParseFiles("index.html"))

	user := []User{
		// {
		// 	ID:          4207965,
		// 	Username:    "mrekk",
		// 	AvatarUrl:   "https://a.ppy.sh/4207965",
		// 	CountryCode: "BR",
		// 	CountrySVG:  "1f1e7-1f1f7",
		// 	UserStatistics: Statistics{
		// 		PP:         6000,
		// 		GlobalRank: 12456,
		// 	},
		// },
	}

	return_template.Execute(w, user)
}

func convertCountryCodeToHexEmoji(country_code string) string {

	runeArray := []rune(country_code)
	var intArray []int32

	intArray = append(intArray, runeArray[0]+127397)
	intArray = append(intArray, runeArray[1]+127397)

	return fmt.Sprintf("%x-%x", intArray[0], intArray[1])

}

func deleteElement(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fmt.Println("oi!!!")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client := &http.Client{}
	bearer := getBearer(client)
	fmt.Println(bearer)

	getUser := func(w http.ResponseWriter, r *http.Request) {
		username := r.PostFormValue("username")

		req, err := http.NewRequest("GET", fmt.Sprintf("https://osu.ppy.sh/api/v2/users/%s?key=username", username), nil)

		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var response User
		err = json.Unmarshal(bodyText, &response)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(response)

		response.CountrySVG = convertCountryCodeToHexEmoji(response.CountryCode)

		htmlResponse := `
		<div
          class="flex flex-row text-left min-w-[475px] mx-4 my-2 border border-slate-500 border-1 p-3"
          id="card"
        >
          <img src="{{ .AvatarUrl }}" class="size-32 flex m-4 rounded-lg" />

          <div class="flex flex-grow flex-col">
            <div class="flex flex-row">
              <img src="https://osu.ppy.sh/assets/images/flags/{{ .CountrySVG }}.svg" class="size-8 flex self-center">
              <a
                href="https://osu.ppy.sh/u/{{ .ID }}"
                class="text-blue-500 underline grow"
                ><div class="text-2xl mx-4">{{ .Username }}</div></a
              >
              <div class="text-right text-red-500">
                <div
                  class="border border-1 border-red bg-red-300 hover:bg-red-200 active:bg-red-400 rounded px-1 pt-1.5 font-bold rounded-full"
                  hx-get="/delete-element/"
                  hx-trigger="click"
                  hx-target="closest #card"
                  hx-swap="outerHTML"
                >
                <span class="material-symbols-outlined">
                  delete
                  </span>
                </div>
              </div>
            </div>

            <div>PP: {{ .UserStatistics.PP }}</div>
            <div>Rank: #{{ .UserStatistics.GlobalRank }}</div>
          </div>
        </div>
		`

		t, err := template.New("foo").Parse(htmlResponse)
		t.Execute(w, response)
	}

	http.HandleFunc("/", test)
	http.HandleFunc("/add-user/", getUser)
	http.HandleFunc("/delete-element/", deleteElement)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
