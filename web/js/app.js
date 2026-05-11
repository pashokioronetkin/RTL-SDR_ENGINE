import { Radar } from './radar.js';
import { RadarWebSocket } from './websocket.js';

const radar = new Radar('radarCanvas');
let allDetections = [];

const lastUpdateSpan = document.getElementById('lastUpdate');
const pingValueSpan = document.getElementById('pingValue');

// Обновляем отображение текущего времени раз в секунду
setInterval(() => {
    lastUpdateSpan.innerText = new Date().toLocaleTimeString();
}, 1000);

function onPingReceived(latencyMs) {
    if (latencyMs !== null && !isNaN(latencyMs)) {
        pingValueSpan.innerText = latencyMs;
    } else {
        pingValueSpan.innerText = '--';
    }
}

function updateUI() {
    const totalSpan = document.getElementById('totalTargets');
    const avgSpan = document.getElementById('avgConfidence');
    const reliableSpan = document.getElementById('reliableRate');
    const typeDistDiv = document.getElementById('typeDist');
    const tbody = document.getElementById('targetsBody');

    totalSpan.innerText = allDetections.length;
    let totalConf = 0, reliableCnt = 0;
    let typeCount = {};
    for (let d of allDetections) {
        totalConf += d.object_info.confidence;
        if (d.object_info.is_reliable) reliableCnt++;
        let t = d.object_info.type;
        typeCount[t] = (typeCount[t] || 0) + 1;
    }
    let avgConf = allDetections.length ? (totalConf / allDetections.length * 100).toFixed(1) : '0.0';
    let reliablePct = allDetections.length ? (reliableCnt / allDetections.length * 100).toFixed(1) : '0.0';
    avgSpan.innerText = avgConf + '%';
    reliableSpan.innerText = reliablePct + '%';
    lastUpdateSpan.innerText = new Date().toLocaleTimeString();

    if (Object.keys(typeCount).length === 0) {
        typeDistDiv.innerHTML = '— НЕТ ДАННЫХ —';
    } else {
        let html = '';
        for (let [type, cnt] of Object.entries(typeCount)) {
            html += `<div class="type-row"><span>${type}</span><span><b>${cnt}</b></span></div>`;
        }
        typeDistDiv.innerHTML = html;
    }

    let rows = '';
    const last15 = [...allDetections].reverse().slice(0, 15);
    for (let d of last15) {
        const timeStr = new Date(d.timestamp).toLocaleString();
        rows += `<tr>
            <td>${timeStr}</td>
            <td><b>${d.object_info.type}</b> (${(d.object_info.confidence*100).toFixed(0)}%)</td>
            <td>${d.location.azimuth.toFixed(1)}°</td>
            <td>${(d.location.range/1000).toFixed(1)}</td>
            <td>${d.location.altitude}</td>
            <td>${d.risk_assessment}</td>
        </tr>`;
    }
    tbody.innerHTML = rows || '<tr><td colspan="6">Нет данных</td></tr>';
}

function addDetectionData(data) {
    if (!data.detections) return;
    let changed = false;
    for (let det of data.detections) {
        const existing = allDetections.find(d => d.detection_id === det.detection_id);
        if (existing) {
            Object.assign(existing, det);
        } else {
            allDetections.push(det);
        }
        radar.addDetection(det);
        changed = true;
    }
    if (changed) {
        updateUI();
        radar.draw();
        lastUpdateSpan.innerText = new Date().toLocaleTimeString();
    }
}

function onWebSocketMessage(msg) {
    if (msg.type === 'new_detection_data') {
        addDetectionData(msg.payload);
    } else if (msg.type === 'get_history_response' && msg.payload && msg.payload.history) {
        for (let sess of msg.payload.history) {
            addDetectionData(sess);
        }
    } else if (msg.type === 'error') {
        console.warn('Server error:', msg.payload.message);
    }
}

const ws = new RadarWebSocket(`ws://${window.location.hostname}:8080/ws`, onWebSocketMessage, onPingReceived);
radar.onTargetClick((target) => {
    alert(`ЦЕЛЬ\nID: ${target.detection_id}\nТип: ${target.object_info.type}\nДостоверность: ${(target.object_info.confidence*100).toFixed(1)}%\nАзимут: ${target.location.azimuth}°\nДальность: ${target.location.range/1000} км\nВысота: ${target.location.altitude} м\nРиск: ${target.risk_assessment}`);
});

setInterval(() => {
    radar.draw();
}, 33);