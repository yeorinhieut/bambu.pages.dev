package main

import (
    "crypto/tls"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

type UpdateRequest struct {
    PrinterIP    string `json:"printerIp"`
    DeviceID     string `json:"sn"`
    AccessCode   string `json:"accessCode"`
    PrinterModel string `json:"printerModel"`
}

var mqttClient mqtt.Client

func main() {
    r := gin.Default()
    r.Use(cors.Default())

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "pong"})
    })

    r.POST("/terminate", func(c *gin.Context) {
        go func() {
            time.Sleep(1 * time.Second)
            os.Exit(0)
        }()
        c.JSON(http.StatusOK, gin.H{"message": "Terminating server."})
    })

    r.POST("/update", func(c *gin.Context) {
        var updateReq UpdateRequest
        if err := c.ShouldBindJSON(&updateReq); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
            return
        }

        payload, err := getPayload(updateReq.PrinterModel)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to retrieve payload."})
            return
        }

        fmt.Println("Payload:", payload) // 콘솔에 payload 출력

        mqttOpts := mqtt.NewClientOptions()
        mqttOpts.AddBroker(fmt.Sprintf("tcps://%s:8883", updateReq.PrinterIP))
        mqttOpts.SetClientID("go-mqtt-client")
        mqttOpts.SetUsername("bblp")
        mqttOpts.SetPassword(updateReq.AccessCode)
        mqttOpts.SetTLSConfig(&tls.Config{
            InsecureSkipVerify: true,
        })
        mqttOpts.SetProtocolVersion(4) // MQTTv311
        mqttOpts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
            fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
        })
        mqttOpts.OnConnect = func(client mqtt.Client) {
            fmt.Println("MQTT connection established")

            // 응답 수신을 위한 주제 구독
            reportTopic := fmt.Sprintf("device/%s/report", updateReq.DeviceID)
            if token := client.Subscribe(reportTopic, 0, messagePubHandler); token.Wait() && token.Error() != nil {
                fmt.Println("Subscription error:", token.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"message": "Subscription error"})
                return
            }

            token := client.Publish(fmt.Sprintf("device/%s/request", updateReq.DeviceID), 0, false, payload)
            token.Wait()
            if token.Error() != nil {
                fmt.Println("Error publishing:", token.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"message": "Error publishing"})
                return
            }
            c.JSON(http.StatusOK, gin.H{"message": "Update request sent."})
        }
        mqttOpts.OnConnectionLost = func(client mqtt.Client, err error) {
            fmt.Printf("Connection lost: %v", err)
        }

        mqttClient = mqtt.NewClient(mqttOpts)
        if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
            fmt.Println(token.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"message": "MQTT connection failed"})
            return
        }
    })

    r.Run("0.0.0.0:1883")
}

func messagePubHandler(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func getPayload(modelName string) (string, error) {
    url := fmt.Sprintf("https://raw.githubusercontent.com/lunDreame/user-bambulab-firmware/main/assets/%s_AMS.json", modelName)
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to retrieve payload, status code: %d", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
