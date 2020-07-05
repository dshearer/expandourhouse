import React, { Component } from 'react';
import _ from 'lodash';
import * as d3 from 'd3';
import $ from 'jquery';
import {COOLEST_HUE, HOTTEST_HUE, MIN_TURNOUT, MAX_TURNOUT, hue} from './heat.js';

const NBR_SLOTS = 8;
const TURNOUTS = [];
for (let i = 0; i < NBR_SLOTS; ++i) {
    const range = MAX_TURNOUT - MIN_TURNOUT;
    const delta = range/NBR_SLOTS;
    const t = (i + 1) * delta;
    TURNOUTS.push(t);
}

class HeatKey extends Component {
    constructor(props) {
        super(props);
        this.state = {width: 0};
    }

    componentDidMount() {
        console.assert(this.g);
        const turnouts = d3.scaleLinear()
            .domain([d3.min(TURNOUTS), d3.max(TURNOUTS)])
            .range([0, this.props.width]);
        d3.select(this.g).call(d3.axisBottom(turnouts));
    }

    componentWillUnmount() {
        window.removeEventListener("resize", this.updateSize);
    }

    updateSize = () => {
        if (!this.container) {
            return;
        }
        this.setState({ width: $(this.container).width() });
    };

    render() {
        // make background gradient
        const colors = TURNOUTS.map(t => `hsl(${hue(t)}, 100%, 50%)`);
        const bgColorCss = `linear-gradient(to right, ${_.join(colors, ', ')})`;
        const keyStyle = {
            background: bgColorCss,
            border: '1px solid black',
            height: '10px',
        };
        return (
            <div style={{width: this.props.width}}>
                <div className="heat-key" style={keyStyle} />
                <svg className="heat-key-legend" width={this.props.width} height={20}>
                    <g ref={el => this.g = el} />
                </svg>
            </div>
        );
    }
}

export {HeatKey};
