export class RadarWebSocket {
    constructor(url, onMessage, onPingCallback) {
        this.url = url;
        this.onMessage = onMessage;
        this.onPingCallback = onPingCallback;
        this.ws = null;
        this.pingInterval = null;
        this.pendingPing = false;
        this.pingStartTime = 0;
        this.connect();
    }

    connect() {
        this.ws = new WebSocket(this.url);
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.setStatus(true);
            this.sendGetHistory(30);
            if (this.pingInterval) clearInterval(this.pingInterval);
            this.pingInterval = setInterval(() => this.measurePing(), 2000);
        };
        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.setStatus(false);
            if (this.pingInterval) clearInterval(this.pingInterval);
            setTimeout(() => this.connect(), 3000);
        };
        this.ws.onerror = (err) => console.error('WS error', err);
        this.ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                if (msg.type === 'command_response' && msg.payload && msg.payload.result === 'pong') {
                    if (this.pendingPing && this.pingStartTime) {
                        const latency = Date.now() - this.pingStartTime;
                        if (this.onPingCallback) this.onPingCallback(latency);
                        this.pendingPing = false;
                    }
                } else {
                    this.onMessage(msg);
                }
            } catch(e) { console.error(e); }
        };
    }

    measurePing() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN && !this.pendingPing) {
            this.pendingPing = true;
            this.pingStartTime = Date.now();
            this.sendCommand('ping', {});
            setTimeout(() => {
                if (this.pendingPing) {
                    if (this.onPingCallback) this.onPingCallback(null);
                    this.pendingPing = false;
                }
            }, 1000);
        }
    }

    sendGetHistory(limit) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'get_history',
                payload: { limit: limit }
            }));
        }
    }

    sendCommand(command, args) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'command',
                payload: { command, args }
            }));
        }
    }

    setStatus(online) {
        const led = document.getElementById('led');
        const connText = document.getElementById('connText');
        if (online) {
            led.className = 'led online';
            connText.innerText = 'СОЕДИНЕНИЕ: АКТИВНО';
        } else {
            led.className = 'led offline';
            connText.innerText = 'СОЕДИНЕНИЕ: ПОТЕРЯНО';
        }
    }
}