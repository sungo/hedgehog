package sonic

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

// FIXME: tokenize auth

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/dghubble/sling"
)

const (
	version    = "0.0.1"
	clientName = "hedgehog"
)

var (
	clientID  = fmt.Sprintf("%s-%s", clientName, version)
	userAgent = fmt.Sprintf("%s/%s", clientName, version)
)

type (
	Sonic struct {
		auth Auth
		base string
	}

	Auth struct {
		User     string `url:"u"`
		Password string `url:"p"`
	}

	Song struct {
		ID      string `json:"id"`
		Album   string `json:"album"`
		Artist  string `json:"artist"`
		Path    string `json:"path"`
		Track   int    `json:"track"`
		Title   string `json:"title"`
		IsVideo bool   `json:"isVideo"`
		Suffix  string `json:"suffix"`
	}
	Songs []Song

	GetStarredResponseWrapper struct {
		Status   string             `json:"status"`
		Response GetStarredResponse `json:"subsonic-response"`
	}

	GetStarredResponse struct {
		Status        string           `json:"status"`
		Version       string           `json:"version"`
		Type          string           `json:"type"`
		ServerVersion string           `json:"serverVersion"`
		Starred       map[string]Songs `json:"starred"`
	}

	GetPlaylistsResponseWrapper struct {
		Status   string               `json:"status"`
		Response GetPlaylistsResponse `json:"subsonic-response"`
	}

	GetPlaylistsResponse struct {
		Data GetPlaylistsWrapper `json:"playlists"`
	}

	GetPlaylistsWrapper struct {
		Playlists ListingOfPlaylists `json:"playlist"`
	}

	PlaylistListing struct {
		ID        string `json:"ID"`
		Name      string `json:"name"`
		SongCount int    `json:"songCount"`
		Duration  string `json:"duration"`
	}
	ListingOfPlaylists []PlaylistListing

	GetPlaylistResponseWrapper struct {
		Status   string              `json:"status"`
		Response GetPlaylistResponse `json:"subsonic-response"`
	}

	GetPlaylistResponse struct {
		Playlist Playlist `json:"playlist"`
	}

	Playlist struct {
		ID        string `json:"ID"`
		Name      string `json:"name"`
		SongCount int    `json:"songCount"`
		Duration  int    `json:"duration"`
		Songs     Songs  `json:"entry"`
	}
)

func New(auth Auth, urlBase string) Sonic {
	return Sonic{auth: auth, base: urlBase}
}

func (client Sonic) url(path string) string {
	return fmt.Sprintf("%s/%s", client.base, path)
}

func (client Sonic) sling() *sling.Sling {
	return sling.New().Set("User-Agent", userAgent)
}

func (client Sonic) GetPlaylists() (ListingOfPlaylists, error) {
	var resp GetPlaylistsResponseWrapper

	params := struct {
		Format   string `url:"f"`
		User     string `url:"u"`
		Password string `url:"p"`
		ClientID string `url:"c"`
	}{"json", client.auth.User, client.auth.Password, clientID}

	_, err := sling.New().Post(client.url("rest/getPlaylists")).
		Set("User-Agent", userAgent).
		BodyForm(params).
		ReceiveSuccess(&resp)
	if err != nil {
		return ListingOfPlaylists{}, err
	}

	return resp.Response.Data.Playlists, nil
}

func (client Sonic) GetPlaylist(id string) (Playlist, error) {
	if id == "" {
		return Playlist{}, errors.New("provide an id")
	}

	var resp GetPlaylistResponseWrapper
	// var resp interface{}

	params := struct {
		Format     string `url:"f"`
		User       string `url:"u"`
		Password   string `url:"p"`
		ClientID   string `url:"c"`
		PlaylistID string `url:"id"`
	}{"json", client.auth.User, client.auth.Password, clientID, id}

	_, err := client.sling().New().
		Post(client.url("rest/getPlaylist")).
		BodyForm(params).
		ReceiveSuccess(&resp)
	if err != nil {
		return Playlist{}, err
	}

	return resp.Response.Playlist, nil
}

func (client Sonic) GetStarred() (map[string]bool, error) {
	var (
		resp GetStarredResponseWrapper
		data = make(map[string]bool)
	)

	params := struct {
		Format   string `url:"f"`
		User     string `url:"u"`
		Password string `url:"p"`
		ClientID string `url:"c"`
	}{"json", client.auth.User, client.auth.Password, clientID}

	_, err := client.sling().New().
		Post(client.url("rest/getStarred")).
		BodyForm(params).
		ReceiveSuccess(&resp)
	if err != nil {
		return data, err
	}
	for idx := range resp.Response.Starred["song"] {
		song := resp.Response.Starred["song"][idx]
		data[song.ID] = true
	}

	return data, nil
}

func (playlist Playlist) Shuffle() Playlist {
	newPlaylist := playlist

	rand.Shuffle(len(playlist.Songs), func(i int, j int) {
		newPlaylist.Songs[i], newPlaylist.Songs[j] = newPlaylist.Songs[j], newPlaylist.Songs[i]
	})

	return newPlaylist
}

func (client Sonic) DownloadSong(song Song) ([]byte, error) {
	params := struct {
		Format   string `url:"f"`
		User     string `url:"u"`
		Password string `url:"p"`
		ClientID string `url:"c"`
		SongID   string `url:"id"`
	}{"json", client.auth.User, client.auth.Password, clientID, song.ID}

	req, err := client.sling().New().
		Post(client.url("rest/download")).
		BodyForm(params).
		Request()
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	return resBody, err
}

func (client Sonic) ScrobbleNowPlaying(song Song) {
	params := struct {
		Format       string `url:"f"`
		User         string `url:"u"`
		Password     string `url:"p"`
		ClientID     string `url:"c"`
		ID           string `url:"id"`
		IsSubmission bool   `url:"submission"`
	}{"json", client.auth.User, client.auth.Password, clientID, song.ID, false}

	client.sling().New().
		Post(client.url("rest/scrobble")).
		BodyForm(params).
		ReceiveSuccess(nil)
}

func (client Sonic) ScrobbleSubmit(song Song) {
	params := struct {
		Format       string `url:"f"`
		User         string `url:"u"`
		Password     string `url:"p"`
		ClientID     string `url:"c"`
		ID           string `url:"id"`
		IsSubmission bool   `url:"submission"`
	}{"json", client.auth.User, client.auth.Password, clientID, song.ID, true}

	client.sling().New().
		Post(client.url("rest/scrobble")).
		BodyForm(params).
		ReceiveSuccess(nil)
}
