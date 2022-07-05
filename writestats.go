package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// "strings"

	// "strconv"

	//"reflect"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"google.golang.org/api/script/v1"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := "1UzB2IIzqNqzs9sWWV65w0VVHUmUaeFH1eGlK4-jyNMc"

	//Read Ordinal Data and store in Map
	type Car struct {
		Manufacturer string
		Model        string
		Year         string
		Country      string
		Designation  string
		CarType      string
		Drivetrain   string
		Setup        string
		Engine       string
		Aspiration   string
		Litreage     string
		Value        string
	}

	ordinalMap := make(map[string]Car)
	ordinalNumber, err := getOrdinalNumber(("log.csv")) // Get ordinal number of current car
	if err != nil {
		log.Fatalf("Unable to retrieve Ordinal Number. CSV file is likely empty.")
	}
	var currentCar Car

	readRange := "Ordinal Data"
	// Read all up-to-date data from the Ordinal Data sheet
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve ordinal data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		fmt.Println("No ordinal data found.")
	} else {
		// Store all information in a map where the Ordinal Number is the key and
		// the value is a struct "Car", which stores all info about the car
		for _, row := range resp.Values {
			ordinalMap[fmt.Sprintf("%v", row[0])] = Car{
				Manufacturer: fmt.Sprintf("%v", row[1]),
				Model:        fmt.Sprintf("%v", row[2]),
				Year:         fmt.Sprintf("%v", row[3]),
				Country:      fmt.Sprintf("%v", row[4]),
				Designation:  fmt.Sprintf("%v", row[5]),
				CarType:      fmt.Sprintf("%v", row[6]),
				Drivetrain:   fmt.Sprintf("%v", row[7]),
				Setup:        fmt.Sprintf("%v", row[8]),
				Engine:       fmt.Sprintf("%v", row[9]),
				Aspiration:   fmt.Sprintf("%v", row[10]),
				Litreage:     fmt.Sprintf("%v", row[11]),
				Value:        fmt.Sprintf("%v", row[12]),
			}
		}

		if val, isPresent := ordinalMap[ordinalNumber]; isPresent { // Check if the ordinal Number is in the map
			currentCar = val
		} else { // If the Ordinal Number is not in the map then the car likely hasn't been added to the sheet
			log.Fatalf("Current Car has not been added to Ordinal Data sheet!\n Please add the car's info and run the program again.\n")
		}
	}

	//Write Data to Output Sheet
	writeRange := "Stat Builder!M8" // or "sheet1:A1" if you have a different sheet
	statValues := calcstats("log.csv")
	carFullName := currentCar.Manufacturer + " " + currentCar.Model
	writeValues := []interface{}{}
	writeValues = append(writeValues, // Builds Stat Line to leaderboard specifications
		carFullName,            // Car Name
		"",                     // Best Lap Time (not handled)
		currentCar.Year,        // Year
		currentCar.Country,     // Country
		"",                     // Country Flag (not handled)
		statValues[0],          // PI Class Number
		currentCar.Designation, // Car Designation
		currentCar.CarType,     // Car Type (Category)
		statValues[1],          // Drivetrain (from actual stats, not default data)
		currentCar.Setup,       // Engine Setup
		currentCar.Litreage,    // Engine Litreage
		currentCar.Engine,      // Engine
		currentCar.Aspiration,  // Aspiration
		statValues[11],         // Peak Boost
		statValues[2],          // Peak Horsepower
		statValues[3],          // Peak Torque
		"",                     // Weight (not handled)
		"",                     // Power to Weight (not handled)
		statValues[4],          // 0-60 Time
		statValues[5],          // 0-100 Time
		statValues[6],          // 60-150 Time
		statValues[7],          // 100-200 Time
		statValues[10],         // Top Speed
		"",                     // Track Top Speed (not handled)
		statValues[8],          // 60-0 Time
		statValues[9],          // 100-0 Time
		"",                     // Lateral Gs at 60mph (not handled)
		"",                     // Lateral Gs at 120mpg (not handled)
		currentCar.Value)       // Car Value

	var vr sheets.ValueRange
	vr.Values = append(vr.Values, writeValues)

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("USER-ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to print data to sheet. %v", err)
	}
	fmt.Println("Successfully printed data to output sheet!")

	// Trigger Apps Script to set data colors
	service, err := script.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Script client: %v", err)
	}

	ScriptId := "1cwTGL840G2QJZNmwiSWK-VUlLq8Wjze6osbEqxDyXBVULAtQQJRURL6k"

	req := script.ExecutionRequest{Function: "remoteSetDataColors", DevMode: true}

	fmt.Println("Triggering color script...")
	_, err = service.Scripts.Run(ScriptId, &req).Do()
	if err != nil {
		log.Fatalf("Unable to trigger script. %v", err)
	}
	fmt.Println("Script successfully set data colors!")
}
