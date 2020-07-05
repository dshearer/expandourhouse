import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';
import numeral from 'numeral';
import {Heat} from './heat.js';
import {HeatKey} from './HeatKey.js';
import './Map.css';

const MAPBOX_TOKEN = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

mapboxgl.accessToken = MAPBOX_TOKEN;

const gNbrVotersLayerName = 'nbr-voters';
const gHighlightStateLayerName = 'highlight-state';

const gDistrictBoundaryLayer = {
    "id": "district-boundary",
    "type": "line",
    "metadata": {"mapbox:group": "1444934295202.7542"},
    "source": "districts",
    "source-layer": "districts",
    "filter": [
        "all",
        [
          "match",
          ["get", "group"],
          ["boundary"],
          true,
          false
        ]
      ],
    "layout": {"line-join": "round", "line-cap": "round"},
    "paint": {
        "line-dasharray": [
            "step",
            ["zoom"],
            ["literal", [2, 0]],
            7,
            ["literal", [2, 2, 6, 2]]
        ],
        "line-width": [
            "interpolate",
            ["linear"],
            ["zoom"],
            7,
            0.75,
            12,
            1.5
        ],
        "line-opacity": ["step", ["zoom"], 0, 5, 1],
        "line-color": [
            "interpolate",
            ["linear"],
            ["zoom"],
            3,
            "hsl(0, 0%, 80%)",
            7,
            "hsl(0, 0%, 70%)"
        ]
    }
};

const gStateBoundaryLayer = {
    "id": "admin-1-boundary",
    "type": "line",
    "metadata": {"mapbox:group": "1444934295202.7542"},
    "source": "states",
    "source-layer": "states",
    "filter": [
        "all",
        [
            "match",
            ["get", "group"],
            ["boundary"],
            true,
            false
        ]
        ],
    "layout": {"line-join": "round", "line-cap": "round"},
    "paint": {
        "line-dasharray": [
            "step",
            ["zoom"],
            ["literal", [2, 0]],
            7,
            ["literal", [2, 2, 6, 2]]
        ],
        "line-width": 2,
        "line-opacity": [
            "interpolate",
            ["linear"],
            ["zoom"],
            2,
            0,
            3,
            1
        ],
        "line-color": [
            "interpolate",
            ["linear"],
            ["zoom"],
            3,
            "hsl(0, 0%, 80%)",
            7,
            "hsl(0, 0%, 70%)"
        ]
    }
};

const gStateBoundaryBgLayer = {
    "id": "admin-1-boundary-bg",
    "type": "line",
    "metadata": {"mapbox:group": "1444934295202.7542"},
    "source": "states",
    "source-layer": "states",
    "filter": [
        "all",
        [
            "match",
            ["get", "group"],
            ["boundary"],
            true,
            false
        ]
        ],
    "layout": {"line-join": "bevel"},
    "paint": {
        "line-blur": ["interpolate", ["linear"], ["zoom"], 3, 0, 8, 3],
        "line-width": [
            "interpolate",
            ["linear"],
            ["zoom"],
            7,
            3.75,
            12,
            5.5
        ],
        "line-opacity": [
            "interpolate",
            ["linear"],
            ["zoom"],
            7,
            0,
            8,
            0.75
        ],
        "line-dasharray": [1, 0],
        "line-translate": [0, 0],
        "line-color": "hsl(0, 0%, 84%)"
    }
};

const gStateLabelLayer = {
    "id": "state-label",
    "type": "symbol",
    "source": "states",
    "source-layer": "states",
    "minzoom": 3,
    "maxzoom": 5,
    "filter": [
        "all",
        [
            "match",
            ["get", "group"],
            ["label"],
            true,
            false
        ]
        ],
    "layout": {
        "text-size": [
            "interpolate",
            ["cubic-bezier", 0.85, 0.7, 0.65, 1],
            ["zoom"],
            4,
            ["step", ["get", "symbolrank"], 10, 6, 9.5, 7, 9],
            9,
            ["step", ["get", "symbolrank"], 24, 6, 18, 7, 14]
        ],
        "text-transform": "uppercase",
        "text-font": ["DIN Offc Pro Bold", "Arial Unicode MS Bold"],
        "text-field": ["get", "titleShort"],
        "text-letter-spacing": 0.15,
        "text-max-width": 6
    },
    "paint": {
        "text-halo-width": 1,
        "text-halo-color": "hsl(0, 0%, 100%)",
        "text-color": "hsl(0, 0%, 66%)"
    }
};

function nbrVotersLayer(state) {
    const layer = {
        "id": gNbrVotersLayerName,
        "type": "symbol",
        "metadata": {"mapbox:group": "1444934295202.7542"},
        "source": "districts",
        "source-layer": "districts",
        "minzoom": 5,
        "maxzoom": 9,
        "filter": [
            "all",
            [
              "match",
              ["get", "group"],
              ["label"],
              true,
              false
            ],
            ["has", "turnout"]
          ],
        "layout": {
            "text-size": [
                "interpolate",
                [
                  "cubic-bezier",
                  0.85,
                  0.7,
                  0.65,
                  1
                ],
                ["zoom"],
                5,
                10,
                9,
                [
                  "step",
                  ["get", "symbolrank"],
                  24,
                  3,
                  14,
                  6,
                  18
                ]
              ],
            "text-transform": "none",
            "text-font": ["DIN Offc Pro Regular", "Arial Unicode MS Regular"],
            "text-field": [
                "step",
                ["zoom"],
                "",
                5,
                ["get", "turnoutStr"]
              ],
            "text-letter-spacing": 0.15,
            "text-max-width": 10,
            "text-allow-overlap": true,
            "text-offset": [0, 0.75]
        },
        "paint": {
            "text-halo-width": 1,
            "text-halo-color": "hsl(0, 0%, 100%)",
            "text-color": "hsl(0, 0%, 29%)"
        }
    };
    if (state) {
        layer.filter.push(["match", ["get", "state"], [state], true, false]);
    }
    return layer;
}

function districtBackgroundLayer(hueStyleFormula) {
    const layer = {
        "id": gHighlightStateLayerName,
        "type": "fill",
        "source": "districts",
        "source-layer": "districts",
        "filter": [
            "all",
            [
              "match",
              ["get", "group"],
              ["boundary"],
              true,
              false
            ],
            ["has", "turnout"],
        ],
        "paint": {
            "fill-color": [
                "concat",
                "hsl(",
                ["to-string", hueStyleFormula],
                ", 100%, 50%)",
            ],
            "fill-opacity": 0.2,
            "fill-antialias": false
        },
    };
    return layer;
}

class Map extends Component {
    getZoom() {
        var v = 0.0044 * this.props.width + 0.432;
        if (this.props.zoom) {
            v += this.props.zoom;
        }
        return v;
    }

    componentDidUpdate() {
        if (this.map) {
            this.map.resize();
            this.map.setMinZoom(this.getZoom());
            this.map.setZoom(this.getZoom());
        }
    }

    componentDidMount() {
        // mapboxgl.clearStorage();

        const opts = {
            container: this.mapContainer,
            maxPitch: 0,
            touchPitch: false,
            dragRotate: false,
            center: [-96.429, 38.3],
            zoom: this.getZoom(),
            minZoom: this.getZoom(),
        };
        if (this.props.center) {
            opts.center = this.props.center;
        }
        this.map = new mapboxgl.Map(opts);

        this.map.setStyle('mapbox://styles/dshearer/ckbu2kkj1030b1io3jfo1ykiq');
        
        this.map.on('style.load', () => {
            const statesSource = `mapbox://dshearer.states-${this.props.congress}`;
            const districtsSource = `mapbox://dshearer.districts-${this.props.congress}`;

            this.map.addSource('states', {
                type: 'vector',
                url: statesSource,
            });
            this.map.addSource('districts', {
                type: 'vector',
                url: districtsSource,
            });

            this.map.addLayer(gStateLabelLayer);
            this.map.addLayer(gStateBoundaryBgLayer);
            this.map.addLayer(gStateBoundaryLayer);
            this.map.addLayer(gDistrictBoundaryLayer);
            this.map.addLayer(districtBackgroundLayer(new Heat(this.props.maxTurnout).styleFormula()));
            this.map.addLayer(nbrVotersLayer());
    
            this.highlightedState = this.props.highlightState;
        });

        if (!window.maps) {
            window.maps = [];
        }
        window.maps.push(this.map);
    }

    componentWillUnmount() {
        if (this.map) {
            this.map.remove();
            this.map = undefined;
        }
    }

    render() {
        const dimensRatio = 2; // width/height
        const height = this.props.width/dimensRatio;
        const figureStyle = {width: this.props.width};
        const mapStyle = {width: this.props.width, height: height};
        return (
            <figure className="map figure" style={figureStyle}>
                <div ref={el => this.mapContainer = el} style={mapStyle} />
                <HeatKey width={this.props.width} maxTurnout={this.props.maxTurnout} />
                <figcaption className="figure-caption text-center">
                Heatmap of number of voters per district for
                the {numeral(this.props.congress).format('0o')} Congress. (Zoom in to see
                the numbers.)</figcaption>
            </figure>
        );
    }
}

export {Map};