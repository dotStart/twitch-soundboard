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
	"github.com/dotStart/twitch-soundboard/registry"
	"github.com/gempir/go-twitch-irc"
	"github.com/hashicorp/errwrap"
	"github.com/op/go-logging"
	"time"
)

type Bot struct {
	logger *logging.Logger
	cfg    Config

	reg    *registry.SoundRegistry
	client *twitch.Client

	indexLink *string

	lastCommand     time.Time
	lastUserCommand map[string]time.Time
}

func New(reg *registry.SoundRegistry, cfg Config) *Bot {
	client := twitch.NewClient(cfg.Name, cfg.Token)

	bot := &Bot{
		logger: logging.MustGetLogger("bot"),
		cfg:    cfg,

		reg:    reg,
		client: client,

		lastCommand:     time.Unix(0, 0),
		lastUserCommand: make(map[string]time.Time),
	}

	client.OnConnect(bot.handleConnect)
	client.OnPrivateMessage(bot.handlePrivateMessage)

	return bot
}

func (b *Bot) Join(ch string) {
	b.logger.Infof("Configured channel \"%s\"", ch)
	b.client.Join(ch)
}

func (b *Bot) Connect() error {
	err := b.client.Connect()
	if err != nil {
		return errwrap.Wrapf("failed to establish connection to Twitch: {{err}}", err)
	}
	return nil
}
