import React, { Component } from 'react';
import _ from 'lodash';
import {Map} from './Map.js';
import CONGRESS_NBR_TO_STYLE_ID from './congressNbrToStyleId';
import {ordinal} from './utils.js';
import './App.css';

class App extends Component {
  constructor(props) {
    super(props);

    this.sortedCongressNbrs = 
      _.sortBy(_.keys(CONGRESS_NBR_TO_STYLE_ID).map(n => parseInt(n)));

    this.state = {congress: this.sortedCongressNbrs[0]};

    this.handleSelectCongress = this.handleSelectCongress.bind(this);
  }

  handleSelectCongress(event) {
    const newCongress = parseInt(event.target.value);
    this.setState({congress: newCongress});
  }

  render() {
    const pickerOpts = this.sortedCongressNbrs.map(n => {
      return (<option key={n} value={n}>{ordinal(n)} Congress</option>);
    });
    const picker = (
      <select value={this.state.congress} onChange={this.handleSelectCongress}>
        {pickerOpts}}
      </select>
    );

    return (
      <div className="App">
        {picker}
        <Map congress={this.state.congress} />
      </div>
    );
  }
}

export default App;
