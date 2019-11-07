import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';

mapboxgl.accessToken = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

class Map extends Component {
    componentDidMount() {
        this.map = new mapboxgl.Map({
            container: this.mapContainer,
            style: 'mapbox://styles/dshearer/ck2jt3oqi330w1cp92ww1hl5z'
        });
    }

    componentWillUnmount() {
        this.map.remove();
    }

    render() {
        const style = {
            position: 'absolute',
            top: 0,
            bottom: 0,
            width: '100%'
        };

        return (<div style={style} ref={el => this.mapContainer = el} />);
    }
}

export {Map};