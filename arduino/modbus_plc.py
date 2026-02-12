#cat << 'EOF' > /home/jetson/modbus_plc.py
#!/usr/bin/env python3
import time
import threading
import Jetson.GPIO as GPIO
from pymodbus.server.sync import StartTcpServer
from pymodbus.datastore import ModbusSequentialDataBlock, ModbusSlaveContext, ModbusServerContext
import sys

# --- 1. 硬體腳位設定 (BOARD 模式) ---
# 輸入: 4個 inputs (對應 Modbus Coil 0~3)
INPUT_PINS = [27, 29, 31, 33]

# 輸出: 2個 outputs (對應 Modbus Coil 16, 17)
# Index 0 -> Pin 35
# Index 1 -> Pin 37
OUTPUT_PINS = [35, 37]

# --- 2. GPIO 初始化 ---
GPIO.setmode(GPIO.BOARD)
GPIO.setwarnings(False)

# 輸入: 啟用內部上拉 (平常=1, 接地觸發=0)
GPIO.setup(INPUT_PINS, GPIO.IN, pull_up_down=GPIO.PUD_UP)
# 輸出: 預設全關 (LOW)
GPIO.setup(OUTPUT_PINS, GPIO.OUT, initial=GPIO.LOW)

# --- 3. Modbus 記憶體 ---
store = ModbusSlaveContext(
    co = ModbusSequentialDataBlock(0, [0]*100),
    di = ModbusSequentialDataBlock(0, [0]*100),
)
context = ModbusServerContext(slaves=store, single=True)

# --- 4. 背景任務 (處理 IO) ---
def background_loop():
    print("Background thread started...", flush=True)
    while True:
        try:
            # --- A. 處理輸入 (4 Inputs) ---
            input_status = []
            for i, pin in enumerate(INPUT_PINS):
                # 0觸發邏輯: 實體接地(0) -> Modbus 顯示 1
                raw_val = GPIO.input(pin)
                val = 1 if raw_val == 0 else 0
                store.setValues(1, i, [val])
                input_status.append(str(val))
            
            # --- B. 處理輸出 (2 Outputs) ---
            # 讀取 Modbus Coil 16 和 17
            # Coil 16 -> 控制 Pin 35
            # Coil 17 -> 控制 Pin 37
            output_status = []
            target_states = store.getValues(1, 16, count=len(OUTPUT_PINS)) # 一次讀取2個狀態
            
            for i, pin in enumerate(OUTPUT_PINS):
                state = target_states[i] # 取得對應的 Modbus 狀態
                if state == 1:
                    GPIO.output(pin, GPIO.HIGH)
                    output_status.append("ON")
                else:
                    GPIO.output(pin, GPIO.LOW)
                    output_status.append("OFF")
            
            # --- C. 印出狀態 ---
            # 每 1 秒印一次就好，比較不會洗版
            print(f"In: {input_status} | Out(35,37): {output_status}", flush=True)
            time.sleep(1)
            
        except Exception as e:
            print(f"Error: {e}", flush=True)

# --- 5. 主程式 ---
if __name__ == "__main__":
    print("------------------------------------------------", flush=True)
    print(" Jetson Modbus PLC (2 Outputs Version)", flush=True)
    print(" Inputs : 27, 29, 31, 33 (Active Low)", flush=True)
    print(" Outputs: 35 (Coil 16), 37 (Coil 17)", flush=True)
    print("------------------------------------------------", flush=True)
    
    # 啟動背景執行緒 (修正原本 LoopingCall 會卡住的問題)
    t = threading.Thread(target=background_loop)
    t.daemon = True
    t.start()
    
    # 啟動伺服器
    StartTcpServer(context, address=("0.0.0.0", 502))
    GPIO.cleanup()
EOF