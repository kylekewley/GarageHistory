package main

import (
    "fmt"
    "database/sql"
    "encoding/json"
    "github.com/kylekewley/gmq/mqtt"
    "github.com/kylekewley/gmq/mqtt/client"
)

/* Broker code */
func SubscribeToTopics(cli *client.Client, db *sql.DB, requestTopic string, updateTopic string) error {
    // Subscribe to topics.
    err := cli.Subscribe(&client.SubscribeOptions{
        SubReqs: []*client.SubReq{
            &client.SubReq{
                TopicFilter: []byte(requestTopic),
                QoS:         mqtt.QoS2,
                Handler: func(topicName, message []byte) {
                    HandleHistoryRequest(db, cli, string(topicName), message)
                },
            },
            &client.SubReq{
                TopicFilter: []byte(updateTopic),
                QoS:         mqtt.QoS2,
                Handler: func(topicName, message []byte) {
                    HandleUpdateMessage(db, string(topicName), message)
                },
            },
        },
    })

    return err
}

func ConnectToBroker(host string, port int, username string, password string, successHandler func(*client.Client)) (*client.Client, error) {
    // Create an MQTT Client.
    cli := client.New(&client.Options{
        ErrorHandler: func(err error) {
            log.Errorf("MQTT client error: %s", err)
        },
        ConnectHandler: successHandler,
      })

    options := &client.ConnectOptions{
        Network:  "tcp",
        Address:  fmt.Sprintf("%s:%d", host, port),
        ClientID: []byte("GarageHistoryServer"),
    }

    // Only set the username if not empty
    if len(username) > 0 {
        options.UserName = []byte(username)
        options.Password = []byte(password)
    }

    // Connect to the MQTT Server.
    err := cli.Connect(options)

    return cli, err
}

/* Handler code */
func HandleUpdateMessage(db *sql.DB, topicName string, message []byte) {
    // Parse the message into an UpdateMessage type
    update := UpdateMessage{}
    err := json.Unmarshal(message, &update)
    if err != nil {
        log.Errorf("Error parsing update message: '%s", string(message))
        return
    }

    if !update.StatusValid() {
        log.Errorf("Invalid status received: '%s'", update.Status)
        return
    }

    // Write the message to the database
    err = WriteHistoryEvent(db, update)

    if err != nil {
        log.Errorf("Error writing update to the database: '%+v' error: '%s'", update, err)
    }else {
        log.Infof("Wrote update from topic '%s' to database: %+v", topicName, update)
    }
}

func HandleHistoryRequest(db *sql.DB, cli *client.Client, topicName string, message []byte) {
    // Parse the message into an UpdateMessage type
    request := HistoryRequestMessage{}
    err := json.Unmarshal(message, &request)
    if err != nil {
        log.Errorf("Error parsing request message: '%s", string(message))
        return
    }

    err = request.Valid()
    if err != nil {
        log.Errorf("Invalid request received: '%+v' Reason: '%s'", request, err)
        return
    }

    // Get the history values from the database
    values,err := QueryHistoryEvents(db, request)

    if err != nil {
        log.Errorf("Error querying the database with request: '%+v' error: '%s'", request, err)
        return
    }else {
        log.Infof("Request: '%+v' returned values: %+v", request, values)
    }

    message,err = json.Marshal(values)
    if err != nil {
        log.Errorf("Error converting array of history records to string: '%+v'", values)
        return
    }

    // Send the values to the topic that they were requested to
    err = cli.Publish(&client.PublishOptions{
        QoS:    mqtt.QoS2,
        TopicName: []byte(request.ReturnTopic),
        Message: []byte(message),
    })

    if err != nil {
        log.Errorf("Error sending history data on topic '%s': '%s'", request.ReturnTopic, err)
        return
    }
}
