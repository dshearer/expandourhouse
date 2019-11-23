import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';
import api from './api';
import {ShowErrors, getAxiosErrors} from './Alerts';

mapboxgl.accessToken = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

class Map extends Component {
    constructor(props) {
        super(props);

        this.state = {
            districts: {},
            errors: [],
        };

        this.handleMapClick = this.handleMapClick.bind(this);
    }

    componentDidMount() {
        api.get('/districts').then(
            resp => {
                console.log(resp.data);
                this.setState({districts: resp.data, errors: []});
            },
            errResp => {
                this.setState({errors: getAxiosErrors(errResp)});
            },
        ).catch(console.log);

        this.map = new mapboxgl.Map({
            container: this.mapContainer,
            style: 'mapbox://styles/dshearer/ck2jt3oqi330w1cp92ww1hl5z'
        });
        this.map.on('click', this.handleMapClick);
    }

    componentWillUnmount() {
        this.map.remove();
    }

    handleMapClick(e) {
        // get features at click location from district layers
        var features = this.map.queryRenderedFeatures(
            e.point,
            { layers: ["districts_1", "districts_2", "districts_3", "districts_4", "districts_5"] }
        );
        if (features.length === 0) {
            return;
        }
        features = features[0];
        console.log(features);

        // get clicked district
        const state = features.properties.state;
        const district = features.properties.number;

        // Ensure that if the map is zoomed out such that multiple
        // copies of the feature are visible, the popup appears
        // over the copy being pointed to.
        const coords = features.geometry.coordinates.slice();
        while (Math.abs(e.lngLat.lng - coords[0]) > 180) {
            coords[0] += e.lngLat.lng > coords[0] ? 360 : -360;
        }
        const key = `${state}${district}`;
        const data = this.state.districts[key];
        if (data === undefined) {
            console.log("No data for " + key);
            return;
        }
        const html = `<div>Pop: ${numberWithCommas(data.Pop)}<br />CVAP: ${numberWithCommas(data.Cvap)}</div>`;
        console.log(coords);
        
        new mapboxgl.Popup()
            .setLngLat(e.lngLat)
            .setHTML(html)
            .addTo(this.map);
    }

    render() {
        const style = {
            position: 'absolute',
            top: 0,
            bottom: 0,
            width: '100%'
        };

        return (
            <div>
                <ShowErrors errors={this.state.errors} />
                <div style={style} ref={el => this.mapContainer = el} />
            </div>
        );
    }
}

export {Map};