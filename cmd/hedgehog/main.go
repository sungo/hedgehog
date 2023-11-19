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
		BasePath     string `kong:"optional,name='base-path',env='SONIC_BASE_PATH',help='base file directory (prepended to the music file path)'"`
		PlaylistName string `kong:"required,name='playlist',env='SONIC_PLAYLIST'"`
		Shuffle      bool   `kong:"optional,name='shuffle'"`
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
		BasePath:     cmd.BasePath,
		PlaylistName: cmd.PlaylistName,
		Shuffle:      cmd.Shuffle,
	})
}
