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
	"github.com/gempir/go-twitch-irc"
	"strings"
)

func (b *Bot) handleConnect() {
	b.logger.Infof("established connection to Twitch")
}

func (b *Bot) handlePrivateMessage(msg twitch.PrivateMessage) {
	if !strings.HasPrefix(msg.Message, "!") {
		return
	}

	if !b.checkLimit(msg.User.ID) {
		b.logger.Debugf("Ignoring command from %s due to rate limit", msg.User.ID)
		return
	}
	b.updateLimit(msg.User.ID)

	if strings.HasPrefix(msg.Message, "!sounds") {
		var link string
		if b.indexLink == nil {
			var err error
			link, err = b.uploadSoundList()
			if err != nil {
				b.logger.Errorf("failed to upload sounds list: %s", err)
				b.client.Say(msg.Channel, "@"+msg.User.Name+": Failed to upload sound list")
				return
			}

			b.indexLink = &link
		} else {
			link = *b.indexLink
		}

		b.client.Say(msg.Channel, "@"+msg.User.Name+": "+link)
		return
	}

	b.reg.Play(msg.Message[1:])
}
