// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package poll

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/twitchdev/twitch-cli/internal/events"
	"github.com/twitchdev/twitch-cli/internal/models"
	"github.com/twitchdev/twitch-cli/internal/util"
)

var transportsSupported = map[string]bool{
	models.TransportEventSub: true,
}

var triggerSupported = []string{"poll-begin", "poll-progress", "poll-end"}

var triggerMapping = map[string]map[string]string{
	models.TransportEventSub: {
		"poll-begin":    "channel.poll.begin",
		"poll-progress": "channel.poll.progress",
		"poll-end":      "channel.poll.end",
	},
}

type Event struct{}

func (e Event) GenerateEvent(params events.MockEventParameters) (events.MockEventResponse, error) {
	var event []byte
	var err error

	if params.Description == "" {
		params.Description = "Pineapple on pizza?"
	}

	switch params.Transport {
	case models.TransportEventSub:
		choices := []models.PollEventSubEventChoice{}
		for i := 1; i < 5; i++ {
			c := models.PollEventSubEventChoice{
				ID:    util.RandomGUID(),
				Title: fmt.Sprintf("Yes but choice %v", i),
			}
			if params.Trigger != "poll-begin" {
				c.BitsVotes = intPointer(int(util.RandomInt(10)))
				c.ChannelPointsVotes = intPointer(int(util.RandomInt(10)))
				c.Votes = intPointer(*c.BitsVotes + *c.ChannelPointsVotes + int(util.RandomInt(10)))
			}
			choices = append(choices, c)
		}

		body := &models.PollEventSubResponse{
			Subscription: models.EventsubSubscription{
				ID:      params.ID,
				Status:  "enabled",
				Type:    triggerMapping[params.Transport][params.Trigger],
				Version: e.SubscriptionVersion(),
				Condition: models.EventsubCondition{
					BroadcasterUserID: params.ToUserID,
				},
				Transport: models.EventsubTransport{
					Method:   "webhook",
					Callback: "null",
				},
				Cost:      0,
				CreatedAt: util.GetTimestamp().Format(time.RFC3339Nano),
			},
			Event: models.PollEventSubEvent{
				ID:                   util.RandomGUID(),
				BroadcasterUserID:    params.ToUserID,
				BroadcasterUserLogin: params.ToUserName,
				BroadcasterUserName:  params.ToUserName,
				Title:                params.Description,
				Choices:              choices,
				BitsVoting: models.PollEventSubEventGoodVoting{
					IsEnabled:     true,
					AmountPerVote: 10,
				},
				ChannelPointsVoting: models.PollEventSubEventGoodVoting{
					IsEnabled:     true,
					AmountPerVote: 500,
				},
				StartedAt: util.GetTimestamp().Format(time.RFC3339Nano),
			},
		}

		if params.Trigger == "poll-end" {
			body.Event.EndedAt = util.GetTimestamp().Add(time.Minute * 15).Format(time.RFC3339Nano)
			body.Event.Status = "completed"
		} else {
			body.Event.EndsAt = util.GetTimestamp().Add(time.Minute * 15).Format(time.RFC3339Nano)
		}

		event, err = json.Marshal(body)
		if err != nil {
			return events.MockEventResponse{}, err
		}
	default:
		return events.MockEventResponse{}, nil
	}

	return events.MockEventResponse{
		ID:       params.ID,
		JSON:     event,
		FromUser: params.FromUserID,
		ToUser:   params.ToUserID,
	}, nil
}

func (e Event) ValidTransport(t string) bool {
	return transportsSupported[t]
}

func (e Event) ValidTrigger(t string) bool {
	for _, ts := range triggerSupported {
		if ts == t {
			return true
		}
	}
	return false
}
func (e Event) GetTopic(transport string, trigger string) string {
	return triggerMapping[transport][trigger]
}

func intPointer(i int) *int {
	return &i
}
func (e Event) GetEventSubAlias(t string) string {
	// check for aliases
	for trigger, topic := range triggerMapping[models.TransportEventSub] {
		if topic == t {
			return trigger
		}
	}
	return ""
}

func (e Event) SubscriptionVersion() string {
	return "1"
}
