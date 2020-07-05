import React, { Component } from 'react';
import _ from 'lodash';
import * as d3 from 'd3';
import {MIN_TURNOUT, Heat} from './heat.js';

const NBR_SLOTS = 10;
const MARGIN = 25;
const BORDER_WIDTH = 1;

class HeatKey extends Component {
    constructor(props) {
        super(props);

        this.turnouts = [];
        const range = this.props.maxTurnout - MIN_TURNOUT;
        const delta = range/NBR_SLOTS;
        for (let i = 0; i <= NBR_SLOTS; ++i) {
            const t = i * delta;
            this.turnouts.push(t);
        }
    }

    componentDidMount() {
        this.drawNumberLine();
    }

    drawNumberLine() {
        if (!this.g) {
            return;
        }

        // remove existing SVG
        const g = d3.select(this.g);
        g.selectAll('*').remove();

        // draw it
        const turnouts = d3.scaleLinear()
            .domain([d3.min(this.turnouts), d3.max(this.turnouts)])
            .range([MARGIN + BORDER_WIDTH, this.props.width - MARGIN - BORDER_WIDTH]);
        g.call(d3.axisBottom(turnouts));
    }

    render() {
        // make background gradient
        const heat = new Heat(this.props.maxTurnout);
        const colors = this.turnouts.map(t => `hsl(${heat.hue(t)}, 100%, 50%)`);
        const bgColorCss = `linear-gradient(to right, ${_.join(colors, ', ')})`;
        const keyStyle = {
            background: bgColorCss,
            border: `${BORDER_WIDTH}px solid black`,
            height: '10px',
            marginLeft: MARGIN,
            marginRight: MARGIN,
        };
        this.drawNumberLine();
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
