#!/usr/bin/env python3
"""
PLC Coil Monitor
================
Connects to the Go arena WebSocket server and:
  - Prints all coil values whenever a plcIoChange event arrives
  - Keys 1–8 each have their own internal counter mapped to Registers 3–10
  - Register 1  = sum of Registers 3, 4, 5, 6
  - Register 2  = sum of Registers 7, 8, 9, 10
  - When Coil 1 goes FALSE, all registers (1–2, 3–10) are cleared to 0

Key → Register map
------------------
    Key 1 → Register 3    Key 5 → Register 7
    Key 2 → Register 4    Key 6 → Register 8
    Key 3 → Register 5    Key 7 → Register 9
    Key 4 → Register 6    Key 8 → Register 10

Totals
------
    Register 1 = Reg 3 + Reg 4 + Reg 5 + Reg 6
    Register 2 = Reg 7 + Reg 8 + Reg 9 + Reg 10

Usage
-----
    python plc_coil_monitor.py [--host HOST] [--port PORT]

Defaults: host=10.0.100.5, port=8080

Controls
--------
    1–8   Increment that key's counter
    q     Quit
"""

import asyncio
import json
import argparse
import logging
import signal
import sys
import threading
from typing import Any
from datetime import datetime

import websockets

# ── Constants ──────────────────────────────────────────────────────────────────

# Maps keyboard digit → PLC register
KEY_REGISTER_MAP: dict[str, int] = {
    "1": 3,
    "2": 4,
    "3": 5,
    "4": 6,
    "5": 7,
    "6": 8,
    "7": 9,
    "8": 10,
}

TOTAL_1_REGISTER   = 1             # Sum of registers 3,4,5,6
TOTAL_2_REGISTER   = 2             # Sum of registers 7,8,9,10
GROUP_A_REGISTERS  = [3, 4, 5, 6]
GROUP_B_REGISTERS  = [7, 8, 9, 10]
ALL_DATA_REGISTERS = GROUP_A_REGISTERS + GROUP_B_REGISTERS

# ── Logging ────────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%H:%M:%S",
)
log = logging.getLogger("plc_monitor")


# ── Message helpers ────────────────────────────────────────────────────────────

def encode_ws_message(message_type: str, data: Any) -> str:
    return json.dumps({"type": message_type, "data": data})


def decode_ws_message(raw: str) -> tuple[str, Any]:
    msg = json.loads(raw)
    if not isinstance(msg, dict):
        raise ValueError(f"Expected JSON object, got: {type(msg)}")
    return msg.get("type", ""), msg.get("data")


# ── PLC Monitor Client ─────────────────────────────────────────────────────────

class PLCMonitor:
    def __init__(self, host: str = "10.0.100.5", port: int = 8080):
        self.ws_url   = f"ws://{host}:{port}/api/plc/websocket"
        self._ws      = None
        self._running = False

        # Coil snapshot for edge detection
        self.coil_state: dict = {}

        # One counter per key (keys "1"–"8")
        self.counters: dict[str, int] = {k: 0 for k in KEY_REGISTER_MAP}

        # Register value cache
        self.reg_values: dict[int, int] = {r: 0 for r in ALL_DATA_REGISTERS}

    # ── WebSocket send helpers ─────────────────────────────────────────────────

    async def set_registers(self, registers: list[dict]) -> None:
        if self._ws is None:
            log.warning("WS not connected; cannot set registers")
            return
        msg = encode_ws_message("setRegisters", registers)
        await self._ws.send(msg)
        log.debug("WS → setRegisters  count=%d", len(registers))

    # ── Register logic ─────────────────────────────────────────────────────────

    def _compute_totals(self) -> tuple[int, int]:
        total1 = sum(self.reg_values.get(r, 0) for r in GROUP_A_REGISTERS)
        total2 = sum(self.reg_values.get(r, 0) for r in GROUP_B_REGISTERS)
        return total1, total2

    async def push_registers(self) -> None:
        """Write all data registers + both totals to the PLC in one batch."""
        total1, total2 = self._compute_totals()

        batch = [{"register": r, "cValue": self.reg_values.get(r, 0)} for r in ALL_DATA_REGISTERS]
        batch.append({"register": TOTAL_1_REGISTER, "cValue": total1})
        batch.append({"register": TOTAL_2_REGISTER, "cValue": total2})

        await self.set_registers(batch)

        log.info(
            "Regs → "
            "3=%d  4=%d  5=%d  6=%d  | Reg1(A-total)=%d  ||  "
            "7=%d  8=%d  9=%d  10=%d | Reg2(B-total)=%d",
            self.reg_values.get(3,  0),
            self.reg_values.get(4,  0),
            self.reg_values.get(5,  0),
            self.reg_values.get(6,  0),
            total1,
            self.reg_values.get(7,  0),
            self.reg_values.get(8,  0),
            self.reg_values.get(9,  0),
            self.reg_values.get(10, 0),
            total2,
        )

    async def reset_all_registers(self) -> None:
        """Zero out all internal counters and write zeros to the PLC."""
        log.info("Resetting all registers to 0 …")
        self.counters   = {k: 0 for k in KEY_REGISTER_MAP}
        self.reg_values = {r: 0 for r in ALL_DATA_REGISTERS}
        {a: 0 for a in GROUP_A_REGISTERS}
        {b: 0 for b in GROUP_B_REGISTERS}

        batch = [{"register": r, "cValue": 0} for r in ALL_DATA_REGISTERS]
        batch.append({"register": TOTAL_1_REGISTER, "cValue": 0})
        batch.append({"register": TOTAL_2_REGISTER, "cValue": 0})

        try:
            await self.set_registers(batch)
            log.info("🔄 Reset complete — all registers cleared to 0")
        except Exception as exc:
            log.error("Reset failed: %s", exc)

    # ── Key handler ────────────────────────────────────────────────────────────

    async def handle_key(self, key: str) -> None:
        """Increment the counter for the given key, update its register, push totals."""
        reg = KEY_REGISTER_MAP[key]
        self.counters[key] += 1
        self.reg_values[reg] = self.counters[key]

        total1, total2 = self._compute_totals()
        log.info(
            "Key [%s] pressed → counter=%d  Reg%d=%d  | Reg1=%d  Reg2=%d",
            key,
            self.counters[key],
            reg,
            self.reg_values[reg],
            total1,
            total2,
        )
        await self.push_registers()

    # ── Coil update ────────────────────────────────────────────────────────────

    def _print_coils(self, coil_list: list) -> None:
        ts = datetime.now().strftime("%H:%M:%S.%f")[:-3]
        coil_strs = "  ".join(
            f"Coil[{i}]={'ON ' if v else 'OFF'}"
            for i, v in enumerate(coil_list)
        )
        print(f"\n[{ts}] COILS → {coil_strs}")

    def on_coil_update(self, coils: dict) -> None:
        coil_list = coils.get("Coils", [])

        # Print every coil value on any change
        if isinstance(coil_list, list):
            self._print_coils(coil_list)

        # Watch Coil 1 for a rising edge (False → True) → reset
        if isinstance(coil_list, list) and len(coil_list) > 1:
            current_coil1 = coil_list[1]
            prev_coils    = self.coil_state.get("Coils", [])
            prev_coil1    = prev_coils[1] if isinstance(prev_coils, list) and len(prev_coils) > 1 else False

            if current_coil1 is True and prev_coil1 is False:
                log.info("⚠️  Coil[1] went TRUE → resetting all registers")
                try:
                    loop = asyncio.get_running_loop()
                    loop.create_task(self.reset_all_registers())
                except RuntimeError:
                    log.error("Could not schedule reset — no running event loop")

        self.coil_state = coils

    # ── Inbound message dispatch ───────────────────────────────────────────────

    def on_server_message(self, msg_type: str, data: Any) -> None:
        if msg_type == "plcIoChange":
            coils = data.get("coils", data) if isinstance(data, dict) else {}
            self.on_coil_update(coils)
        elif msg_type == "plcRegisterSetSuccess":
            log.debug("Server ack (registers): %s", data)
        elif msg_type == "plcInputSetSuccess":
            log.debug("Server ack (inputs): %s", data)
        elif msg_type == "error":
            log.warning("Server error: %s", data)
        else:
            log.debug("Unhandled '%s': %s", msg_type, data)

    # ── WebSocket connection ───────────────────────────────────────────────────

    async def connect(self) -> None:
        log.info("Connecting to %s …", self.ws_url)
        async with websockets.connect(
            self.ws_url,
            ping_interval=20,
            ping_timeout=30,
        ) as ws:
            self._ws      = ws
            self._running = True
            log.info("Connected. Press keys 1–8 to increment counters. Press 'q' to quit.")
            async for raw in ws:
                if not self._running:
                    break
                try:
                    msg_type, data = decode_ws_message(raw)
                    self.on_server_message(msg_type, data)
                except (ValueError, json.JSONDecodeError) as exc:
                    log.warning("Bad message: %s — %s", exc, raw[:120])
        self._ws      = None
        self._running = False
        log.info("Disconnected.")

    async def run_forever(self, reconnect_delay: float = 3.0) -> None:
        while True:
            try:
                await self.connect()
            except (OSError, websockets.exceptions.WebSocketException) as exc:
                log.warning("Connection lost: %s. Retrying in %.0fs …", exc, reconnect_delay)
                await asyncio.sleep(reconnect_delay)

    def stop(self) -> None:
        self._running = False
        if self._ws:
            asyncio.ensure_future(self._ws.close())


# ── Keyboard listener ──────────────────────────────────────────────────────────

def keyboard_listener(client: PLCMonitor, loop: asyncio.AbstractEventLoop) -> None:
    """
    Reads keypresses and dispatches to handle_key().
    With 'readchar' installed (pip install readchar): true single-keypress,
    no ENTER needed.  Without it: line-buffered fallback (type digit + ENTER).
    """
    try:
        import readchar
        _USE_READCHAR = True
        print(
            "Keyboard ready (single-key mode). "
            "Press 1–8 to increment, 'q' to quit."
        )
    except ImportError:
        _USE_READCHAR = False
        print(
            "Keyboard ready (line-buffered mode — type digit then ENTER).\n"
            "Tip: pip install readchar for instant single-keypress detection.\n"
            "Press 1–8 + ENTER to increment, 'q' + ENTER to quit."
        )

    while True:
        try:
            if _USE_READCHAR:
                key = readchar.readkey()
            else:
                line = sys.stdin.readline()
                if line is None:
                    break
                key = line.strip()

            if key.lower() == "q":
                log.info("Quit requested.")
                client.stop()
                for task in asyncio.all_tasks(loop):
                    task.cancel()
                break
            elif key in KEY_REGISTER_MAP:
                asyncio.run_coroutine_threadsafe(client.handle_key(key), loop)
            else:
                if key:  # suppress blank lines in line-buffered mode
                    print(f"  (unknown key '{key}' — use 1–8 or q)")

        except (EOFError, KeyboardInterrupt):
            break


# ── Main ───────────────────────────────────────────────────────────────────────

async def main(host: str, port: int) -> None:
    client = PLCMonitor(host=host, port=port)
    loop   = asyncio.get_running_loop()

    def _handle_signal(*_):
        log.info("Shutting down …")
        client.stop()
        for task in asyncio.all_tasks(loop):
            task.cancel()

    signal.signal(signal.SIGINT,  _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    kb_thread = threading.Thread(
        target=keyboard_listener,
        args=(client, loop),
        daemon=True,
        name="keyboard-listener",
    )
    kb_thread.start()

    await client.run_forever()


# ── Entry point ────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="PLC Coil Monitor — 8-key counter")
    #parser.add_argument("--host", default="10.0.100.5", help="Arena server host (default: 10.0.100.5)")
    parser.add_argument("--host", default="localhost", help="Arena server host (default: 10.0.100.5)")
    parser.add_argument("--port", default=8080, type=int,  help="Arena server port (default: 8080)")
    args = parser.parse_args()

    try:
        asyncio.run(main(args.host, args.port))
    except (KeyboardInterrupt, asyncio.CancelledError):
        sys.exit(0)
