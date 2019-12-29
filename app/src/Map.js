import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';
import api from './api';
import {ShowErrors, getAxiosErrors} from './Alerts';
import CONGRESS_NBR_TO_STYLE_ID from './congressNbrToStyleId';

const MAPBOX_TOKEN = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

mapboxgl.accessToken = MAPBOX_TOKEN;

function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

// turns 1 into '1st', etc.
function ordinal(number) {
    if (number < 0) {
      throw 'Ordinal got negative number';
    }
    const suffixes = ['th', 'st', 'nd', 'rd', 'th', 'th', 'th', 'th', 'th', 'th'];
    var suffix;
    if (((number % 100) == 11) || ((number % 100) == 12) || ((number % 100) == 13)) {
      suffix = suffixes[0];
    }
    else {
      suffix = suffixes[number % 10];
    }
    return `${number}${suffix}`;
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

    componentDidMount() {
        // get congress info
        api.get(`/congresses`).then(
            resp => {
                console.log(resp.data);
                const startYear = resp.data['' + this.props.congressNbr].startYear;
                this.setState({congressStartYear: startYear, errors: []});
            },
            errResp => {
                this.setState({errors: getAxiosErrors(errResp)});
            },
        ).catch(e => console.log(e));

        // make map object
        const styleUrl = 'mapbox://styles/dshearer/' +
            CONGRESS_NBR_TO_STYLE_ID['' + this.props.congressNbr];
        this.map = new mapboxgl.Map({
            container: this.mapContainer,
            // style: 'mapbox://styles/dshearer/ck3xoweef0c8k1cp75twv29le', // for 100th congress
            style: styleUrl,
        });
        this.map.on('click', this.handleMapClick);
    }

    componentWillUnmount() {
        this.map.remove();
    }

    handleMapClick(e) {
        // get features at click location from district layers
        const features = this.map.queryRenderedFeatures(
            e.point,
            { layers: ["districts_1", "districts_2", "districts_3", "districts_4", "districts_5"] }
        );
        if (features.length === 0) {
            return;
        }
        const feature = features[0];
        console.log('feature:');
        console.log(feature);

        // get clicked district
        const state = feature.properties.state;
        const district = feature.properties.district;
        const districtTitle = feature.properties.titleLong;

        // Ensure that if the map is zoomed out such that multiple
        // copies of the feature are visible, the popup appears
        // over the copy being pointed to.
        const coords = feature.geometry.coordinates.slice();
        while (Math.abs(e.lngLat.lng - coords[0]) > 180) {
            coords[0] += e.lngLat.lng > coords[0] ? 360 : -360;
        }

        // get district facts
        api.get(`/congresses/${this.props.congressNbr}/states/${state}/districts/${district}`).then(
            resp => {
                console.log(resp.data);
                const tableRows = [];
                for (const factType in resp.data) {
                    const fact = resp.data[factType];
                    const value = numberWithCommas(fact.value);
                    tableRows.push(`<tr><th scope="row">${factType}</th><td>${value}</td></tr>`);
                }
                const table = "<table>" + tableRows.join() + "</table>";
                const html = `
                <aside>
                    <header>
                    ${districtTitle}
                    </header>
                    ${table}
                </aside>
                `;
                new mapboxgl.Popup()
                    .setLngLat(e.lngLat)
                    .setHTML(html)
                    .addTo(this.map);
            },
            errResp => {
                this.setState({errors: getAxiosErrors(errResp)});
            },
        ).catch(e => console.log(e));
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
            height: '500px',
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
                            {ordinal(this.props.congressNbr)} Congress ({this.state.congressStartYear})
                        </figcaption>
                    </figure>
                </div>
            </div>
        );
    }
}

export {Map};