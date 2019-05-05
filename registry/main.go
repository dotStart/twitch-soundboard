/*
 * Copyright 2019 Johannes Donath <johannesd@torchmind.com>
 * and other copyright owners as documented in the project's IP log.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package registry

import (
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)
import "github.com/hashicorp/errwrap"
import "github.com/fsnotify/fsnotify"
import "github.com/op/go-logging"
import "github.com/faiface/beep"

type SoundRegistry struct {
	logger *logging.Logger
	cfg    Config

	watcher            *fsnotify.Watcher
	notifyShutdownFlag chan bool

	queue             chan string
	queueShutdownFlag chan bool

	sounds map[string]beep.StreamSeeker
}

func New(cfg Config) (*SoundRegistry, error) {
	_, err := os.Stat(cfg.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errwrap.Wrapf("failed to access sound registry: {{err}}", err)
		}

		err = os.MkdirAll(cfg.Path, 0774)
		if err != nil {
			return nil, errwrap.Wrapf("failed to create registry directory: {{err}}", err)
		}
	}

	err = speaker.Init(cfg.SampleRate, cfg.SampleRate.N(100*time.Millisecond))
	if err != nil {
		return nil, errwrap.Wrapf("failed to open speaker: {{err}}", err)
	}

	reg := &SoundRegistry{
		logger: logging.MustGetLogger("registry"),
		cfg:    cfg,

		queue:  make(chan string, cfg.QueueSize),
		sounds: make(map[string]beep.StreamSeeker),
	}

	reg.Scan()

	return reg, nil
}

func (sr *SoundRegistry) Listen() error {
	var err error
	sr.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return errwrap.Wrapf("failed to create registry watcher: {{err}}", err)
	}

	sr.notifyShutdownFlag = make(chan bool)
	go sr.handleEvents()

	err = sr.watcher.Add(sr.cfg.Path)
	if err != nil {
		return errwrap.Wrapf("failed to register registry watcher: {{err}}", err)
	}

	sr.logger.Info("listening to registry updates")
	<-sr.notifyShutdownFlag
	return nil
}

func (sr *SoundRegistry) Close() error {
	sr.notifyShutdownFlag <- true
	sr.queueShutdownFlag <- true
	return sr.watcher.Close()
}
