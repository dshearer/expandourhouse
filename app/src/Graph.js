import React, { Component } from 'react';
import * as d3 from 'd3';
import STATS from './stats.js';
import {congressStartYear} from './congress.js';
import './Graph.css';
const _ = require('lodash');

class Graph extends Component {
    componentDidMount() {
        const margin = {top: 10, right: 50, bottom: 40, left: 70},
            width = 1060 - margin.left - margin.right,
            height = 400 - margin.top - margin.bottom;

        // prepare data
        const data = _.keys(STATS).map(congress => {
            const rec = STATS[congress];
            rec.year = d3.timeParse('%Y')(congressStartYear(parseInt(congress)));
            return rec;
        });

        // make the lines
        const x = d3.scaleUtc()
            .domain(d3.extent(data, d => d.year))
            .range([ 0, width ]);
        const votersY = d3.scaleLinear()
            .domain([0, d3.max(data, d => d.voters)]).nice()
            .range([ height, 0 ]);
        const nbrRepsY = d3.scaleLinear()
            .domain([0, d3.max(data, d => d.nbrReps)]).nice()
            .range([ height, 0 ]);
        const votersLine = d3.line()
            .defined(d => d.voters !== undefined)
            .x(d => x(d.year))
            .y(d => votersY(d.voters));
        const nbrRepsLine = d3.line()
            .defined(d => d.nbrReps !== undefined)
            .x(d => x(d.year))
            .y(d => nbrRepsY(d.nbrReps));

        const svg = d3.select(this.container)
            .append("svg")
            .attr("width", width + margin.left + margin.right)
            .attr("height", height + margin.top + margin.bottom)
            .append("g")
            .attr("transform",  "translate(" + margin.left + "," + margin.top + ")");

        // make the axes
        svg.append("g")
            .attr("transform", "translate(0," + height + ")")
            .call(d3.axisBottom(x));
        svg.append("g")
            .attr("class", "axisVoters")
            .call(d3.axisLeft(votersY));
        svg.append("g")
            .attr("class", "axisNbrReps")
            .attr("transform", "translate(" + width + ", 0)")
            .call(d3.axisRight(nbrRepsY));

        // make the axis labels
        svg.append("text")             
            .attr("transform",
                    "translate(" + (width/2) + " ," + 
                                (height + margin.top + 25) + ")")
            .style("text-anchor", "middle")
            .text("Year");
        svg.append("text")
            .attr("class", "axisVotersLabel")
            .attr("transform", "rotate(-90)")
            .attr("y", 0 - margin.left)
            .attr("x", 0 - (height / 2))
            .attr("dy", "1em")
            .style("text-anchor", "middle")
            .text("Nbr of voters per district");    
        svg.append("text")
            .attr("class", "axixNbrRepsLabel")
            .attr("transform", "rotate(90)")
            .attr("y", -width - 35)
            .attr("x", height / 2)
            // .attr("dy", "1em")
            .style("text-anchor", "middle")
            .text("Nbr of congresspeople");     

        svg.append("path")
            .datum(data.filter(votersLine.defined()))
            .attr("class", "lineMissing")
            .attr("fill", "none")
            .attr("d", votersLine);

        svg.append("path")
            .datum(data)
            .attr("class", "lineVoters")
            .attr("fill", "none")
            .attr("d", votersLine);

        svg.append("path")
            .datum(data.filter(nbrRepsLine.defined()))
            .attr("class", "lineMissing")
            .attr("fill", "none")
            .attr("d", nbrRepsLine);

        svg.append("path")
            .datum(data)
            .attr("class", "lineNbrReps")
            .attr("fill", "none")
            .attr("d", nbrRepsLine);
    }

    render() {
        return (<div ref={el => this.container = el}></div>);
    }

}

export {Graph};