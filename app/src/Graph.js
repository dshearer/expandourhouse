import React, { Component } from 'react';
import * as d3 from 'd3';
import {CongressStats} from './stats.js';
import {startYearForCongress} from './congress.js';
import './Graph.css';
const _ = require('lodash');

function isPresElectYear(year) {
    const delta = year - 1789;
    return delta%4 == 0;
}

const gMargin = {top: 10, right: 50, bottom: 40, left: 70};

class Graph extends Component {
    componentDidMount() {
        const adjWidth = 500 - gMargin.left - gMargin.right;
        const adjHeight = 250 - gMargin.top - gMargin.bottom;

        // prepare data
        var minCongress = this.props.minCongress;
        if (minCongress === undefined) {
            minCongress = 1
        }
        var maxCongress = this.props.maxCongress;
        if (maxCongress === undefined) {
            maxCongress = 1000;
        }
        const data = _.keys(CongressStats)
            .filter(congress => parseInt(congress) >= minCongress)
            .filter(congress => parseInt(congress) <= maxCongress)
            .map(congress => {
                const rec = CongressStats[congress];
                const year = startYearForCongress(parseInt(congress));
                rec.year = d3.timeParse('%Y')(year);
                rec.isPresElect = isPresElectYear(year);
                return rec;
            });

        // make the lines
        const x = d3.scaleUtc()
            .domain(d3.extent(data, d => d.year))
            .range([ 0, adjWidth ]);
        const votersY = d3.scaleLinear()
            .domain([0, d3.max(data, d => d.voters)]).nice()
            .range([ adjHeight, 0 ]);
        const nbrRepsY = d3.scaleLinear()
            .domain([0, d3.max(data, d => d.nbrReps)]).nice()
            .range([ adjHeight, 0 ]);
        const nbrRepsLine = d3.line()
            .defined(d => d.nbrReps !== undefined)
            .x(d => x(d.year))
            .y(d => nbrRepsY(d.nbrReps));

        const svg = d3.select(this.axisYear);

        // make the axes
        d3.select(this.axisYear).call(d3.axisBottom(x));
        d3.select(this.axisVoters).call(d3.axisLeft(votersY));
        d3.select(this.axisNbrReps).call(d3.axisLeft(nbrRepsY));

        const votersLine = d3.line()
            .defined(d => d.voters !== undefined)
            .x(d => x(d.year))
            .y(d => votersY(d.voters));
        d3.select(this.lineVotersMissing)
            .datum(data.filter(votersLine.defined()))
            .attr("d", votersLine);
        d3.select(this.lineVoters)
            .datum(data)
            .attr("d", votersLine);

        d3.select(this.lineNbrRepsMissing)
            .datum(data.filter(nbrRepsLine.defined()))
            .attr("d", nbrRepsLine);
        d3.select(this.lineNbrReps)
            .datum(data)
            .attr("d", nbrRepsLine);
    }

    render() {
        const adjWidth = 500 - gMargin.left - gMargin.right;
        const adjHeight = 250 - gMargin.top - gMargin.bottom;

        return (
            <svg width={this.props.width} 
                height={this.props.height} viewBox="0 0 500 250">
                <g transform={`translate(${gMargin.left}, ${gMargin.top})`}>
                    <g transform={`translate(0, ${adjHeight})`} 
                        ref={el => this.axisYear = el} />

                    <g className="axisVoters" ref={el => this.axisVoters = el} />

                    <g className="axisNbrReps" 
                        transform={`translate(${adjWidth}, 0)`} 
                        ref={el => this.axisNbrReps = el} />

                    <text className="axisVotersLabel" 
                        transform="rotate(-90)"
                        textAnchor="middle"
                        x={-adjHeight/2}
                        y={-gMargin.left}
                        dy="1em"># of voters per rep</text>

                    <text className="axisNbrRepsLabel" 
                        transform="rotate(90)"
                        textAnchor="middle"
                        x={adjHeight/2}
                        y={-adjWidth - 35}
                        dy="1em"># of reps</text>

                    <path className="lineMissing" fill="none"
                        ref={el => this.lineVotersMissing = el} />
                    <path className="lineVoters" fill="none"
                        ref={el => this.lineVoters = el} />
                    
                    <path className="lineMissing" fill="none"
                        ref={el => this.lineNbrRepsMissing = el} />
                    <path className="lineNbrReps" fill="none"
                        ref={el => this.lineNbrReps = el} />
                </g>
            </svg>
        );
    }

}

export {Graph};