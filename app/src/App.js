import React, { Component } from 'react';
import numeral from 'numeral';
import $ from 'jquery';
import 'bootstrap';
import {Map} from './Map.js';
import {Graph} from './Graph.js';
import './App.css';
import {congressForYear} from './congress.js';
import {CongressStats} from './stats.js';
import {WindowWidthWatcher} from './windowWidthWatcher.js';

function medianVoters(congress) {
  const factStat = CongressStats[''+congress];
  console.assert(factStat !== undefined);
  return factStat.medianVoters;
}

function medianVotersApprox(congress) {
  const m = medianVoters(congress);
  const factor = 100;
  const tmp = Math.round(m/factor);
  return tmp*factor;
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

    // footnotes
    const whiteList = $.fn.tooltip.Constructor.Default.whiteList;
    whiteList.cite = [];
    $(document).ready(() => {
      $('[data-toggle="tooltip"]').each(function() {
        var $elem = $(this);
        $elem.tooltip({
            html: true,
            container: $elem,
            delay: {hide: 400},
            whiteList: whiteList,
        });
      });
    });
  }

  componentWillUnmount() {
    this.winWatcher.cleanup();
  }

  render() {
    return (
      <div>
        {/* <section className="mx-3">
          <p>You are represented in the House of Representatives
          by <strong>one</strong> person.</p>

          <p>You share this person with about <strong>300,000</strong>
          &nbsp;active voters.</p>

          <p>Is your voice <em>really</em> heard?</p>
        </section> */}

        <section className="container">
          <h2>Intro</h2>

          <p>
            For a representative government to be truly democratic, there must at least be a chance that the
            average citizen could, if needed, get the attention of his or her representative.
            If that representative had, say, ten constituents, then that is certainly the case.
            With 100 or 1,000 it would be a bit harder, but it's doable.
          </p>

          <p>What if that representative had 700,000 constituents?</p>

          <p>
            In 2010 (the last time districts were drawn), the average district size
            was <strong>710,767</strong>. Montana&rsquo;s single district 
            had <strong>994,416</strong> people. With such large districts,
            in what sense does a member of Congress represent anyone?
          </p>

          <h2>How Did It Get This Way?</h2>

          <p>
            Comparing the ways representation in the House has worked over our history is difficult, due to 
            the changing availability of the franchise, the infamous &ldquo;three-fifths&rdquo; clause, and
            the fact that, until 1971, states could have multiple congresspeople per district, or some
            congresspeople assigned to districts and others to the whole state.
          </p>

          <p>
            We get around these problems by focusing on the property of <em>representative burden</em> &mdash; how
            many people was the average congressperson worried about pleasing? &mdash; and just ignoring those
            districts where this property is not defined (e.g., districts with multiple congresspeople).
            We measure representative burden with election turnout: if a congressperson was elected in an election
            that had 20,000 people turn out to vote, we assume that this is the approximate number of people
            that might be inclined to bother the congressperson with their thoughts and concerns. The numbers
            of voters presented below come from turnout and exclude all districts from states that do not use
            single-member districts exclusively.
          </p>

          <p className="lead">
            In the beginning, the House of Representatives expanded
            along with population.
          </p>

          <p>
            Here are the numbers for the first twenty Congresses:
          </p>

          <div className="text-center">
            <Graph width={this.state.figureWidth*1.2} height={(this.state.figureWidth*1.2)/2}
              maxCongress={20} />
          </div>

          <p>
            As measured by turnout, district sizes did tend to increase, but so did the size of the House.
          </p>

          <p>
            Below are maps showing the districts, and their turnouts if you zoom in, for the 1st (1789) and
            15th (1817) Congresses. Zoom in on New York, and you can see how turnout was kept at roughly
            the same level by giving this state more congresspeople.
          </p>

          <div className="d-flex justify-content-center">
            <div className="mx-1">
              <Map congress={1}
                width={this.state.figureWidth}
                zoom={0.75}
                center={[-81.105, 38.807]}
                maxTurnout={800} />
            </div>
            <div className="mx-1">
              <Map congress={15}
                width={this.state.figureWidth}
                zoom={0.75}
                center={[-81.105, 38.807]}
                maxTurnout={800} />
            </div>
          </div>

          <p>
            We don&rsquo;t have turnout data for elections after 1825 and before the 1970s,
            so let&rsquo;s jump to the present day and look at the big picture:
          </p>

          <div className="text-center">
            <Graph width={this.state.figureWidth*1.2} height={(this.state.figureWidth*1.2)/2} />
          </div>

          <p>Two things to note:</p>
          <ul>
            <li>The total number of congresspeople stopped increasing in the early 20th century</li>
            <li>The representative burden in the 20th century is <em>much</em> greater than it was in the
            18th and 19th centuries</li>
          </ul>

          <p className="lead center">
            In <strong>1922</strong>, Congress froze the House at <strong>435</strong> congresspeople.
          </p>

          <p>
            And that&rsquo;s where it has remained, while the population of our country grew.
          </p>

          <p>Let&rsquo;s compare the representative burden of 1817 with that of 2013:</p>

          <div className="d-flex justify-content-center">
            <div className="mx-1">
              <Map congress={congressForYear(1817)}
                width={this.state.figureWidth}
                maxTurnout={400000} />
            </div>
            <div className="mx-1">
              <Map congress={congressForYear(2013)}
                width={this.state.figureWidth}
                maxTurnout={400000} />
            </div>
          </div>

          <p>
            As stated above, the median turnout per congressperson in 1817
            is {numeral(medianVotersApprox(congressForYear(1817))).format('0,0')}.
            In 2013 it was {numeral(medianVotersApprox(congressForYear(2013))).format('0,0')}.
          </p>

          <p>
            The representative burden for members of Congress is now much, much greater than
            it was during the early sessions of Congress, and it has only increased
            since the size of the House was frozen. <strong>The House of Representatives must be
            expanded to match the size of our country, </strong> so that Americans can once again
            have a voice in their government.
          </p>
        </section>

        <section className="container" style={{fontSize: '85%'}}>
          <h2 style={{fontSize: '115%'}}>Sources</h2>

          <p>
            District boundary data comes from <cite><a href="http://cdmaps.polisci.ucla.edu/">Digital Boundary
            Definitions of United States Congressional Districts, 1789-2012</a></cite>, by
            Jeffrey B. Lewis, Brandon DeVine, Lincoln Pitcher, and Kenneth C. Martis.
          </p>

          <p>
            State boundary data comes from <cite><a href="https://publications.newberry.org/ahcbp/index.html">
            Atlas of Historical County Boundaries</a></cite>, by The Newberry.
          </p>

          <p>
            18th- and 19th-century turnout data comes from <cite><a href="https://elections.lib.tufts.edu/">
            Lampi Collection of American Electoral Returns, 1787–1825</a></cite>, by the American
            Antiquarian Society.
          </p>

          <p>
            20th- and 21st-century turnout data comes from <cite><a href="https://doi.org/10.7910/DVN/IG0UN2">
            U.S. House 1976–2018</a></cite>, by the MIT Election Data and Science Lab.
          </p>
        </section>
      </div>
    );
  }
}

export default App;
