#include <ESP8266WiFi.h>
#include <Ethernet.h>
#include <PubSubClient.h>
#include <SPI.h>
#include <MFRC522.h>
#define pinRST D0
#define pinSS D8
#define sensorPin A0

MFRC522 rfid(pinSS, pinRST);
MFRC522::MIFARE_Key key;

const char* ssid     = "ssid";
const char* password = "password";

const char* tempTopic = "smart-home/bath/temp";
const char* levelTopic = "smart-home/bath/level";

IPAddress server(192, 168, 2, 227);
const char* clientID = "ESP8266Client";

#define BLINKLED D1
#define TEMPLED D2
#define LEVELLED D3

WiFiClient espClient;
PubSubClient mqttClient(espClient);

unsigned int temp = 0, level, favLevel = 0;

bool blinkLED = false;
byte nuidPICC[4];

long long timeNow = 0;

void setup() {
  // put your setup code here, to run once:
  Serial.begin(115200);
  delay(1000);
  Serial.println();

  WiFi.mode(WIFI_STA);
  WiFi.disconnect();
  Serial.println(WiFi.macAddress());

  delay(1000);

  if (WiFi.status() != WL_CONNECTED)
    WiFi.begin(ssid, password);

  Serial.println();
  Serial.print("Connecting");
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.print(".");
  }
  Serial.println();
  Serial.println("Connected");

  Serial.print("IP address:\t");
  Serial.println(WiFi.localIP());

  mqttClient.setServer(server, 1883);
  mqttClient.setCallback(callback);

  pinMode(BLINKLED, OUTPUT);
  pinMode(TEMPLED, OUTPUT);
  pinMode(LEVELLED, OUTPUT);

  SPI.begin(); // Init SPI bus
  Serial.println("started spi");
  rfid.PCD_Init();
  Serial.println("starting loop");
}

void loop() {
  if (!mqttClient.connected()) {
    reconnect();
  }

  level = analogRead(sensorPin);

  mqttClient.loop();
  adjustLEDLights();

  if (!rfid.PICC_IsNewCardPresent() || !rfid.PICC_ReadCardSerial())
    return;

  newCard();
}

void adjustLEDLights() {
  if (millis() - timeNow > 1000 || millis() - timeNow < 0) {
    timeNow = millis();
    Serial.print(level);
    Serial.print(" ==> ");
    Serial.println(map(level, 0, 1023, 0, 100));

    if (abs(map(level, 0, 1023, 0, 100) - favLevel) < 5) {
      if (blinkLED)
        Serial.println("water level has been set!!!");
      blinkLED = false;
    } else {
      blinkLED = !blinkLED;
    }

    if (blinkLED)
      digitalWrite(BLINKLED, HIGH);
    else
      digitalWrite(BLINKLED, LOW);
  }

  analogWrite(TEMPLED, map(temp, 0, 100, 0, 1023));
  analogWrite(LEVELLED, level);
}

void reconnect() {
  Serial.print("MQTT");
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.print("Attempting MQTT connection...");
    // Attempt to connect
    if (mqttClient.connect(clientID)) {
      Serial.println("connected");
      mqttClient.subscribe("smart-home/bath/+");
    } else {
      Serial.print("failed, rc=");
      Serial.print(mqttClient.state());

      Serial.println(" try again in 5 seconds");
      delay(5000);
    }
  }
}

void callback(char* topic, byte* payload, unsigned int length) {
  unsigned int value = (char)payload[0];

  if (strcmp(topic, tempTopic) == 0) {
    temp = value;
    Serial.print("water temp: ");
    Serial.println(value);
  }
  else {
    favLevel = value;
    Serial.print("water level: ");
    Serial.println(value);
  }
}

void newCard() {
  Serial.print(F("PICC type: "));
  MFRC522::PICC_Type piccType = rfid.PICC_GetType(rfid.uid.sak);
  Serial.println(rfid.PICC_GetTypeName(piccType));

  if (rfid.uid.uidByte[0] != nuidPICC[0] ||
      rfid.uid.uidByte[1] != nuidPICC[1] ||
      rfid.uid.uidByte[2] != nuidPICC[2] ||
      rfid.uid.uidByte[3] != nuidPICC[3] ) {
    Serial.println(F("A new card has been detected."));

    for (byte i = 0; i < 4; i++) {
      nuidPICC[i] = rfid.uid.uidByte[i];
    }
  }
  else Serial.println(F("Card read previously."));

  Serial.print("UID tag :");
  String content = "";
  byte letter;
  for (byte i = 0; i < rfid.uid.size; i++) {
    Serial.print(rfid.uid.uidByte[i] < 0x10 ? " 0" : " ");
    Serial.print(rfid.uid.uidByte[i], HEX);
    content.concat(String(rfid.uid.uidByte[i] < 0x10 ? " 0" : " "));
    content.concat(String(rfid.uid.uidByte[i], HEX));
  }
  Serial.println();
  Serial.println("Message: " + content);
  mqttClient.publish("smart-home/id", nuidPICC, 4, true);

  // Halt PICC
  rfid.PICC_HaltA();
  // Stop encryption on PCD
  rfid.PCD_StopCrypto1();
}
