import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';
import {ShowErrors} from './Alerts';
import {ordinal} from './utils.js';
import CONGRESS_NBR_TO_STYLE_ID from './congressNbrToStyleId';

const MAPBOX_TOKEN = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

mapboxgl.accessToken = MAPBOX_TOKEN;

function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

const gFirstCongressStartYear = 1789;

function congressStartYear(congress) {
    return gFirstCongressStartYear + 2*(congress-1);
}

class Map extends Component {
    constructor(props) {
        super(props);

        this.state = {
            congressStartYear: undefined,
            errors: [],
        };

        this.handleMapClick = this.handleMapClick.bind(this);
    }

    styleUrlForCongress(congress) {
        return 'mapbox://styles/dshearer/' +
            CONGRESS_NBR_TO_STYLE_ID['' + congress];
    }

    loadMap() {
        console.log("Loading map: " + this.props.congress);
        if (!this.props.congress) {
            return;
        }
        if (!this.map) {
            this.map = new mapboxgl.Map({
                container: this.mapContainer,
                style: this.styleUrlForCongress(this.props.congress),
            });
            this.map.on('click', this.handleMapClick);
        } else {
            this.map.setStyle(this.styleUrlForCongress(this.props.congress));
        }
    }

    componentDidMount() {
        mapboxgl.clearStorage();
        this.loadMap();
    }

    componentWillUnmount() {
        if (this.map) {
            this.map.remove();
        }
    }

    componentDidUpdate(prevProps) {
        console.log(this.props);
        console.log('componentDidUpdate ' + this.props.congress + ' ' + prevProps.congress);
        if (this.props.congress === prevProps.congress) {
            return;
        }
        this.loadMap();
    }

    handleMapClick(e) {
        // get features at click location from district layers
        const features = this.map.queryRenderedFeatures(
            e.point,
            { layers: ["irreg-district-bg", "reg-district-bg"]},
        );
        console.log(features);
        if (features.length === 0) {
            return;
        }
        const feature = features[0];
        console.log('feature:');
        console.log(feature);

        // get clicked district
        const districtTitle = feature.properties.titleLong;
        const turnout = feature.properties.turnout;

        if (!turnout) {
            return;
        }

        // Ensure that if the map is zoomed out such that multiple
        // copies of the feature are visible, the popup appears
        // over the copy being pointed to.
        const coords = feature.geometry.coordinates.slice();
        while (Math.abs(e.lngLat.lng - coords[0]) > 180) {
            coords[0] += e.lngLat.lng > coords[0] ? 360 : -360;
        }

        // make popup
        const html = `
        <aside>
            <header>
            ${districtTitle}
            </header>
            <table>
                <tr><th scope="row">Turnout</th><td>${numberWithCommas(turnout)}</td></tr>
            </table>
        </aside>
        `;
        new mapboxgl.Popup()
            .setLngLat(e.lngLat)
            .setHTML(html)
            .addTo(this.map);
    }

    render() {
        const containerStyle = {
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'stretch',
            margin: '2em',
        };
        const mapStyle = {
            // width: '1000px',
            height: '600px',
            margin: 0,
        };
        const captionStyle = {
            textAlign: 'center',
            marginBottom: '1em',
        };

        return (
            <div>
                <ShowErrors errors={this.state.errors} />
                <div style={containerStyle}>
                    <figure style={mapStyle} ref={el => this.mapContainer = el}>
                        <figcaption style={captionStyle}>
                            {ordinal(this.props.congress)} Congress ({congressStartYear(this.props.congress)})
                        </figcaption>
                    </figure>
                </div>
            </div>
        );
    }
}

export {Map};