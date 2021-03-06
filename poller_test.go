package gomessagestore_test

import (
	"context"
	"testing"

	. "github.com/blackhatbrigade/gomessagestore"
	mock_gomessagestore "github.com/blackhatbrigade/gomessagestore/mocks"
	mock_repository "github.com/blackhatbrigade/gomessagestore/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

type getMessagesParams struct {
	position int64
}
type getMessagesReturns struct {
	messages []Message
	err      error
}

type processMessagesParams struct {
	messages []Message
}
type processMessagesReturns struct {
	msgsHandled int
	lastPos     int64
	err         error
}

type setPositionParams struct {
	position int64
}
type setPositionReturns struct {
	err error
}

func TestPoller(t *testing.T) {
	errorCalled := false
	errorHandler := func(error) {
		errorCalled = true
	}

	tests := []struct {
		name               string
		expectedErrors     []error
		subOpts            []SubscriberOption
		handlers           []MessageHandler
		foundPosition      int64
		getMsgsParams      []getMessagesParams
		getMsgsReturns     []getMessagesReturns
		processMsgsParams  []processMessagesParams
		processMsgsReturns []processMessagesReturns
		setPosParams       []setPositionParams
		setPosReturns      []setPositionReturns
		foundPositionError error
		callPollNumTimes   int
		onError            bool
	}{{
		name: "It ran",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
		},
		handlers:           []MessageHandler{},
		callPollNumTimes:   1,
		getMsgsParams:      []getMessagesParams{{0}},
		getMsgsReturns:     []getMessagesReturns{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil}},
		processMsgsParams:  []processMessagesParams{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))}},
		processMsgsReturns: []processMessagesReturns{{2, 1012, nil}},
		expectedErrors:     []error{nil},
	}, {
		name: "GetPosition Errors are returned",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
		},
		handlers:           []MessageHandler{},
		foundPositionError: potato,
		expectedErrors:     []error{potato},
		callPollNumTimes:   1,
	}, {
		name: "GetMessages Errors are returned",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
		},
		handlers:         []MessageHandler{},
		expectedErrors:   []error{potato},
		callPollNumTimes: 1,
		getMsgsParams:    []getMessagesParams{{0}},
		getMsgsReturns:   []getMessagesReturns{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), potato}},
	}, {
		name: "ProcessMessages Errors cause the onError func to be called",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			OnError(errorHandler),
		},
		handlers:           []MessageHandler{},
		expectedErrors:     []error{potato},
		callPollNumTimes:   1,
		getMsgsParams:      []getMessagesParams{{0}},
		getMsgsReturns:     []getMessagesReturns{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil}},
		processMsgsParams:  []processMessagesParams{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))}},
		processMsgsReturns: []processMessagesReturns{{2, 1012, potato}},
		onError:            true,
	}, {
		name: "ProcessMessages Without Errors doesn't cause the onError func to be called",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			OnError(errorHandler),
		},
		handlers:           []MessageHandler{},
		expectedErrors:     []error{nil},
		callPollNumTimes:   1,
		getMsgsParams:      []getMessagesParams{{0}},
		getMsgsReturns:     []getMessagesReturns{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil}},
		processMsgsParams:  []processMessagesParams{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))}},
		processMsgsReturns: []processMessagesReturns{{2, 1012, nil}},
	}, {
		name: "SetPosition Errors are returned",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(7),
		},
		handlers:           []MessageHandler{},
		expectedErrors:     []error{potato},
		callPollNumTimes:   1,
		getMsgsParams:      []getMessagesParams{{0}},
		getMsgsReturns:     []getMessagesReturns{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil}},
		processMsgsParams:  []processMessagesParams{{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))}},
		processMsgsReturns: []processMessagesReturns{{10, 1012, nil}},
		setPosParams:       []setPositionParams{{1013}},
		setPosReturns:      []setPositionReturns{{potato}},
	}, {
		name: "When called twice, Poll uses a changed value for starting position",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 2,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{5, 9000, nil},
		},
		expectedErrors: []error{nil, nil},
	}, {
		name: "SetPosition is called when the correct amount of messages are processed",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(7),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 2,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{5, 9000, nil},
		},
		setPosParams:   []setPositionParams{{9001}},
		setPosReturns:  []setPositionReturns{{nil}},
		expectedErrors: []error{nil, nil},
	}, {
		name: "SetPosition is called (multiple times) when the correct amount of messages are processed",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(5),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 3,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
			{9001},
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{3, 9000, nil},
			{2, 1000000, nil},
		},
		setPosParams: []setPositionParams{
			{1013},
			{1000001},
		},
		setPosReturns: []setPositionReturns{
			{nil},
			{nil},
		},
		expectedErrors: []error{nil, nil, nil},
	}, {
		name: "SetPosition is called (multiple, multiple times) when the correct amount of messages are processed",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(5),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 5,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
			{4001},
			{6001},
			{9001},
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 109)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 112)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 109))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 112))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{1, 4000, nil},
			{7, 6000, nil},
			{1, 9000, nil},
			{2, 1000000, nil}, // shouldn't call, because we only have 3 messages here
		},
		setPosParams: []setPositionParams{
			{1013},
			{6001},
		},
		setPosReturns: []setPositionReturns{
			{nil},
			{nil},
		},
		expectedErrors: []error{nil, nil, nil, nil, nil},
	}, {
		name: "SetPosition is called when the correct amount of messages are processed, unless ProcessMessages errors out",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(5),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 3,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
			{1013}, // won't advance; will check the same message over and over and over until it passes
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{3, 9000, potato},
			{2, 1000000, nil},
		},
		setPosParams: []setPositionParams{
			{1013},
		},
		setPosReturns: []setPositionReturns{
			{nil},
		},
		expectedErrors: []error{nil, potato, nil},
	}, {
		name: "If SetPosition errors out, it doesn't reset the count of the number of messages handled",
		subOpts: []SubscriberOption{
			SubscribeToCommandStream("some cat"),
			UpdatePositionEvery(5),
		},
		handlers:         []MessageHandler{},
		callPollNumTimes: 3,
		getMsgsParams: []getMessagesParams{
			{0},
			{1013},
			{9001},
		},
		getMsgsReturns: []getMessagesReturns{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103)), nil},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106)), nil},
		},
		processMsgsParams: []processMessagesParams{
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 100))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 103))},
			{eventsToMessageSlice(getLotsOfSampleEvents(3, 106))},
		},
		processMsgsReturns: []processMessagesReturns{
			{5, 1012, nil},
			{3, 9000, nil},
			{2, 1000000, nil},
		},
		setPosParams: []setPositionParams{
			{1013},
			{9001},
		},
		setPosReturns: []setPositionReturns{
			{potato},
			{nil},
		},
		expectedErrors: []error{potato, nil, nil},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()

			// mocks and expectations
			mockRepo := mock_repository.NewMockRepository(ctrl)

			myWorker := mock_gomessagestore.NewMockSubscriptionWorker(ctrl)

			myWorker.
				EXPECT().
				GetPosition(ctx).
				Return(test.foundPosition, test.foundPositionError)

			var lastCall *gomock.Call
			for index, _ := range test.getMsgsParams {
				thisCall := myWorker.
					EXPECT().
					GetMessages(ctx, test.getMsgsParams[index].position).
					Return(test.getMsgsReturns[index].messages, test.getMsgsReturns[index].err)
				if lastCall != nil {
					thisCall.After(lastCall)
				}
				lastCall = thisCall
			}
			lastCall = nil

			for index, _ := range test.processMsgsParams {
				thisCall := myWorker.
					EXPECT().
					ProcessMessages(ctx, test.processMsgsParams[index].messages).
					Return(test.processMsgsReturns[index].msgsHandled, test.processMsgsReturns[index].lastPos, test.processMsgsReturns[index].err)
				if lastCall != nil {
					thisCall.After(lastCall)
				}
				lastCall = thisCall
			}
			lastCall = nil

			for index, _ := range test.setPosParams {
				thisCall := myWorker.
					EXPECT().
					SetPosition(ctx, test.setPosParams[index].position).
					Return(test.setPosReturns[index].err)
				if lastCall != nil {
					thisCall.After(lastCall)
				}
				lastCall = thisCall
			}
			lastCall = nil

			// setup
			var logrusLogger = logrus.New()
			msgStore := NewMessageStoreFromRepository(mockRepo, logrusLogger)
			opts, err := GetSubscriberConfig(test.subOpts...)
			myPoller, err := CreatePoller(
				msgStore,
				myWorker,
				opts,
			)
			if err != nil {
				t.Errorf("Failed on CreatePoller() Got: %s\n", err)
				return
			}

			// call
			for c := 0; c < test.callPollNumTimes; c++ {
				errorCalled = false // always reset our check
				err = myPoller.Poll(ctx)

				// assertions
				if err != test.expectedErrors[c] {
					t.Errorf("Failed on Poll()\nWant: %s\nHave: %s\n", test.expectedErrors[c], err)
					return
				}

				if test.onError != errorCalled {
					t.Errorf("Poll failed to call onError \nWant: %t\nHave: %t\n", test.onError, errorCalled)
				}
			}
		})
	}
}
