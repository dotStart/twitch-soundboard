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
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func splitPath(path string) (string, string, error) {
	stripped := filepath.Base(path)

	i := strings.LastIndex(stripped, ".")
	if i == -1 {
		return "", "", fmt.Errorf("illegal sound file: %s", path)
	}

	return stripped[:i], stripped[i+1:], nil
}

func (sr *SoundRegistry) Scan() {
	sr.logger.Infof("scanning registry for sounds")

	files, err := ioutil.ReadDir(sr.cfg.Path)
	if err != nil {
		sr.logger.Warningf("failed to scan registry directory: %s", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(sr.cfg.Path, file.Name())
		name, ext, err := splitPath(path)
		if err != nil {
			sr.logger.Debugf("ignoring invalid sound file \"%s\": %s", path, err)
			continue
		}

		sr.load(path, name, ext)
	}
}

func (sr *SoundRegistry) load(path string, name string, ext string) {
	f, err := os.Open(path)
	if err != nil {
		sr.logger.Warningf("failed to open sound file \"%s\": %s", path, err)
		return
	}

	var streamer beep.StreamSeeker
	var format beep.Format
	switch ext {
	case "flac":
		streamer, format, err = flac.Decode(f)
	case "mp3":
		streamer, format, err = mp3.Decode(f)
	case "ogg":
		streamer, format, err = vorbis.Decode(f)
	case "wav":
		streamer, format, err = wav.Decode(f)
	default:
		sr.logger.Debugf("ignoring file of unknown type: %s", path)
		return
	}

	if err != nil {
		sr.logger.Warningf("failed to read sound file \"%s\": %s", path, err)
		return
	}

	sr.logger.Debugf("== File Report \"%s\" ==", path)
	sr.logger.Debugf("File Format: %s", ext)
	sr.logger.Debugf("Sample Rate: %d Hz", format.SampleRate)
	sr.logger.Debugf("Channels: %d", format.NumChannels)

	if format.SampleRate != sr.cfg.SampleRate {
		sr.logger.Warningf("sample rate mismatch: sound \"%s\" may play at an undesired rate")
	}

	sr.sounds[name] = streamer
	sr.logger.Infof("loaded sound \"%s\" from file \"%s\"", name, path)
}
