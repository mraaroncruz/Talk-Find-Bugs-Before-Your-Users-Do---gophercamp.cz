import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

const GRID_SIZE = 20;

// Brno city center — grid covers a wider area around the city
const BRNO_CENTER = [49.1951, 16.6068];
const CELL_SIZE = 0.003; // ~330m per cell — covers ~6km across Brno
const GRID_ORIGIN = [
  BRNO_CENTER[0] - (GRID_SIZE / 2) * CELL_SIZE,
  BRNO_CENTER[1] - (GRID_SIZE / 2) * CELL_SIZE,
];

function cellBounds(x, y) {
  const lat = GRID_ORIGIN[0] + y * CELL_SIZE;
  const lng = GRID_ORIGIN[1] + x * CELL_SIZE;
  return [
    [lat, lng],
    [lat + CELL_SIZE, lng + CELL_SIZE],
  ];
}

export default function Grid({ territory, runners, activeRunner }) {
  const mapRef = useRef(null);
  const mapInstanceRef = useRef(null);
  const overlayRef = useRef(null);

  // Initialize map once
  useEffect(() => {
    if (mapInstanceRef.current) return;

    const map = L.map(mapRef.current, {
      center: BRNO_CENTER,
      zoom: 13,
      zoomControl: false,
      attributionControl: false,
    });

    L.tileLayer("https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png", {
      maxZoom: 19,
    }).addTo(map);

    L.control.zoom({ position: "bottomright" }).addTo(map);

    overlayRef.current = L.layerGroup().addTo(map);
    mapInstanceRef.current = map;

    return () => {
      map.remove();
      mapInstanceRef.current = null;
    };
  }, []);

  // Update tile overlays when territory or active runner changes
  useEffect(() => {
    const overlay = overlayRef.current;
    if (!overlay) return;

    overlay.clearLayers();

    const tileMap = {};
    for (const tile of territory) {
      tileMap[`${tile.x},${tile.y}`] = tile;
    }

    for (let y = 0; y < GRID_SIZE; y++) {
      for (let x = 0; x < GRID_SIZE; x++) {
        const tile = tileMap[`${x},${y}`];
        const bounds = cellBounds(x, y);

        if (!tile?.owner_id) continue;

        const runner = runners.find((r) => r.id === tile.owner_id);
        const color = runner?.color || "#6366f1";
        const isActive = tile.owner_id === activeRunner.id;

        L.rectangle(bounds, {
          color: color,
          weight: 0,
          fillColor: color,
          fillOpacity: isActive ? 0.65 : 0.35,
        })
          .bindTooltip(`${tile.owner} (${x},${y})`, { sticky: true })
          .addTo(overlay);
      }
    }
  }, [territory, runners, activeRunner]);

  return (
    <div
      ref={mapRef}
      style={{
        width: "100%",
        height: "60vh",
        borderRadius: "8px",
        overflow: "hidden",
      }}
    />
  );
}
