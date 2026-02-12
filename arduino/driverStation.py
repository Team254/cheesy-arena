#cat << 'EOF' > /home/jetson/modbus_plc.py
#!/usr/bin/env python3
import time
import threading
import Jetson.GPIO as GPIO
from pymodbus.server.sync import StartTcpServer
from pymodbus.datastore import ModbusSequentialDataBlock, ModbusSlaveContext, ModbusServerContext
import sys

# --- 硬體腳位設定 (BOARD 模式) ---
# 依序對應: RE1, RA1, RE2, RA2, RE3, RA3
INPUT_PINS = [29, 31, 32, 33, 35, 37]

# 輸出腳位
#OUTPUT_PIN = 37

# --- GPIO 初始化 ---
GPIO.setmode(GPIO.BOARD)
GPIO.setwarnings(True)

# 啟用內部上拉 (PUD_UP)
# 平常懸空 = 1 (High)
# 接地觸發 = 0 (Low)
GPIO.setup(INPUT_PINS, GPIO.IN, pull_up_down=GPIO.PUD_UP) 
#GPIO.setup(OUTPUT_PIN, GPIO.OUT, initial=GPIO.LOW)

# --- Modbus 記憶體 ---
store = ModbusSlaveContext(
    co = ModbusSequentialDataBlock(0, [0]*100),
    di = ModbusSequentialDataBlock(0, [0]*100),
)
context = ModbusServerContext(slaves=store, single=True)

# --- 背景任務 (高速掃描) ---
def background_loop():
    print("Background thread started...", flush=True)
    while True:
        try:
            input_status = []
            # 掃描 6 顆按鈕
            for i, pin in enumerate(INPUT_PINS):
                # 0觸發邏輯 (Active Low)
                # 實體接地(0) -> Modbus變1
                # 實體懸空(1) -> Modbus變0
                raw_val = GPIO.input(pin)
                val = 1 if raw_val == 0 else 0
                
                store.setValues(1, i, [val])
                input_status.append(str(val))
            
            # 控制輸出 (Coil 16)
            #target = store.getValues(1, 16, count=1)
            # if target[0] == 1:
            #    GPIO.output(OUTPUT_PIN, GPIO.HIGH)
            #else:
            #    GPIO.output(OUTPUT_PIN, GPIO.LOW)
            
            # 掃描頻率 0.05秒
            print(f"In: {input_status}" , flush=True)
            time.sleep(0.05)
            
        except Exception as e:
            print(f"Error: {e}", flush=True)

if __name__ == "__main__":
    print("------------------------------------------------", flush=True)
    print(" Jetson PLC Running: 6 Buttons Mode", flush=True)
    print(" Pins: 29, 31, 32, 33, 35, 37", flush=True)
    print("------------------------------------------------", flush=True)
    
    t = threading.Thread(target=background_loop)
    t.daemon = True
    t.start()
    
    StartTcpServer(context, address=("0.0.0.0", 502))
    GPIO.cleanup()
EOF