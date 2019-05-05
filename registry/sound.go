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
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

func (sr *SoundRegistry) PollQueue() {
	sr.logger.Infof("starting play queue")

	for {
		select {
		case name := <-sr.queue:
			sr.playSound(name)
		case <-sr.queueShutdownFlag:
			return
		}
	}
}

func (sr *SoundRegistry) playSound(name string) {
	streamer, exists := sr.sounds[name]
	if !exists {
		sr.logger.Debugf("ignoring play request - no such sound: %s", name)
		return
	}

	err := streamer.Seek(0)
	if err != nil {
		sr.logger.Warningf("failed to seek to start of sound \"%s\"", name)
		return
	}

	sr.logger.Infof("playing sound \"%s\"", name)

	volume := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   sr.cfg.Volume,
		Silent:   false,
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func (sr *SoundRegistry) Play(name string) {
	select {
	case sr.queue <- name:
		sr.logger.Debugf("queued sound \"%s\"", name)
	default:
		sr.logger.Debugf("queue overflow - ignoring sound request for \"%s\"", name)
	}
}

func (sr *SoundRegistry) ListSounds() []string {
	keys := make([]string, 0)
	for key := range sr.sounds {
		keys = append(keys, key)
	}
	return keys
}
