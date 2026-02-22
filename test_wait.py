import requests
import json
import threading
import time

def wait_thread():
    # 1. SSE接続を開始して sessionId を取得
    r = requests.get('http://localhost:8080/sse', stream=True)
    # 最初の行に sessionId が含まれているはず
    session_id = ""
    for line in r.iter_lines():
        if line:
            session_id = line.decode('utf-8')
            break
    
    print(f"Session ID acquired: {session_id}")
    
    # 2. その sessionId を使って wait_notify を呼び出す
    payload = {
        "jsonrpc": "2.0",
        "method": "tools/call",
        "params": {
            "name": "wait_notify",
            "arguments": {"agent_id": "Gemini-Automated-Tester", "timeout_sec": 30}
        },
        "id": "wait_test"
    }
    print("Calling wait_notify and waiting for your post...")
    resp = requests.post(f"http://localhost:8080/message?sessionId={session_id}", json=payload)
    print("\n--- TEST RESULT ---")
    print(resp.text)
    print("-------------------\n")

threading.Thread(target=wait_thread).start()
time.sleep(2) # 接続待ち
