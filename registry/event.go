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
	"github.com/fsnotify/fsnotify"
	"runtime"
	"time"
)

func (sr *SoundRegistry) handleEvents() {
	for {
		select {
		case event, ok := <-sr.watcher.Events:
			if !ok {
				sr.logger.Warningf("failed to poll for registry events - aborted")
				return
			}

			switch event.Op {
			case fsnotify.Create:
				sr.handleCreation(event.Name)
			case fsnotify.Remove:
				sr.handleRemoval(event.Name)
			case fsnotify.Write:
				sr.handleChange(event.Name)
			}
		case err, ok := <-sr.watcher.Errors:
			if !ok {
				sr.logger.Warningf("failed to poll for registry errors - aborted")
				return
			}

			sr.logger.Warningf("error encountered while polling registry events: %s", err)
		}
	}
}

func (sr *SoundRegistry) handleCreation(path string) {
	sr.logger.Debugf("received creation notification for file: %s", path)

	name, ext, err := splitPath(path)
	if err != nil {
		sr.logger.Debugf("ignored invalid sound file: %s", path)
		return
	}

	_, exists := sr.sounds[name]
	if exists {
		sr.logger.Debugf("ignored sound file \"%s\": already registered", path)
		return
	}

	// Windows file locking work around
	if runtime.GOOS == "windows" {
		sr.logger.Debugf("delaying sound registration")
		time.Sleep(time.Second * 2)
	}

	sr.load(path, name, ext)
}

func (sr *SoundRegistry) handleRemoval(path string) {
	sr.logger.Debugf("received removal notification for file: %s", path)

	name, _, err := splitPath(path)
	if err != nil {
		sr.logger.Debugf("ignored invalid sound file: %s", path)
		return
	}

	_, exists := sr.sounds[name]
	if !exists {
		sr.logger.Debugf("ignored invalid sound file: %s", path)
		return
	}

	delete(sr.sounds, name)
	sr.logger.Infof("removed file \"%s\"", path)
}

func (sr *SoundRegistry) handleChange(path string) {
	sr.logger.Debugf("received change notification for file: %s", path)

	name, ext, err := splitPath(path)
	if err != nil {
		sr.logger.Debugf("ignored invalid sound file: %s", path)
		return
	}

	sr.load(path, name, ext)
}
