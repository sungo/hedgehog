package main

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

import (
	"github.com/alecthomas/kong"

	"git.sr.ht/~sungo/hedgehog/pkg/player"
)

type (
	Cmd struct {
		User           string `kong:"required,name='user',env='SONIC_USER',help='subsonic user name'"`
		Password       string `kong:"required,name='password',env='SONIC_PASSWORD',help='subsonic password (sent in the url unencrypted)'"`
		URL            string `kong:"required,name='url',env='SONIC_URL',help='url to the server (like https://music.wat)'"`
		PlaylistName   string `kong:"required,name='playlist',env='SONIC_PLAYLIST',help='which playlist to play'"`
		Shuffle        bool   `kong:"optional,negatable,name='shuffle',env='SONIC_SHUFFLE',help='shuffle the track order'"`
		Repeat         bool   `kong:"optional,negatable,default=true,name='repeat',env='SONIC_REPEAT',help='when we run out of stuff to play, start over (with --shuffle, the list is reshuffled)'"`
		ReloadOnRepeat bool   `kong:"optional,negatable,default=true,name'reload-on-repeat',env='SONIC_RELOAD_REPEAT',help='when we run out of stuff to play, automatically refresh the playlist'"`
	}
)

func main() {
	ctx := kong.Parse(&Cmd{})
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd Cmd) Run() error {
	return player.Start(player.Config{
		User:           cmd.User,
		Password:       cmd.Password,
		URL:            cmd.URL,
		PlaylistName:   cmd.PlaylistName,
		Shuffle:        cmd.Shuffle,
		Repeat:         cmd.Repeat,
		ReloadOnRepeat: cmd.ReloadOnRepeat,
	})
}
