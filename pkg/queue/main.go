package queue

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

import (
	"fmt"
	"os"
	"time"

	"git.sr.ht/~sungo/hedgehog/pkg/sonic"
)

type Entry struct {
	Meta        sonic.Song
	LocalFile   string
	Downloading bool
}

type (
	UpNext []*Entry
	Queue  struct {
		Playlist sonic.Playlist
		Shuffle  bool
		Idx      int
		Depth    int
		Playing  *Entry
		UpNext   UpNext
		Client   *sonic.Sonic

		songs sonic.Songs
	}
)

func (queue *Queue) Fetch(entry *Entry) error {
	entry.Downloading = true
	defer func() { entry.Downloading = false }()

	song := entry.Meta
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("hedgehog-*.%s", song.Suffix))
	if err != nil {
		return err
	}
	// fmt.Printf("==> [BK] Downloading %s as %s\n", song.Title, tmpFile.Name())

	data, err := queue.Client.DownloadSong(song)
	if err != nil {
		return err
	}

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	entry.LocalFile = tmpFile.Name()
	return nil
}

func (entry *Entry) Remove() {
	if entry.LocalFile == "" {
		return
	}

	for entry.Downloading == true {
		time.Sleep(250 * time.Millisecond)
	}
	os.Remove(entry.LocalFile)
	entry.LocalFile = ""
}

func (up UpNext) Clear() {
	for idx := range up {
		up[idx].Remove()
	}
	up = make(UpNext, 0)
}

func (queue *Queue) CleanUp() {
	queue.UpNext.Clear()
	if queue.Playing != nil {
		queue.Playing.Remove()
		queue.Playing = nil
	}
}

func (queue *Queue) WhatsNext() *Entry {
	if len(queue.Playlist.Songs) == 0 {
		return nil
	}

	if len(queue.songs) == 0 {
		if queue.Shuffle {
			shuffled := queue.Playlist.Shuffle()
			queue.songs = shuffled.Songs
		} else {
			queue.songs = queue.Playlist.Songs
		}
	}

	if len(queue.UpNext) > 0 {
		next := queue.UpNext[0]
		queue.Playing = next
		queue.UpNext = queue.UpNext[1:]
	}

	for len(queue.UpNext) < queue.Depth {
		if len(queue.songs) == 0 {
			break
		}
		next := Entry{Meta: queue.songs[0]}
		if queue.Playing == nil {
			queue.Playing = &next

			if err := queue.Fetch(&next); err != nil {
				panic(err)
			}
		} else {
			queue.UpNext = append(queue.UpNext, &next)
			go func() {
				if err := queue.Fetch(&next); err != nil {
					panic(err)
				}
			}()
		}
		queue.songs = queue.songs[1:]
	}
	for queue.Playing.Downloading == true {
		time.Sleep(250 * time.Millisecond)
	}
	return queue.Playing
}
