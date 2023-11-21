package mpv

// Code originally developed by sungo (https://sungo.io)
// Distributed under the terms of the 0BSD license https://opensource.org/licenses/0BSD

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/blang/mpv"
)

type Instance struct {
	running    bool
	socketPath string

	mpv *mpv.Client
	cmd *exec.Cmd
}

func New(socketPath string) Instance {
	return Instance{socketPath: socketPath}
}

func (inst *Instance) PauseToggle() {
	if inst.mpv == nil {
		return
	}
	ok, _ := inst.mpv.Pause()
	inst.mpv.SetPause(!ok)
}

func (inst *Instance) MuteToggle() {
	if inst.mpv == nil {
		return
	}
	ok, _ := inst.mpv.Mute()
	inst.mpv.SetMute(!ok)
}

func (inst *Instance) LaunchAndBlock(ctx context.Context, started chan bool) chan error {
	errChan := make(chan error)

	go func() {
	LOOP:
		for {
			inst.running = true
			runErr := make(chan error)

			go inst.runOne(runErr, started)
			select {
			case err := <-runErr:
				errChan <- err
			case <-ctx.Done():
				errChan <- nil
				break LOOP

			}
		}
		inst.running = false
	}()

	return errChan
}

type PlayNotification struct {
	PercentComplete float64
}

func (inst *Instance) Next() {
	if inst.mpv == nil {
		return
	}
	inst.mpv.Exec("stop")
}

func (inst *Instance) Play(path string) chan PlayNotification {
	notif := make(chan PlayNotification)
	inst.mpv.Loadfile(path, mpv.LoadFileModeReplace)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			pct, err := inst.mpv.PercentPosition()
			if err != nil {
				close(notif)
				return
			}
			notif <- PlayNotification{PercentComplete: pct}
			if pct <= 0 {
				close(notif)
				return
			}
		}
	}()

	return notif
}

func (inst *Instance) Shutdown() {
	if inst.cmd != nil {
		inst.cmd.Process.Kill()
	}
}

func (inst *Instance) runOne(errChan chan error, started chan bool) {
	inst.cmd = exec.Command(
		"mpv",
		"--idle",
		fmt.Sprintf("--input-ipc-server=%s", inst.socketPath),
	)

	err := inst.cmd.Start()
	if err != nil {
		inst.mpv = nil
		errChan <- err
		return
	}
	time.Sleep(1 * time.Second)

	ipcc := mpv.NewIPCClient(inst.socketPath)
	inst.mpv = mpv.NewClient(ipcc)

	started <- true
	err = inst.cmd.Wait()

	inst.mpv = nil
	inst.cmd = nil

	if err != nil {
		errChan <- err
	}
}
