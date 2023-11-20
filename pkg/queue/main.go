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
	entryList []*Entry
	Queue     struct {
		Playlist sonic.Playlist
		Shuffle  bool
		Depth    int
		Client   *sonic.Sonic

		Playing  *Entry
		upNext   entryList
		previous entryList

		songs sonic.Songs
	}
)

func New() *Queue {
	queue := Queue{
		upNext:   make(entryList, 0),
		previous: make(entryList, 0),
	}

	return &queue
}

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

func (list entryList) Clear() {
	for idx := range list {
		list[idx].Remove()
	}
	list = make(entryList, 0)
}

func (queue *Queue) CleanUp() {
	queue.upNext.Clear()
	queue.previous.Clear()
	if queue.Playing != nil {
		queue.Playing.Remove()
		queue.Playing = nil
	}
}

func (queue *Queue) Previous() {
	if len(queue.Playlist.Songs) == 0 {
		return
	}
	if len(queue.songs) == 0 {
		return
	}
	if len(queue.previous) == 0 {
		return
	}

	prev := queue.previous[len(queue.previous)-1]
	if prev.LocalFile == "" {
		if err := queue.Fetch(prev); err != nil {
			panic(err)
		}

		for prev.Downloading == true {
			time.Sleep(250 * time.Millisecond)
		}
	}

	if queue.Playing == nil {
		queue.upNext = append(entryList{prev}, queue.upNext...)
	} else {
		for queue.Playing.Downloading == true {
			time.Sleep(250 * time.Millisecond)
		}

		queue.upNext = append(entryList{prev, queue.Playing}, queue.upNext...)
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

	if len(queue.upNext) > 0 {
		if queue.Playing != nil {
			queue.previous = append(queue.previous, queue.Playing)
		}

		queue.Playing = queue.upNext[0]
		queue.upNext = queue.upNext[1:]
	}

	for len(queue.upNext) < queue.Depth {
		if len(queue.songs) == 0 {
			break
		}
		nextQueued := &Entry{Meta: queue.songs[0]}

		if queue.Playing == nil {
			queue.Playing = nextQueued

			if err := queue.Fetch(nextQueued); err != nil {
				panic(err)
			}
		} else {
			queue.upNext = append(queue.upNext, nextQueued)
			go func() {
				if err := queue.Fetch(nextQueued); err != nil {
					panic(err)
				}
			}()
		}
		queue.songs = queue.songs[1:]
	}

	if queue.Playing.Downloading == false && queue.Playing.LocalFile == "" {
		if err := queue.Fetch(queue.Playing); err != nil {
			panic(err)
		}
	}

	for queue.Playing.Downloading == true {
		time.Sleep(250 * time.Millisecond)
	}
	return queue.Playing
}
