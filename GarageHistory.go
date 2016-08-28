package main

import (
    "flag"
    "os"
    "os/signal"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// Declare exit values
const (
    Success = iota
    ErrorConnecting
    ErrorSubscribing
    ErrorDisconnecting
    ErrorOpeningDatabase
    ErrorCreatingTable
)

type Status string
const (
    Opened Status = "O"
    Closed Status = "C"
)



func main() {
    /////////////////////// Parse Command Line Args /////////////////////////
    // Parse the logging level
    var logLevel string
    flag.StringVar(&logLevel, "l", "INFO", "The logging level string. {DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL}")

    // Port for the broker
    port := flag.Int("p", 1883, "The port number for the MQTT broker")

    // MQTT Host
    var hostname string
    flag.StringVar(&hostname, "h", "localhost", "The hostname of the MQTT broker")

    // Update topic. This is the topic where garage status updates are received
    var updateTopic string
    flag.StringVar(&updateTopic, "u", "home/garage/door/update",
                    "The topic to listen for garage door updates on")

    // History request topic. This is the topic where we listen for requests
    var requestTopic string
    flag.StringVar(&requestTopic, "r", "home/server/garage/historyrequest",
                    "The topic to listen for history requests on")

    // Database path
    var databasePath string
    flag.StringVar(&databasePath, "d", "./GarageHistory.db", "The path to the "+
                    "sqlite database storing the history")
    // Parse additional arguments here...

    // Do the actual parsing
    flag.Parse()

    // Try to get the log level from the cmd line
    level, err := logging.LogLevel(logLevel)

    // If the log level can't be parsed, default to info
    if err != nil {
        level = logging.INFO
        SetupLogging(level)
        log.Warningf("The command line log level '%s' could not be parsed. "+
                     "Defaulting to INFO. Use the -h option for help.", logLevel)
    }else {
        // Setup logging with the parsed level
        SetupLogging(level)
    }
    log.Debug("Logging setup properly")
    ///////////////////// Done With Command Line Args /////////////////////////

    //// Connecting to the Database
    db, err := sql.Open("sqlite3", databasePath)
    if err != nil {
        log.Criticalf("Error opening database connection: %s", err)
        os.Exit(ErrorOpeningDatabase)
    }
    log.Debugf("Made connection to database at path: %s", databasePath)
    defer db.Close()

    // Setup the table structure
    err = CreateHistoryTableIfNeeded(db)
    if err != nil {
        log.Criticalf("Error when creating history table: %s", err)
        os.Exit(ErrorCreatingTable)
    }
    log.Debugf("Created tables successfully")

    //// Connect to the Broker
    cli, err := ConnectToBroker(hostname, *port)

    // Make sure the connection went smoothly
    if err != nil {
        log.Criticalf("Fatal error connecting to MQTT Broker: %s", err)
        os.Exit(ErrorConnecting)
    }
    log.Debugf("Successfully connected to MQTT broker %s:%i", hostname, *port)

    // Subscribe to the request and update topics that we need to listen to
    err = SubscribeToTopics(cli, db, requestTopic, updateTopic)

    // Make sure we subscribed to topics okay
    if err != nil {
        log.Criticalf("Fatal Error subscribing to topics: %s", err)
        os.Exit(ErrorSubscribing)
    }
    log.Debugf("Subscribed to topics '%s' and '%s'", requestTopic, updateTopic)

    if err != nil {
        log.Errorf("Unable to send status request message. "+
        "The current door status could be incorrect: %s", err)
    }

    ////////////////////////////////////////////////////////
    // Set up channel on which to send signal notifications.
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, os.Interrupt, os.Kill)

    log.Info("Initialization successful. Waiting for requests or updates")

    // Wait for receiving a signal.
    <-sigc

    // Disconnect the Network Connection.
    if err := cli.Disconnect(); err != nil {
        log.Errorf("Error while disconnecting: %s", err)
        os.Exit(ErrorDisconnecting)
    }
}
