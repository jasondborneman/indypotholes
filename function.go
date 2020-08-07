package indypotholes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dghubble/oauth1"
	"github.com/jasondborneman/go-twitter/twitter"
	"github.com/wcharczuk/go-chart"
	"google.golang.org/api/iterator"
)

type FieldAliases struct {
	OBJECTID        string `json:"OBJECTID"`
	DEPARTMENT      string `json:"DEPARTMENT"`
	TRASHHAULER     string `json:"TRASH_HAULER"`
	HEAVYTRASHDAY   string `json:"HEAVY_TRASH_DAY"`
	CENSUSTRACT2000 string `json:"CENSUS_TRACT_2000"`
	CALLBACKS       string `json:"CALLBACKS"`
	COUNCILDISTRICT string `json:"COUNCIL_DISTRICT"`
	INTERSECTION    string `json:"INTERSECTION"`
	TRASHDAY        string `json:"TRASH_DAY"`
	INCIDENTADDRESS string `json:"INCIDENT_ADDRESS"`
	TOWNSHIP        string `json:"TOWNSHIP"`
	TRASHDISTRICT   string `json:"TRASH_DISTRICT"`
	CITIZENADDRESS  string `json:"CITIZEN_ADDRESS"`
	CATEGORY        string `json:"CATEGORY"`
	KEYWORD         string `json:"KEYWORD"`
	SUBCATEGORY     string `json:"SUB_CATEGORY"`
	SOURCE          string `json:"SOURCE"`
	SRNUMBER        string `json:"SR_NUMBER"`
	PARENTSRNUMBER  string `json:"PARENT_SR_NUMBER"`
	HOMEPHONE       string `json:"HOME_PHONE"`
	SUBAREA         string `json:"SUBAREA"`
	LASTNAME        string `json:"LAST_NAME"`
	FIRSTNAME       string `json:"FIRST_NAME"`
	CONTACTTYPE     string `json:"CONTACT_TYPE"`
	WORKPHONE       string `json:"WORK_PHONE"`
	AGENT           string `json:"AGENT"`
	FOLLOWUP        string `json:"FOLLOW_UP"`
	CREATEDBY       string `json:"CREATED_BY"`
	OPENED          string `json:"OPENED"`
	OWNER           string `json:"OWNER"`
	DATEMODIFIED    string `json:"DATE_MODIFIED"`
	CLOSED          string `json:"CLOSED"`
	SEVERITY        string `json:"SEVERITY"`
	VERIFIEDADDRESS string `json:"VERIFIED_ADDRESS"`
	STATUS          string `json:"STATUS"`
	SUBSTATUS       string `json:"SUBSTATUS"`
	DATEOFTRANSFER  string `json:"DATE_OF_TRANSFER"`
	DESCRIPTION     string `json:"DESCRIPTION"`
	SR_NUMBER       string `json:"SR_NUMBER"`
}

type Field struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Alias  string `json:"alias"`
	Length int    `json:"length,omitempty"`
}

type Attributes struct {
	OBJECTID        int         `json:"OBJECTID"`
	DEPARTMENT      string      `json:"DEPARTMENT"`
	TRASHHAULER     string      `json:"TRASH_HAULER"`
	HEAVYTRASHDAY   string      `json:"HEAVY_TRASH_DAY"`
	CENSUSTRACT2000 int         `json:"CENSUS_TRACT_2000"`
	CALLBACKS       int         `json:"CALLBACKS"`
	COUNCILDISTRICT interface{} `json:"COUNCIL_DISTRICT"`
	INTERSECTION    interface{} `json:"INTERSECTION"`
	TRASHDAY        string      `json:"TRASH_DAY"`
	INCIDENTADDRESS string      `json:"INCIDENT_ADDRESS"`
	TOWNSHIP        string      `json:"TOWNSHIP"`
	TRASHDISTRICT   int         `json:"TRASH_DISTRICT"`
	CITIZENADDRESS  string      `json:"CITIZEN_ADDRESS"`
	CATEGORY        string      `json:"CATEGORY"`
	KEYWORD         string      `json:"KEYWORD"`
	SUBCATEGORY     string      `json:"SUB_CATEGORY"`
	SOURCE          string      `json:"SOURCE"`
	SRNUMBER        string      `json:"SR_NUMBER"`
	PARENTSRNUMBER  interface{} `json:"PARENT_SR_NUMBER"`
	HOMEPHONE       string      `json:"HOME_PHONE"`
	SUBAREA         string      `json:"SUBAREA"`
	LASTNAME        string      `json:"LAST_NAME"`
	FIRSTNAME       string      `json:"FIRST_NAME"`
	CONTACTTYPE     interface{} `json:"CONTACT_TYPE"`
	WORKPHONE       string      `json:"WORK_PHONE"`
	AGENT           int         `json:"AGENT"`
	FOLLOWUP        interface{} `json:"FOLLOW_UP"`
	CREATEDBY       interface{} `json:"CREATED_BY"`
	OPENED          int64       `json:"OPENED"`
	OWNER           interface{} `json:"OWNER"`
	DATEMODIFIED    int64       `json:"DATE_MODIFIED"`
	CLOSED          interface{} `json:"CLOSED"`
	SEVERITY        interface{} `json:"SEVERITY"`
	VERIFIEDADDRESS interface{} `json:"VERIFIED_ADDRESS"`
	STATUS          string      `json:"STATUS"`
	SUBSTATUS       interface{} `json:"SUBSTATUS"`
	DATEOFTRANSFER  int64       `json:"DATE_OF_TRANSFER"`
	DESCRIPTION     string      `json:"DESCRIPTION"`
}

type Feature struct {
	Attributes Attributes `json:"attributes"`
}

type PotholeResponse struct {
	DisplayFieldName string       `json:"displayFieldName"`
	FieldAliases     FieldAliases `json:"fieldAliases"`
	Fields           []Field      `json:"fields"`
	Features         []Feature    `json:"features"`
}

func getStreetView(pothole Feature) []byte {
	googleAPIKey := os.Getenv("GOOGLEAPIKEY")
	address := fmt.Sprintf("%s,Indianapolis,IN", pothole.Attributes.INCIDENTADDRESS)
	StreetViewurl := "https://maps.googleapis.com/maps/api/streetview"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", StreetViewurl, nil)
	q := req.URL.Query()
	q.Add("key", googleAPIKey)
	q.Add("location", address)
	q.Add("size", "640x480")
	q.Add("pitch", "-60")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	imageAsByteArr, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return imageAsByteArr
}

func tweet(image []byte, graph []byte, message string) {
	fmt.Println(message)
	config := oauth1.NewConfig(os.Getenv("TWITTERCONSUMERKEY"), os.Getenv("TWITTERCONSUMERSECRET"))
	token := oauth1.NewToken(os.Getenv("TWITTERACCESSTOKEN"), os.Getenv("TWITTERACCESSSECRET"))
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	tweetParams := &twitter.StatusUpdateParams{}
	picRes, _, err := client.Media.Upload(image, "image/jpeg")
	if err != nil {
		fmt.Println(err)
		return
	}
	graphRes, _, err2 := client.Media.Upload(graph, "image/png")
	if err != nil {
		fmt.Println(err2)
		return
	}
	tweetParams.MediaIds = []int64{}
	if picRes.MediaID > 0 {
		tweetParams.MediaIds = append(tweetParams.MediaIds, picRes.MediaID)
	}

	if graphRes.MediaID > 0 {
		tweetParams.MediaIds = append(tweetParams.MediaIds, graphRes.MediaID)
	}
	_, _, err3 := client.Statuses.Update(message, tweetParams)
	if err3 != nil {
		fmt.Println(err3)
		return
	}
}

func createClient(ctx context.Context) *firestore.Client {
	projectID := os.Getenv("GCPPROJECT")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		fsErrMessage := fmt.Sprintf("Failed to create client: %v", err)
		fmt.Println(fsErrMessage)
	}
	return client
}

func createGraph(xVals []float64, yVals []float64) []byte {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
			ValueFormatter: func(v interface{}) string {
				loc, _ := time.LoadLocation("America/Indiana/Indianapolis")
				typed := v.(float64)
				typedDate := time.Unix(0, int64(typed)).In(loc)
				return fmt.Sprintf("%02d/%02d/%d %02d:%02d", typedDate.Month(), typedDate.Day(), typedDate.Year(), typedDate.Hour(), typedDate.Minute())
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
			GridLines: []chart.GridLine{
				chart.GridLine{Value: 0.0},
				chart.GridLine{Value: 200.0},
				chart.GridLine{Value: 400.0},
				chart.GridLine{Value: 600.0},
				chart.GridLine{Value: 800.0},
				chart.GridLine{Value: 1000.0},
				chart.GridLine{Value: 1200.0},
				chart.GridLine{Value: 1400.0},
				chart.GridLine{Value: 1600.0},
				chart.GridLine{Value: 1800.0},
				chart.GridLine{Value: 2000.0},
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: xVals,
				YValues: yVals,
			},
		},
	}

	graphBuffer := bytes.NewBuffer([]byte{})
	gErr := graph.Render(chart.PNG, graphBuffer)
	if gErr != nil {
		fmt.Println(gErr)
	}
	return graphBuffer.Bytes()
}

func storeDataAndGetCountData(potholeCount int) ([]float64, []float64) {
	ctx := context.Background()
	dataClient := createClient(ctx)
	loc, _ := time.LoadLocation("America/Indiana/Indianapolis")
	_, _, fbErr := dataClient.Collection("potholeCount").Add(ctx, map[string]interface{}{
		"dateNano": time.Now().In(loc).UnixNano(),
		"count":    potholeCount,
	})
	if fbErr != nil {
		fbErrMessage := fmt.Sprintf("Failed adding pothole data to database: %v", fbErr)
		fmt.Println(fbErrMessage)
	}

	xVals := []float64{}
	yVals := []float64{}
	iter := dataClient.Collection("potholeCount").OrderBy("dateNano", firestore.Asc).StartAfter(time.Now().AddDate(0, -1, 0).UnixNano()).Documents(ctx)
	for {
		doc, fbReadErr := iter.Next()
		if fbReadErr == iterator.Done {
			break
		}
		if fbReadErr != nil {
			fbReadErrMessage := fmt.Sprintf("Failed retrieving pothole count data: %v", fbReadErr)
			fmt.Println(fbReadErrMessage)
			break
		} else {
			fmt.Println(doc.Data())
			xVals = append(xVals, float64(doc.Data()["dateNano"].(int64)))
			theCount := float64(doc.Data()["count"].(int64))
			fmt.Println(theCount)
			yVals = append(yVals, theCount)
		}
	}
	dataClient.Close()
	return xVals, yVals
}

func IndyPotholes(w http.ResponseWriter, r *http.Request) {
	potholeURL := "http://xmaps.indy.gov/arcgis/rest/services/PotholeViewer/PotholesClosed/MapServer/0/query?f=json&where=1%3D1&returnGeometry=false&spatialRel=esriSpatialRelIntersects&outFields=*&orderByFields=OPENED%20DESC"
	resp, err := http.Get(potholeURL)
	if err != nil {
		log.Fatal(err)
	}
	potholesBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var potholes PotholeResponse
	json.Unmarshal([]byte(potholesBody), &potholes)

	var potholeCount int
	potholeCount = len(potholes.Features)
	xVals, yVals := storeDataAndGetCountData(potholeCount)
	graphImage := createGraph(xVals, yVals)

	var randomPothole Feature
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := potholeCount - 1
	randomPothole = potholes.Features[rand.Intn(max-min+1)+min]
	imageBytes := getStreetView(randomPothole)
	randomDate := time.Unix(randomPothole.Attributes.OPENED/1000, 0)
	randomAddress := randomPothole.Attributes.INCIDENTADDRESS
	srNumber := randomPothole.Attributes.SRNUMBER
	detailURL := "http://maps.indy.gov/RequestIndy/proxy.ashx?http%3A%2F%2Fmaps.indy.gov%2FRequestIndyServices%2FServiceRequests%2F"
	message := fmt.Sprintf(`Current Open Potholes: %d
This Pothole entered at: %s
This Pothole Address: %s
Detail: %s%s`, potholeCount, randomDate, randomAddress, detailURL, srNumber)
	tweet(imageBytes, graphImage, message)
}
