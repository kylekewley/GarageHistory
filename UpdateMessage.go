package main

type UpdateMessage struct {
    GarageID int
    StatusChange string
}

type RequestUpdateMessage struct {
    // Leave empty for all IDs
    GarageIDs []int
}

func (message UpdateMessage) StatusValid() bool {
    return message.StatusChange == "O" || message.StatusChange == "C"
}
