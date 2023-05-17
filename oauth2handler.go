package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func generateStateOauthCookie(w http.ResponseWriter) (string, error) {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration, HttpOnly: true, Path: "/oauth2/callback"}
	http.SetCookie(w, cookie)

	return state, nil
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	oauthStateString, _ := generateStateOauthCookie(w)
	url := discordOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func discordOauthCallback(w http.ResponseWriter, r *http.Request) {
	oauthStateCookie, noCookieError := r.Cookie("oauthstate")
	if noCookieError != nil {
		fmt.Printf("No oauth cookie found, no entry allowed\n")
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	oauthStateParam := r.FormValue("state")
	if oauthStateCookie.Value != oauthStateParam {
		fmt.Printf("Invalid OAuth state received. Expected '%s', received '%s'\n", oauthStateCookie, oauthStateParam)
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	code := r.FormValue("code")
	token, err := discordOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Auth code -> access token exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	fmt.Printf("Access token: '%s'\n", token.AccessToken)

	userInfo, err := getUserInfo(token)
	if err != nil {
		fmt.Printf("Error getting user info: '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	fmt.Fprintf(w, "Hello, '%s'\n", userInfo.Tag)
}

type User struct {
	Id    string `json:"id"`
	Tag   string `json:"username"`
	Email string `json:"email"`
}

func getUserInfo(token *oauth2.Token) (*User, error) {
	res, err := discordOauthConfig.Client(context.Background(), token).Get("https://discord.com/api/users/@me")
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status code: %d", res.StatusCode))
	}

	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(bytes, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
