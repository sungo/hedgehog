package player

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.sr.ht/~sungo/hedgehog/pkg/mpv"
	"git.sr.ht/~sungo/hedgehog/pkg/queue"
	"git.sr.ht/~sungo/hedgehog/pkg/sonic"

	"github.com/eiannone/keyboard"
	progressbar "github.com/schollz/progressbar/v3"
)

type Config struct {
	User     string
	Password string
	URL      string

	PlaylistName string
	Shuffle      bool
}

func Start(config Config) error {
	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	tempDir, err := os.MkdirTemp("", "hedgehog-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	client := sonic.New(
		sonic.Auth{
			User:     config.User,
			Password: config.Password,
		},
		config.URL,
	)
	fmt.Println("Fetching playlists...")
	playlists, err := client.GetPlaylists()
	if err != nil {
		return err
	}

	var playlistID string

	for idx := range playlists {
		playlist := playlists[idx]

		if playlist.Name == config.PlaylistName {
			playlistID = playlist.ID
			break
		}
	}

	if playlistID == "" {
		return errors.New("unable to find playlist")
	}
	fmt.Printf("Found playlist '%s' as %s'\n", config.PlaylistName, playlistID)

	playlist, err := client.GetPlaylist(playlistID)
	if err != nil {
		return err
	}

	if len(playlist.Songs) == 0 {
		return errors.New("empty playlist")
	}

	if config.Shuffle {
		fmt.Println("Shuffling...")
		playlist = playlist.Shuffle()
	}

	q := queue.New()
	q.Playlist = playlist
	q.Depth = 3
	q.Shuffle = config.Shuffle
	q.Client = &client
	q.TempDir = tempDir
	defer q.CleanUp()

	started := make(chan bool)
	music := mpv.New(fmt.Sprintf("%s/mpv.sock", tempDir))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bye := func() {
		cancel()
		q.CleanUp()
		music.Shutdown()
		os.RemoveAll(tempDir)
	}

	fmt.Println("Launching backend...")
	go func() {
		select {
		case err := <-music.LaunchAndBlock(ctx, started):
			bye()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}()

	<-started
	defer music.Shutdown()

	// TIME TO PLAY SONGS YAY
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		<-sigc
		bye()
		os.Exit(1)
	}()

	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				panic(err)
			}
			// fmt.Printf("You pressed: rune %q, key %X\r\n", char, key)
			switch {
			case key == keyboard.KeyCtrlC:
				fallthrough
			case char == 'q':
				fallthrough
			case key == keyboard.KeyEsc:
				bye()
				os.Exit(0)

			case char == 'm':
				music.MuteToggle()

			case char == 'p':
				fallthrough
			case char == '<':
				q.Previous()
				music.Next()

			case char == 'n':
				fallthrough
			case char == '>':
				music.Next()

			case key == keyboard.KeySpace:
				music.PauseToggle()
			}
		}
	}()

	for {

		bar := progressbar.NewOptions(100,
			progressbar.OptionFullWidth(),
			progressbar.OptionClearOnFinish(),
		)

		var (
			lastPercent float64

			song = q.WhatsNext()
		)

		bar.Describe(fmt.Sprintf("|> %s : %s", song.Meta.Artist, song.Meta.Title))
		client.ScrobbleNowPlaying(song.Meta)
		for msg := range music.Play(song.LocalFile) {

			lastPercent = msg.PercentComplete
			bar.Set(int(msg.PercentComplete))
		}

		if lastPercent >= 75 {
			client.ScrobbleSubmit(song.Meta)
		}

		bar.Finish()

		fmt.Printf("=> %s - %s\n", song.Meta.Artist, song.Meta.Title)
		song.Remove()
	}

	return nil
}
