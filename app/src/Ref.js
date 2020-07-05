import React, { Component } from 'react';

const gRefs = {
    '2010CongApp': 'U.S. Census Bureau (2011). <cite>Congressional Apportionment: 2010 Census Brief</cite>. ' +
        'Retrieved from <a href="https://www.census.gov/library/publications/2011/dec/c2010br-08.html">' +
        'https://www.census.gov/library/publications/2011/dec/c2010br-08.html</a>',
};

class Ref extends Component {
    render() {
        const ref = gRefs[this.props.name];
        if (!ref) {
            console.error("No such ref: " + this.props.name);
            return null;
        }
        return (
            <span className="footnote" data-toggle="tooltip" title={ref} />
        );
    }
}

export {Ref};
