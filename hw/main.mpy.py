import utime
from machine import Pin, reset
from umqtt.simple import MQTTClient
import ntptime
import urequests
import network
from sys import print_exception
import ujson

# WiFi connection configuration
WIFI_SSID = "Wokwi-GUEST"
WIFI_PASSWORD = ""

# MQTT broker configuration
MQTT_BROKER = "broker.hivemq.com"
MQTT_TOPIC = "siyangz/home/gate"

# GPIO configuration
LED_PIN = 17
LOCK_PIN = 2

# Server API configuration
API_URL = "https://backend-ri6qxvjyda-uw.a.run.app/api/log"

# Time synchronization configuration
NTP_SERVER = "asia.pool.ntp.org"

# Logging function
def log(module: str, message: str):
    current_time = get_current_time()
    if module:
        print("{} [{}] {}".format(current_time, module, message))
    else:
        print("{} {}".format(current_time, message))

# Get current time
def get_current_time() -> str:
    year, month, day, hour, minute, second, _, _ = utime.localtime()
    return "{:04d}-{:02d}-{:02d}T{:02d}:{:02d}:{:02d}Z".format(
        year, month, day, hour, minute, second
    )

# MQTT message callback function
def mqtt_callback(topic, message):
    log("callback", "topic: {}, message: {}".format(topic, message))

    if topic.decode("utf-8") != MQTT_TOPIC:
        return

    message_data = ujson.loads(message.decode("utf-8"))
    log("callback", "passthrough: {}".format(message_data["passthrough"]))
    if message_data["command"] != "open":
        return

    log("locker", "opened")
    unlock_door()
    log("locker", "lock closed")

    log("locker", "notify API...")
    notify_door_open()
    log("locker", "notify API done")

# Unlock the door
def unlock_door():
    lock_pin = Pin(LOCK_PIN, Pin.OUT)
    lock_pin.on()
    utime.sleep(1)
    lock_pin.off()

# Notify that the door is open
def notify_door_open():
    response = urequests.post(API_URL, json={"event": "gate_open"})
    log(None, "API call response: status_code: {}, headers: {}, text: {}".format(
        response.status_code, ujson.dumps(response.headers), ujson.dumps(response.text)))
    if response.status_code == 200:
        log(None, "API call successful")
    else:
        log(None, "API call failed")

# Main function
def main():
    log(None, "Program started")

    log(None, "Connecting to WiFi...")
    # Connect to WiFi
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)
    wlan.connect(WIFI_SSID, WIFI_PASSWORD)
    while not wlan.isconnected():
        pass
    log(None, "WiFi connected")

    log(None, "Synchronizing time...")
    # Synchronize time
    ntptime.host = NTP_SERVER
    ntptime.settime()
    log(None, "Time synchronized")

    log(None, "Connecting to MQTT broker...")
    # Connect to MQTT broker
    MQTT_CLIENT_ID = "esp32-{}".format(str(round(utime.time_ns()/1_000_000)))
    log(None, "MQTT client ID: {}".format(MQTT_CLIENT_ID))
    client = MQTTClient(MQTT_CLIENT_ID, MQTT_BROKER)
    client.set_callback(mqtt_callback)
    client.connect()
    client.subscribe(MQTT_TOPIC)
    log(None, "Connected to MQTT broker")

    # Status LED
    led_pin = Pin(LED_PIN, Pin.OUT)
    led_pin.on()
    utime.sleep(0.05)
    led_pin.off()
    utime.sleep(0.30)
    led_pin.on()
    utime.sleep(0.05)
    led_pin.off()

    log(None, "Awaiting message...")
    while True:
        # Check MQTT messages
        client.check_msg()

        utime.sleep_ms(100)


if __name__ == "__main__":
    log(None, "Sleeping for 1 second...")
    utime.sleep(1)  # make we could exit script by Ctrl+C

    try:
        main()
    except Exception as e:
        print_exception(e)
        print("Error occurred: {}".format(str(e)))
        reset()
