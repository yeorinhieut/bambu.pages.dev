from flask import Flask, request, jsonify
from flask_cors import CORS
import paho.mqtt.client as mqtt
import json
import ssl
import threading
import os
import time
import requests

app = Flask(__name__)
CORS(app)

def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("MQTT 연결 성공")
        client.publish(userdata['topic'], userdata['payload'])
    else:
        print(f"MQTT 연결 실패, 코드: {rc}")

def on_publish(client, userdata, mid):
    print("펌웨어 업데이트 요청이 성공적으로 전송되었습니다.")
    client.disconnect()

@app.route('/ping', methods=['GET'])
def ping():
    return jsonify({"message": "pong"})

@app.route('/terminate', methods=['POST'])
def terminate():
    print("Terminating proxy server.")
    func = request.environ.get('werkzeug.server.shutdown')
    if func is not None:
        func()
    response = jsonify({"message": "Terminating server."})
    response.status_code = 200

    def shutdown_server():
        time.sleep(1)
        os._exit(0)

    threading.Thread(target=shutdown_server).start()
    return response

def get_payload(model_name):
    url = f"https://raw.githubusercontent.com/lunDreame/user-bambulab-firmware/main/assets/{model_name}_AMS.json"
    response = requests.get(url)
    if response.status_code == 200:
        return response.json()
    else:
        return None

@app.route('/update', methods=['POST'])
def update():
    data = request.json
    printer_ip = data['printerIp']
    device_id = data['sn']
    access_code = data['accessCode']
    model_name = data['printerModel']

    payload = get_payload(model_name)
    if not payload:
        return jsonify({"message": "Failed to retrieve payload."}), 400

    mqtt_broker = printer_ip
    mqtt_port = 8883
    mqtt_username = "bblp"
    mqtt_password = access_code
    mqtt_topic = f"device/{device_id}/request"

    client = mqtt.Client(userdata={'topic': mqtt_topic, 'payload': json.dumps(payload)}, protocol=mqtt.MQTTv311)
    client.username_pw_set(mqtt_username, mqtt_password)
    client.tls_set(cert_reqs=ssl.CERT_NONE)
    client.tls_insecure_set(True)
    print(payload)

    client.on_connect = on_connect
    client.on_publish = on_publish

    client.connect(mqtt_broker, mqtt_port, 60)
    client.loop_start()

    return jsonify({"message": "Update request sent."})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=1883)
