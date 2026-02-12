#include <SPI.h>
#include <Ethernet.h>
#include <Mudbus.h> // 請確認已安裝此庫，若無可使用其它 ModbusTCP 庫

// --- 1. 網路設定 ---
byte mac[] = { 0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xED }; // MAC 地址 (隨意，不要跟網域內重複即可)
IPAddress ip(10, 0, 100, 45);                        // Arduino IP
IPAddress gateway(10, 0, 100, 1);                    // Gateway
IPAddress subnet(255, 255, 255, 0);                  // Subnet Mask

// 建立 Modbus 物件
Mudbus Mb;

// --- 2. 腳位定義 ---
const int PIN_GATE_1 = 2;
const int PIN_GATE_2 = 3;
const int PIN_GATE_3 = 5; // 跳過 Pin 4 (SD Card)
const int PIN_GATE_4 = 6;
const int PIN_OUTPUT = 7;

void setup() {
  // 初始化序列埠 (除錯用)
  Serial.begin(9600);
  
  // 初始化網路
  Ethernet.begin(mac, ip, gateway, gateway, subnet); // W5100 初始化
  
  // 等待網路啟動
  delay(1000);
  Serial.print("Arduino Modbus Slave is ready at: ");
  Serial.println(Ethernet.localIP());

  // --- 3. 初始化 IO ---
  // 設定光閘為輸入，並啟用內部上拉電阻 (Input Pullup)
  // 這樣光閘沒動作時是 HIGH，遮斷/導通到地時是 LOW (視您的光閘類型而定)
  pinMode(PIN_GATE_1, INPUT_PULLUP);
  pinMode(PIN_GATE_2, INPUT_PULLUP);
  pinMode(PIN_GATE_3, INPUT_PULLUP);
  pinMode(PIN_GATE_4, INPUT_PULLUP);

  // 設定輸出
  pinMode(PIN_OUTPUT, OUTPUT);
  digitalWrite(PIN_OUTPUT, LOW); // 預設關閉
}

void loop() {
  // 處理 Modbus 通訊
  Mb.Run();

  // --- 4. 讀取光閘並寫入暫存器 (Register 40001 / Mb.R[0]) ---
  // 我們使用 "Bit Packing" 技術，將 4 個狀態塞進一個 16-bit 整數
  // 讀取狀態 (假設光閘被遮擋時為 LOW，這裡反轉為 1 代表觸發，視您的感測器邏輯調整 !)
  int g1 = !digitalRead(PIN_GATE_1); 
  int g2 = !digitalRead(PIN_GATE_2);
  int g3 = !digitalRead(PIN_GATE_3);
  int g4 = !digitalRead(PIN_GATE_4);

  // 組合數值: 
  // g1 在第 0 位 (值=1)
  // g2 在第 1 位 (值=2)
  // g3 在第 2 位 (值=4)
  // g4 在第 3 位 (值=8)
  uint16_t packedStatus = (g1) | (g2 << 1) | (g3 << 2) | (g4 << 3);

  // 將組合好的數值放入 Modbus Holding Register 0 (對應 40001)
  Mb.R[0] = packedStatus;

  // --- 5. 讀取 PLC 指令並控制輸出 (Coil 00001 / Mb.C[0]) ---
  // 讀取 Modbus Coil 0 的狀態
  bool outputState = Mb.C[0];
  
  // 控制實體腳位
  digitalWrite(PIN_OUTPUT, outputState);

  // --- (選用) 序列埠除錯顯示 ---
  // 為了避免洗版，您可以每秒印一次，或者只在數值改變時印
  /*
  static unsigned long lastPrint = 0;
  if (millis() - lastPrint > 1000) {
    lastPrint = millis();
    Serial.print("Reg[0]: "); Serial.print(packedStatus, BIN);
    Serial.print(" | Coil[0]: "); Serial.println(outputState);
  }
  */
}