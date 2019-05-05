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
package bot

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

const hastebinBaseUrl string = "https://hasteb.in"

type UploadResult struct {
	Key string `json:"key"`
}

func (b *Bot) uploadSoundList() (string, error) {
	b.logger.Debugf("generating sound list")

	var list string
	list += "Available Sounds\n"
	list += "================\n"
	list += "\n"

	for _, sound := range b.reg.ListSounds() {
		list += " * "
		list += sound
		list += "\n"
	}

	b.logger.Debugf("sound list document: %s", list)
	b.logger.Debugf("uploading sound list document")

	resp, err := http.Post(hastebinBaseUrl+"/documents", "text/plain", strings.NewReader(list))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b.logger.Debugf("service responded with status \"%s\"", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Upload failed")
	}

	b.logger.Debugf("decoding service response")
	decoder := json.NewDecoder(resp.Body)

	var result UploadResult
	err = decoder.Decode(&result)
	if err != nil {
		return "", err
	}

	b.logger.Debugf("document stored with key %s", result.Key)
	return hastebinBaseUrl + "/" + result.Key + ".md", nil
}
