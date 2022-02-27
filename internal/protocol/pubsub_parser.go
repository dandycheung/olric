// Copyright 2018-2022 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"github.com/buraksezer/olric/internal/util"
	"github.com/tidwall/redcon"
)

func ParsePublishCommand(cmd redcon.Command) (*Publish, error) {
	if len(cmd.Args) < 3 {
		return nil, errWrongNumber(cmd.Args)
	}

	return NewPublish(
		util.BytesToString(cmd.Args[1]), // Channel
		util.BytesToString(cmd.Args[2]), // Message
	), nil
}

func ParseSubscribeCommand(cmd redcon.Command) (*Subscribe, error) {
	if len(cmd.Args) < 2 {
		return nil, errWrongNumber(cmd.Args)
	}

	var channels []string
	args := cmd.Args[1:]
	for len(args) > 0 {
		arg := util.BytesToString(args[0])
		channels = append(channels, arg)
		args = args[1:]
	}
	return NewSubscribe(channels...), nil
}

func ParsePSubscribeCommand(cmd redcon.Command) (*PSubscribe, error) {
	if len(cmd.Args) < 2 {
		return nil, errWrongNumber(cmd.Args)
	}

	var patterns []string
	args := cmd.Args[1:]
	for len(args) > 0 {
		arg := util.BytesToString(args[0])
		patterns = append(patterns, arg)
		args = args[1:]
	}
	return NewPSubscribe(patterns...), nil
}

func ParsePubSubChannelsCommand(cmd redcon.Command) (*PubSubChannels, error) {
	if len(cmd.Args) < 2 {
		return nil, errWrongNumber(cmd.Args)
	}

	ps := NewPubSubChannels()
	if len(cmd.Args) >= 3 {
		ps.SetPattern(util.BytesToString(cmd.Args[2]))
	}
	return ps, nil
}

func ParsePubSubNumpatCommand(cmd redcon.Command) (*PubSubNumpat, error) {
	if len(cmd.Args) < 2 {
		return nil, errWrongNumber(cmd.Args)
	}

	return NewPubSubNumpat(), nil
}

func ParsePubSubNumsubCommand(cmd redcon.Command) (*PubSubNumsub, error) {
	if len(cmd.Args) < 2 {
		return nil, errWrongNumber(cmd.Args)
	}

	var channels []string
	args := cmd.Args[2:]
	for len(args) > 0 {
		arg := util.BytesToString(args[0])
		channels = append(channels, arg)
		args = args[1:]
	}
	return NewPubSubNumsub(channels...), nil
}