package main

type UpdateMessage struct {
    DoorName string
    Status string
    Timestamp int64
}

func (message UpdateMessage) StatusValid() bool {
    return message.Status == "open" || message.Status == "closed"
}
