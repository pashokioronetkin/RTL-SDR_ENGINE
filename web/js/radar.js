export class Radar {
    constructor(canvasId) {
        this.canvas = document.getElementById(canvasId);
        this.ctx = this.canvas.getContext('2d');
        this.width = 800;
        this.height = 800;
        this.canvas.width = this.width;
        this.canvas.height = this.height;

        this.detections = [];
        this.detectionMap = new Map();

        this.offsetX = 0;
        this.offsetY = 0;
        this.scale = 1.0;
        this.BASE_RADIUS_PX = 370;
        this.MAX_RANGE_KM = 20;

        this.isDragging = false;
        this.dragStart = { x: 0, y: 0 };

        // FPS счётчик
        this.frameCount = 0;
        this.lastTime = performance.now();
        this.fps = 0;
        this.fpsElement = document.getElementById('fpsValue');

        this.initControls();
    }

    setDetections(detections) {
        this.detections = detections;
        this.detectionMap.clear();
        for (let d of detections) {
            this.detectionMap.set(d.detection_id, d);
        }
        this.draw();
    }

    addDetection(det) {
        if (this.detectionMap.has(det.detection_id)) {
            const existing = this.detectionMap.get(det.detection_id);
            Object.assign(existing, det);
            return false;
        }
        this.detectionMap.set(det.detection_id, det);
        this.detections.push(det);
        return true;
    }

    polarToScreen(azimuthDeg, rangeMeters) {
        const rangeKm = rangeMeters / 1000;
        const radiusPx = (rangeKm / this.MAX_RANGE_KM) * this.BASE_RADIUS_PX * this.scale;
        const angleRad = (azimuthDeg - 90) * Math.PI / 180;
        const cx = this.width/2 + this.offsetX;
        const cy = this.height/2 + this.offsetY;
        return {
            x: cx + radiusPx * Math.cos(angleRad),
            y: cy + radiusPx * Math.sin(angleRad)
        };
    }

    draw() {
        if (!this.ctx) return;
        this.ctx.clearRect(0, 0, this.width, this.height);
        const cx = this.width/2 + this.offsetX;
        const cy = this.height/2 + this.offsetY;

        this.ctx.fillStyle = '#010a07';
        this.ctx.fillRect(0, 0, this.width, this.height);
        this.ctx.beginPath();
        this.ctx.arc(cx, cy, this.BASE_RADIUS_PX * this.scale, 0, 2*Math.PI);
        this.ctx.fillStyle = '#05100c';
        this.ctx.fill();
        this.ctx.strokeStyle = '#3faa7a';
        this.ctx.lineWidth = 1.8;
        this.ctx.stroke();

        const ranges = [2,5,10,20];
        for (let km of ranges) {
            const rad = (km / this.MAX_RANGE_KM) * this.BASE_RADIUS_PX * this.scale;
            this.ctx.beginPath();
            this.ctx.arc(cx, cy, rad, 0, 2*Math.PI);
            this.ctx.strokeStyle = '#2f8068';
            this.ctx.setLineDash([5,8]);
            this.ctx.stroke();
            this.ctx.fillStyle = '#ccffcc';
            this.ctx.font = '12px monospace';
            if (km === 20) this.ctx.fillText(km+'km', cx+rad-30, cy-4);
            else this.ctx.fillText(km+'km', cx+rad+5, cy-4);
        }
        this.ctx.setLineDash([]);

        for (let deg=0; deg<360; deg+=30) {
            const radAng = (deg-90)*Math.PI/180;
            const x2 = cx + (this.BASE_RADIUS_PX * this.scale) * Math.cos(radAng);
            const y2 = cy + (this.BASE_RADIUS_PX * this.scale) * Math.sin(radAng);
            this.ctx.beginPath();
            this.ctx.moveTo(cx, cy);
            this.ctx.lineTo(x2, y2);
            this.ctx.strokeStyle = '#2a6a54';
            this.ctx.stroke();
            this.ctx.fillStyle = '#b3ffcf';
            this.ctx.font = '11px monospace';
            let offset = 12;
            let xLabel = x2 + Math.cos(radAng) * offset;
            let yLabel = y2 + Math.sin(radAng) * offset;
            this.ctx.fillText(deg+'°', xLabel, yLabel);
        }

        this.ctx.beginPath();
        this.ctx.arc(cx, cy, 5, 0, 2*Math.PI);
        this.ctx.fillStyle = '#9fffc0';
        this.ctx.fill();
        this.ctx.beginPath();
        this.ctx.arc(cx, cy, 2, 0, 2*Math.PI);
        this.ctx.fillStyle = '#000';
        this.ctx.fill();

        for (let det of this.detections) {
            const {x, y} = this.polarToScreen(det.location.azimuth, det.location.range);
            if (x<0 || x>this.width || y<0 || y>this.height) continue;
            let color = '#ffcc66';
            if (det.risk_assessment === 'high') color = '#ff6666';
            else if (det.risk_assessment === 'low') color = '#66ff99';
            else color = '#ffaa44';
            this.ctx.beginPath();
            this.ctx.arc(x, y, 8, 0, 2*Math.PI);
            this.ctx.fillStyle = color;
            this.ctx.fill();
            this.ctx.strokeStyle = '#ffffff';
            this.ctx.lineWidth = 1.2;
            this.ctx.stroke();
            this.ctx.fillStyle = '#ffffff';
            this.ctx.font = `bold ${Math.max(12, 13*this.scale)}px monospace`;
            this.ctx.fillText(det.object_info.type, x+12, y-6);
        }

        // Обновление FPS после каждого кадра
        this.frameCount++;
        const now = performance.now();
        const delta = now - this.lastTime;
        if (delta >= 1000) {
            this.fps = Math.round((this.frameCount * 1000) / delta);
            if (this.fpsElement) this.fpsElement.textContent = this.fps;
            this.frameCount = 0;
            this.lastTime = now;
        }
    }

    onTargetClick(callback) {
        this.canvas.addEventListener('click', (e) => {
            const rect = this.canvas.getBoundingClientRect();
            const sx = this.canvas.width / rect.width;
            const sy = this.canvas.height / rect.height;
            const mouseX = (e.clientX - rect.left) * sx;
            const mouseY = (e.clientY - rect.top) * sy;
            let best = 20;
            let selected = null;
            for (let det of this.detections) {
                const {x, y} = this.polarToScreen(det.location.azimuth, det.location.range);
                const dist = Math.hypot(mouseX - x, mouseY - y);
                if (dist < best) { best = dist; selected = det; }
            }
            if (selected) callback(selected);
        });
    }

    initControls() {
        this.canvas.addEventListener('mousedown', (e) => {
            this.isDragging = true;
            const rect = this.canvas.getBoundingClientRect();
            const sx = this.canvas.width / rect.width;
            const sy = this.canvas.height / rect.height;
            this.dragStart.x = (e.clientX - rect.left) * sx;
            this.dragStart.y = (e.clientY - rect.top) * sy;
            this.canvas.style.cursor = 'grabbing';
        });
        window.addEventListener('mousemove', (e) => {
            if (!this.isDragging) return;
            const rect = this.canvas.getBoundingClientRect();
            const sx = this.canvas.width / rect.width;
            const sy = this.canvas.height / rect.height;
            const curX = (e.clientX - rect.left) * sx;
            const curY = (e.clientY - rect.top) * sy;
            this.offsetX += curX - this.dragStart.x;
            this.offsetY += curY - this.dragStart.y;
            this.dragStart.x = curX;
            this.dragStart.y = curY;
            this.draw();
        });
        window.addEventListener('mouseup', () => {
            this.isDragging = false;
            this.canvas.style.cursor = 'crosshair';
        });
        this.canvas.addEventListener('wheel', (e) => {
            e.preventDefault();
            let delta = e.deltaY > 0 ? -0.1 : 0.1;
            let ns = this.scale + delta;
            if (ns < 0.3) ns = 0.3;
            if (ns > 3.0) ns = 3.0;
            this.scale = ns;
            this.draw();
        });
    }
}