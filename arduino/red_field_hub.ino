#include <SPI.h>
#include <Ethernet.h>
#include <Mudbus.h>

// 網路設定
byte mac[] = { 0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xED }; 
IPAddress ip(10, 0, 100, 45);
IPAddress gateway(10, 0, 100, 1);
IPAddress subnet(255, 255, 255, 0);

Mudbus Mb;

// 實體腳位設定 (只設定這 5 個)
const int PIN_SENSORS[] = {2, 3, 5, 6}; // 輸入
const int PIN_OUTPUT = 7;               // 輸出

void setup() {
  Serial.begin(115200);
  
  // W5100 硬體修正
  pinMode(10, OUTPUT);
  digitalWrite(10, HIGH); 
  delay(500);

  Ethernet.begin(mac, ip, gateway, gateway, subnet);
  Serial.print("IP: "); Serial.println(Ethernet.localIP());

  // IO 初始化
  for (int i = 0; i < 4; i++) {
    pinMode(PIN_SENSORS[i], INPUT_PULLUP);
  }
  pinMode(PIN_OUTPUT, OUTPUT);
}

void loop() {
  Mb.Run();

  // ============================================
  // 區域 1：實體輸入 (Arduino -> PLC)
  // 對應 Modbus 位址 00001 ~ 00004
  // ============================================
  for (int i = 0; i < 4; i++) {
    // 讀取光閘，並填入前 4 個 Coil
    Mb.C[i] = !digitalRead(PIN_SENSORS[i]); 
  }

  // ============================================
  // 區域 2：實體輸出 (PLC -> Arduino)
  // 對應 Modbus 位址 00005
  // ============================================
  bool plc_command = Mb.C[16];
  if (plc_command) {
    digitalWrite(PIN_OUTPUT, HIGH);
  } else {
    digitalWrite(PIN_OUTPUT, LOW);
  }

  // ============================================
  // 區域 3：虛擬/預留區 (PLC <-> Arduino)
  // 對應 Modbus 位址 00006 ~ 00016 (即 Mb.C[5] ~ Mb.C[15])
  // ============================================
  // 這裡沒寫程式碼，代表這些位址是「空接」的。
  // PLC 可以寫入 1 或 0 進來，數值會存在 Arduino 記憶體裡，
  // 但不會觸發任何硬體動作。這樣就湊滿 16 個了。
  
  // (選用) 如果你想監看 PLC 有沒有偷改第 16 個 Coil:
  // if (Mb.C[15] == 1) { Serial.println("PLC Triggered Virtual Coil 16!"); }
}