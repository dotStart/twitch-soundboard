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
package main

import (
	"flag"
	"fmt"
	"github.com/dotStart/twitch-soundboard/bot"
	"github.com/dotStart/twitch-soundboard/registry"
	"github.com/faiface/beep"
	"github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

func main() {
	var logLevel string
	var registryPath string
	var sampleRate uint
	var queueSize uint
	var volume float64
	var globalRateLimit time.Duration
	var userRateLimit time.Duration

	var helpFlag bool

	flag.StringVar(&logLevel, "log-level", "info", "alters the log granularity")
	flag.BoolVar(&helpFlag, "help", false, "displays this message")
	flag.StringVar(&registryPath, "registry", "sounds", "specifies the location at which sound files will be stored")
	flag.DurationVar(&globalRateLimit, "global-rate-limit", 2*time.Second, "specifies the global rate limit")
	flag.DurationVar(&userRateLimit, "user-rate-limit", 10*time.Second, "specifies the user rate limit")
	flag.UintVar(&sampleRate, "sample-rate", 44100, "specifies the desired sample rate")
	flag.UintVar(&queueSize, "queue-size", 8, "specifies how many sounds may queue up at once")
	flag.Float64Var(&volume, "volume", -1, "specifies the volume at which sounds will play")

	flag.Parse()

	fmt.Println("Twitch SoundBoard v0.1.0") // TODO: compile time version information
	fmt.Println("Copyright (C) 2019 Johannes Donath <https://github.com/dotStart>")
	fmt.Println()

	if flag.NArg() < 3 {
		fmt.Fprint(os.Stderr, "error: must specify bot name and token as well as at least one channel")
		os.Exit(1)
	}

	name := flag.Arg(0)
	token := flag.Arg(1)

	channels := make([]string, flag.NArg()-2)
	for i := 0; i < flag.NArg()-2; i++ {
		channels[i] = flag.Arg(2 + i)
	}

	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	level, err := logging.LogLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: illegal log level \"%s\": %s", logLevel, err)
		os.Exit(1)
	}

	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.4s}] %{module} : %{color:reset} %{message}`)
	backend := logging.AddModuleLevel(logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format))
	backend.SetLevel(level, "")
	logging.SetBackend(backend)

	logger := logging.MustGetLogger("soundboard")
	logger.Infof("set log level to %s", level)
	logger.Infof("fetching sounds from %s", registryPath)
	logger.Infof("set sample rate to %d Hz", sampleRate)

	cfg := registry.DefaultConfig()
	cfg.Path = registryPath
	cfg.SampleRate = beep.SampleRate(sampleRate)
	cfg.QueueSize = queueSize
	cfg.Volume = volume

	var g errgroup.Group

	reg, err := registry.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
		os.Exit(1)
	}

	b := bot.New(reg, bot.Config{
		Name:            name,
		Token:           token,
		GlobalRateLimit: globalRateLimit,
		UserRateLimit:   userRateLimit,
	})

	for _, ch := range channels {
		b.Join(ch)
	}

	g.Go(func() error {
		return reg.Listen()
	})
	g.Go(func() error {
		reg.PollQueue()
		return nil
	})
	g.Go(func() error {
		return b.Connect()
	})

	err = g.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
		os.Exit(2)
	}
}
