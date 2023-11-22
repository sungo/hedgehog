package main

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

import (
	"github.com/alecthomas/kong"

	"git.sr.ht/~sungo/hedgehog/pkg/player"
)

type (
	Cmd struct {
		User         string `kong:"required,name='user',env='SONIC_USER',help='subsonic user name'"`
		Password     string `kong:"required,name='password',env='SONIC_PASSWORD',help='subsonic password (sent in the url unencrypted)'"`
		URL          string `kong:"required,name='url',env='SONIC_URL',help='url to the server (like https://music.wat)'"`
		PlaylistName string `kong:"required,name='playlist',env='SONIC_PLAYLIST'"`
		Shuffle      bool   `kong:"optional,name='shuffle'"`
		Repeat       bool   `kong:"optional,name='repeat'"`
	}
)

func main() {
	ctx := kong.Parse(&Cmd{})
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (cmd Cmd) Run() error {
	return player.Start(player.Config{
		User:         cmd.User,
		Password:     cmd.Password,
		URL:          cmd.URL,
		PlaylistName: cmd.PlaylistName,
		Shuffle:      cmd.Shuffle,
		Repeat:       cmd.Repeat,
	})
}
