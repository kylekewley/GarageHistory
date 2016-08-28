package main

import (
    "time"
    "errors"
)

type HistoryRequestMessage struct {
    CurrentValue bool
    StartUnixTime int
    EndUnixTime int
    ReturnTopic string
}

func (message HistoryRequestMessage) Valid() error {
    if !message.CurrentValue {
        if message.EndUnixTime < message.StartUnixTime  {
            return errors.New("The end time is greater than the start timestamp")
        }else if (message.StartUnixTime < 0) {
            return errors.New("The start time is less than zero")
        }else if (message.EndUnixTime < 0) {
            return errors.New("The end time is less than zero")
        }else if len(message.ReturnTopic) == 0 {
            return errors.New("There was no return topic supplied")
        }
    }

    return nil
}

type HistoryResponse struct {
    DoorName string
    Timestamp time.Time
    Status string
}
