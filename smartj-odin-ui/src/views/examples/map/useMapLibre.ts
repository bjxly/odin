import { onBeforeUnmount, type Ref } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

const TIANDITU_KEY = '845a8b9646865970deb015e815052b36'
const SUBDOMAINS = [0, 1, 2, 3, 4, 5, 6, 7]

export type TiandituLayerType = 'vec' | 'img' | 'ter'

interface TiandituConfig {
  base: string
  annotation: string
}

const TIANDITU_LAYERS: Record<TiandituLayerType, TiandituConfig> = {
  vec: {
    base: `https://t{s}.tianditu.gov.cn/vec_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=vec&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
    annotation: `https://t{s}.tianditu.gov.cn/cva_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cva&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
  },
  img: {
    base: `https://t{s}.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
    annotation: `https://t{s}.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
  },
  ter: {
    base: `https://t{s}.tianditu.gov.cn/ter_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=ter&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
    annotation: `https://t{s}.tianditu.gov.cn/cta_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cta&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=${TIANDITU_KEY}`,
  },
}

export interface MapOptions {
  center?: [number, number]
  zoom?: number
  pitch?: number
  projection?: maplibregl.ProjectionSpecification | string
}

export function useMapLibre(containerRef: Ref<HTMLElement | undefined>, options: MapOptions = {}) {
  let map: maplibregl.Map | null = null

  function buildStyle(layerType: TiandituLayerType = 'vec', projection?: string): maplibregl.StyleSpecification {
    const config = TIANDITU_LAYERS[layerType]
    const style: maplibregl.StyleSpecification = {
      version: 8,
      sources: {
        'tianditu-base': {
          type: 'raster',
          tiles: [config.base.replace('{s}', String(SUBDOMAINS[Math.floor(Math.random() * SUBDOMAINS.length)]))],
          tileSize: 256,
          attribution: '天地图',
        },
        'tianditu-annotation': {
          type: 'raster',
          tiles: [config.annotation.replace('{s}', String(SUBDOMAINS[Math.floor(Math.random() * SUBDOMAINS.length)]))],
          tileSize: 256,
          attribution: '天地图',
        },
      },
      layers: [
        {
          id: 'tianditu-base-layer',
          type: 'raster',
          source: 'tianditu-base',
          minzoom: 0,
          maxzoom: 18,
        },
        {
          id: 'tianditu-annotation-layer',
          type: 'raster',
          source: 'tianditu-annotation',
          minzoom: 0,
          maxzoom: 18,
        },
      ],
    }
    if (projection) {
      ;(style as any).projection = { type: projection }
    }
    return style
  }

  function initMap(layerType: TiandituLayerType = 'img'): maplibregl.Map {
    if (!containerRef.value) throw new Error('Map container not found')

    const style = buildStyle(layerType, options.projection as string | undefined)

    map = new maplibregl.Map({
      container: containerRef.value,
      style,
      center: options.center || [104, 35],
      zoom: options.zoom || 4,
      pitch: options.pitch || 0,
      attributionControl: false,
    })

    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), 'top-right')

    return map
  }

  function switchLayer(layerType: TiandituLayerType) {
    if (!map) return
    const config = TIANDITU_LAYERS[layerType]

    const baseSource = map.getSource('tianditu-base') as maplibregl.RasterTileSource
    const annoSource = map.getSource('tianditu-annotation') as maplibregl.RasterTileSource

    if (baseSource) {
      baseSource.setTiles([config.base.replace('{s}', String(SUBDOMAINS[Math.floor(Math.random() * SUBDOMAINS.length)]))])
    }
    if (annoSource) {
      annoSource.setTiles([config.annotation.replace('{s}', String(SUBDOMAINS[Math.floor(Math.random() * SUBDOMAINS.length)]))])
    }
  }

  function zoomIn() {
    map?.zoomIn()
  }

  function zoomOut() {
    map?.zoomOut()
  }

  function cleanup() {
    if (map) {
      map.remove()
      map = null
    }
  }

  onBeforeUnmount(() => {
    cleanup()
  })

  return {
    initMap,
    switchLayer,
    zoomIn,
    zoomOut,
    cleanup,
    getMap: () => map,
  }
}
