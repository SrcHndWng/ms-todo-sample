package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const tasksURL = "https://graph.microsoft.com/beta/me/outlook/tasks"
const refreshTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
const redirectURL = "https://login.microsoftonline.com/common/oauth2/nativeclient"

// TasksResult contains To-Do tasks from api.
type TasksResult struct {
	OdataContext  string `json:"@odata.context"`
	OdataNextLink string `json:"@odata.nextLink"`
	Value         []struct {
		OdataEtag            string        `json:"@odata.etag"`
		ID                   string        `json:"id"`
		CreatedDateTime      time.Time     `json:"createdDateTime"`
		LastModifiedDateTime time.Time     `json:"lastModifiedDateTime"`
		ChangeKey            string        `json:"changeKey"`
		Categories           []interface{} `json:"categories"`
		AssignedTo           string        `json:"assignedTo"`
		HasAttachments       bool          `json:"hasAttachments"`
		Importance           string        `json:"importance"`
		IsReminderOn         bool          `json:"isReminderOn"`
		Owner                string        `json:"owner"`
		ParentFolderID       string        `json:"parentFolderId"`
		Sensitivity          string        `json:"sensitivity"`
		Status               string        `json:"status"`
		Subject              string        `json:"subject"`
		DueDateTime          interface{}   `json:"dueDateTime"`
		Recurrence           interface{}   `json:"recurrence"`
		ReminderDateTime     interface{}   `json:"reminderDateTime"`
		StartDateTime        interface{}   `json:"startDateTime"`
		Body                 struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		CompletedDateTime struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		} `json:"completedDateTime"`
	} `json:"value"`
}

// RefreshResult contains token refresh result from api
type RefreshResult struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func getTasks(accessToken string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", tasksURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := new(http.Client)
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	return
}

func refreshTokens(oldRefreshToken string) (accessToken, refreshToken string, err error) {
	values := url.Values{}
	values.Add("client_id", os.Getenv("MS_TO-DO_CLIENT_ID"))
	values.Add("scope", "offline_access user.read tasks.read")
	values.Add("refresh_token", oldRefreshToken)
	values.Add("redirect_uri", redirectURL)
	values.Add("grant_type", "refresh_token")

	req, err := http.NewRequest(
		"POST",
		refreshTokenURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result RefreshResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return
	}

	accessToken = result.AccessToken
	refreshToken = result.RefreshToken

	return
}

func getTokens() (accessToken, refreshToken string) {
	accessToken = os.Getenv("MS_TO-DO_ACCESS_TOKEN")
	refreshToken = os.Getenv("MS_TO-DO_REFRESH_TOKEN")
	return
}

func setTokens(accessToken, refreshToken string) {
	os.Setenv("MS_TO-DO_ACCESS_TOKEN", accessToken)
	os.Setenv("MS_TO-DO_REFRESH_TOKEN", refreshToken)
}

func printTasks(resp *http.Response) (err error) {
	var result TasksResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return
	}

	fmt.Println("print tasks status, name...")
	for _, v := range result.Value {
		println(v.Status + " : " + v.Subject)
	}

	return
}

func main() {
	for {
		accessToken, refreshToken := getTokens()

		resp, err := getTasks(accessToken)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Println("token unauthorized.")
			resp.Body.Close()

			newAccessToken, newRefreshToken, err := refreshTokens(refreshToken)
			if err != nil {
				panic(err)
			}

			setTokens(newAccessToken, newRefreshToken)

			resp, err = getTasks(newAccessToken)
			if err != nil {
				panic(err)
			}
		}

		printTasks(resp)

		time.Sleep(10 * time.Second)
	}
}
