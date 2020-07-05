import React, { Component } from 'react';
import mapboxgl from 'mapbox-gl';
import numeral from 'numeral';
import _ from 'lodash';
import CONGRESS_NBR_TO_STYLE_ID from './congressNbrToStyleId';
import {startYearForCongress} from './congress.js';
import {Map} from './Map.js';
import './Map.css';

const MAPBOX_TOKEN = 'pk.eyJ1IjoiZHNoZWFyZXIiLCJhIjoiY2syam1qaThuMTEzazNsbnZxNHhidnZqcyJ9.Q0wOV0EePfEaRyw1oEK3UA';

mapboxgl.accessToken = MAPBOX_TOKEN;

class MapWithPicker extends Component {
    constructor(props) {
        super(props);

        this.sortedCongressNbrs = 
          _.sortBy(_.keys(CONGRESS_NBR_TO_STYLE_ID).map(n => parseInt(n)));

        this.state = {
            congress: this.sortedCongressNbrs[0],
        };

        this.handleSelectCongress = this.handleSelectCongress.bind(this);
    }

    handleSelectCongress(event) {
        const newCongress = parseInt(event.target.value);
        this.setState({congress: newCongress});
    }

    render() {
        const pickerOpts = this.sortedCongressNbrs.map(n => {
          return (<option key={n} value={n}>{numeral(n).format('0o')} Congress ({startYearForCongress(n)})</option>);
        });
        const picker = (
          <select value={this.state.congress} onChange={this.handleSelectCongress}>
            {pickerOpts}}
          </select>
        );

        return (
            <div>
                {picker}
                <div style={{height: '500px', width: '700px'}}>
                    <Map congress={this.state.congress} />
                </div>
            </div>
        );
    }
}

export {MapWithPicker};