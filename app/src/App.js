import React, { Component } from 'react';
import numeral from 'numeral';
import {MapWithPicker} from './MapWithPicker.js';
import {Map} from './Map.js';
import {Graph} from './Graph.js';
import './App.css';
import {startYearForCongress, congressForYear} from './congress.js';
import {CongressStats} from './stats.js';
import {WindowWidthWatcher} from './windowWidthWatcher.js';

function medianVoters(congress) {
  const factStat = CongressStats[''+congress];
  console.assert(factStat !== undefined);
  return factStat.voters;
}

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {figureWidth: 500};

    this.updateFigureWidth = this.updateFigureWidth.bind(this);
  }

  updateFigureWidth(maxWidth) {
    const padding = 30;
    const w = Math.min(maxWidth, 500) - padding;
    this.setState({figureWidth: w});
  }

  componentDidMount() {
    this.winWatcher = new WindowWidthWatcher([320, 360, 375, 411, 414]);
    this.winWatcher.addListener(this.updateFigureWidth);
  }

  componentWillUnmount() {
    this.winWatcher.cleanup();
  }

  render() {
    const factClasses = "d-flex align-items-center flex-column flex-lg-row";
    return (
      <div>
        <section className="mx-3">
          <p>You are represented in the House of Representatives
          by <strong>one</strong> person.</p>

          <p>You share this person with about <strong>300,000</strong>
          &nbsp;active voters.</p>

          <p>Is your voice <em>really</em> heard?</p>
        </section>

        <section className="d-flex align-items-center flex-column">
          <h2 className="mx-3">How Did It Get This Way?</h2>

          <div className={factClasses}>
            <p style={{maxWidth: this.state.figureWidth}}>
              In the beginning, the House of Representatives expanded
            along with population.</p>
            <div>
              <Graph width={this.state.figureWidth} height={this.state.figureWidth/2} maxCongress={20} />
            </div>
          </div>

          <div className={factClasses}>
            <p style={{maxWidth: this.state.figureWidth}}>
              In 1796, the median number of voters per congressperson
              was {numeral(medianVoters(congressForYear(1797))).format('0,0')}.</p>
            <div>
              <Map congress={congressForYear(1797)} width={this.state.figureWidth} />
            </div>
          </div>

          <div className={factClasses}>
          <p style={{maxWidth: this.state.figureWidth}}>
            In 1816, the median number of voters per congressperson
            was {numeral(medianVoters(congressForYear(1817))).format('0,0')}.</p>
            <div>
              <Map congress={congressForYear(1817)} width={this.state.figureWidth} />
            </div>
          </div>

          <div className={factClasses}>
            <p style={{maxWidth: this.state.figureWidth}}>
              But in <strong>1922</strong>, Congress froze the House
              at <strong>435</strong> congresspeople. That&rsquo;s where it
              is <strong>today</strong>.</p>
            <div>
              <Graph width={500} height={250} />
            </div>
          </div>

          <div className={factClasses}>
            <p style={{maxWidth: this.state.figureWidth}}>
            In 1976, the median number of voters per congressperson
            was {numeral(medianVoters(congressForYear(1977))).format('0,0')}.</p>
            <div>
              <Map congress={congressForYear(1977)} width={this.state.figureWidth} />
            </div>
          </div>

          <div className={factClasses}>
            <p style={{maxWidth: this.state.figureWidth}}>
            In 2004, the median number of voters per congressperson
            was {numeral(medianVoters(congressForYear(2005))).format('0,0')}.</p>
            <div>
              <Map congress={congressForYear(2005)} width={this.state.figureWidth} />
            </div>
          </div>

        </section>
      </div>
    );
  }
}

export default App;
