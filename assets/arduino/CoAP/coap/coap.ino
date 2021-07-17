#include <ESP8266WiFi.h>
#include "coap_client.h"
#include <ArduinoJson.h>
#include <SPI.h>
#include <MFRC522.h>

#define pinRST D0
#define pinSS D8
#define sensorPin A0

MFRC522 rfid(pinSS, pinRST);
MFRC522::MIFARE_Key key;

const char* ssid     = "ssid";
const char* password = "password";

IPAddress ip(192, 168, 2, 227);
int port = 5683;
String DEVICE_SECRET_KEY  = "esp8266";
char* path = "bath";


#define BLINKLED D1
#define TEMPLED D2
#define LEVELLED D3

WiFiClient espClient;
coapClient coap;

unsigned int temp = 0, level, favLevel = 0;

bool blinkLED = false;
byte nuidPICC[4];

long long timeNow = 0;

void setup() {
  // put your setup code here, to run once:
  Serial.begin(115200);
  delay(2000);
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

  coap.response(callback_response);
  coap.start();

  pinMode(BLINKLED, OUTPUT);
  pinMode(TEMPLED, OUTPUT);
  pinMode(LEVELLED, OUTPUT);

  SPI.begin(); // Init SPI bus
  Serial.println("started spi");
  rfid.PCD_Init();
  Serial.println("starting loop");
}

void loop() {
  level = analogRead(sensorPin);

  bool state = coap.loop();
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

void callback_response(coapPacket &packet, IPAddress ip, int port) {
  char p[packet.payloadlen + 1];
  memcpy(p, packet.payload, packet.payloadlen);
  p[packet.payloadlen] = NULL;
  Serial.println(p);

  //response from coap server
  if (packet.type == 3 && packet.code == 0) {
    Serial.println("ping ok");
  }

  StaticJsonDocument<200> user;
  DeserializationError error = deserializeJson(user, p);
  if (error) {
    Serial.print(F("deserializeJson() failed: "));
    //    Serial.println(error.f_str());
    return;
  }

  temp = user["water_temp"];
  Serial.print("water temp: ");
  Serial.println(temp);

  favLevel = user["water_level"];
  Serial.print("water level: ");
  Serial.println(favLevel);
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

  char msg[content.length() + 1];
  Serial.println("Message: " + content);
  content.toCharArray(msg, content.length() + 1);
  int msgid = coap.post(ip, port, path, msg, content.length());

  // Halt PICC
  rfid.PICC_HaltA();
  // Stop encryption on PCD
  rfid.PCD_StopCrypto1();
}
