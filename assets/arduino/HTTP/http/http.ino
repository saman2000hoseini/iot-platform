#include <ESP8266WiFi.h>
#include <ESP8266HTTPClient.h>
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

String endpoint = "http://192.168.2.227:65432/bath";


#define BLINKLED D1
#define TEMPLED D2
#define LEVELLED D3

WiFiClient espClient;

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

void httpCall(String content) {
  WiFiClient client;
  HTTPClient http;

  http.useHTTP10(true);
  http.begin(client, endpoint);

  http.addHeader("Content-Type", "application/x-www-form-urlencoded");

  int httpResponseCode = http.POST("id=" + content);

  if (httpResponseCode > 0) {
    Serial.print("HTTP Response code: ");
    Serial.println(httpResponseCode);

    if (httpResponseCode == 200) {
      StaticJsonDocument<200> user;
      DeserializationError error = deserializeJson(user, http.getString());
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
  }
  else {
    Serial.print("Error code: ");
    Serial.println(httpResponseCode);
  }
  // Free resources
  http.end();
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

  httpCall(content);

  // Halt PICC
  rfid.PICC_HaltA();
  // Stop encryption on PCD
  rfid.PCD_StopCrypto1();
}
