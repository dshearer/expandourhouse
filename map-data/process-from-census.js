// parse args
const args = process.argv.slice(2);
if (args.length != 2) {
  process.stderr.write('usage: node process.js INPUT_DATA OUTPUT_FILE\n');
  process.exit(1);
}
const inputFile = args[0];
const outputFile = args[1];

var fs = require('fs'),
    fiveColorMap = require('five-color-map'),
    turf = require('@turf/turf');

const fipsToState = {};
for (const rec of JSON.parse(fs.readFileSync('states.json', 'utf8'))) {
  fipsToState[rec['FIPS']] = rec;
}
// load the congressional district data

process.stdout.write("Reading " + inputFile + "\n");
var geojson = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// figure out name of field that holds district ID (it contains the ordinal
// of the session of congress)
const districtIdField = `CD${geojson.features[0].properties['CDSESSN']}FP`;

// some states have district 'ZZ' which represents the area of
// a state, usually over water, that is not included in any
// congressional district --- filter these out

var filtered = geojson.features.filter(function(d) {
  return d.properties[districtIdField] !== 'ZZ' ? true : false;
});
var districts = { 'type': 'FeatureCollection', 'features': filtered };

// use the five-color-map package to assign color numbers to each
// congressional district so that no two touching districts are
// assigned the same color number

var colored = fiveColorMap(districts);

// turns 1 into '1st', etc.
function ordinal(number) {
  const suffixes = ['th', 'st', 'nd', 'rd', 'th', 'th', 'th', 'th', 'th', 'th'];
  var suffix;
  if (((number % 100) == 11) || ((number % 100) == 12) || ((number % 100) == 13)) {
    suffix = suffixes[0];
  }
  else {
    suffix = suffixes[number % 10];
  }
  return `${number}${suffix}`;
}

// add additional metadata to the GeoJSON for rendering later and
// compute bounding boxes for each congressional district and each
// state so that we know how to center and zoom maps

var districtBboxes = {},
    stateBboxes = {};

// empty FeatureCollection to contain final map data
var mapData = { 'type': 'FeatureCollection', 'features': [] }

colored.features.map(function(d) {

  // Census TIGER files have INTPTLON/INTPTLAT which conveniently
  // provides a point where a label for the polygon can be placed.
  // Create a turf.point to hold information for rending labels.
  var pt = turf.point([parseFloat(d.properties['INTPTLON']), parseFloat(d.properties['INTPTLAT'])]);

  // Get the district number in two-digit form ("00" (for at-large
  // districts), "01", "02", ...). The Census data's CD114FP field
  // holds it in this format. Except for the island territories
  // which have "98", but are just at-large and should be "00".
  var number = d.properties[districtIdField];
  if (number === undefined) {
    throw "Doesn't have number: " + d.properties['NAMELSAD'];
  }
  if (number == "98") {
    number = "00";
  }
  // map the state FIPS code in the STATEFP attribute to the USPS
  // state abbreviation and the state's name
  const stateFips = parseInt(d.properties['STATEFP']);
  const state = fipsToState[stateFips]['USPS'];
  const state_name = fipsToState[stateFips]['Name'];

  // add the district number and USPS state code to the metadata
  d.properties.number = number;
  d.properties.state = state;

  // add metadata to the label
  pt.properties = JSON.parse(JSON.stringify(d.properties)); // copy hack to avoid mutability issues
  const titleShortNbr = number == "00" ? "At Large" : parseInt(number);
  pt.properties.title_short = `${state} ${titleShortNbr}`;
  const titleLongNbr = number == "00" ? "At Large" : ordinal(parseInt(number));
  pt.properties.title_long = `${state_name}'s ${titleLongNbr} Congressional District`;

  // add a type property to distinguish between labels and boundaries
  pt.properties.group = 'label';
  d.properties.group = 'boundary';

  // add both the label point and congressional district to the mapData feature collection
  mapData.features.push(pt);
  mapData.features.push(d);

  // collect bounding boxes for the districts
  var bounds = turf.bbox(d);
  districtBboxes[state + number] = bounds;

  // and for the states
  if (stateBboxes[state]) {
    stateBboxes[state].features.push(turf.bboxPolygon(bounds));
  } else {
    stateBboxes[state] = { type: 'FeatureCollection', features: [] };
    stateBboxes[state].features.push(turf.bboxPolygon(bounds));
  }
});

// get the bounding boxes of all of the bounding boxes for each state
for (var s in stateBboxes) {
  stateBboxes[s] = turf.bbox(stateBboxes[s]);
}

// write out data for the next steps
console.log('writing data...');

fs.writeFileSync(outputFile, JSON.stringify(mapData, undefined, 1));

/*
fs.writeFileSync('./example/states.js', 'var states = ' + JSON.stringify(stateCodes, null, 2));

var bboxes = {};
for (var b in districtBboxes) { bboxes[b] = districtBboxes[b] };
for (var b in stateBboxes) { bboxes[b] = stateBboxes[b] };
fs.writeFileSync('./example/bboxes.js', 'var bboxes = ' + JSON.stringify(bboxes, null, 2));
*/

console.log('finished processing, ready for tiling');
