// parse args
const args = process.argv.slice(2);
if (args.length != 2) {
  process.stderr.write('usage: node process.js INPUT_DATA OUTPUT_FILE\n');
  process.exit(1);
}
const inputFile = args[0];
const outputFile = args[1];

const fs = require('fs'),
    fiveColorMap = require('five-color-map'),
    turf = require('@turf/turf');

const fipsToState = {};
for (const rec of JSON.parse(fs.readFileSync('states.json', 'utf8'))) {
  fipsToState[rec['FIPS']] = rec;
}

/*
Each feature has an ID property with this format:
SSSBBBEEEDD where SSS is the state FIPS code, BBB is the number of first 
Congress in which that district was used, EEE is the last Congress in 
which that district was used and DD is the district number.
*/
function parseId(featureId) {
  return {
    'stateFips': parseInt(featureId.substring(0, 3)),
    'congress': parseInt(featureId.substring(3, 6)),
    'district': parseInt(featureId.substring(10, 12)),
  };
}

// turns 1 into '1st', etc.
function ordinal(number) {
  if (number < 0) {
    throw 'Ordinal got negative number';
  }
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

// load the congressional district data
process.stdout.write("Reading " + inputFile + "\n");
var districts = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// clean up districts' properties
districts.features.forEach(function(d) {
  const distId = d.properties.ID;
  const {stateFips, congress, district} = parseId(distId);
  d.properties = {
    'id': distId,
    'district': district,
    'congress': congress,
    'stateFips': stateFips,
    'state': fipsToState[stateFips]['USPS'],
    'group': 'boundary',
  };
});

// filter out invalid districts (e.g., -1 for Indian lands)
districts.features = districts.features.filter(function(d) {
  return d.properties.district >= 0;
});

// add titles
districts.features.forEach(function(d) {
  const district = d.properties.district;
  const titleShortNbr = district === 0 ? "At Large" : district;
  const titleLongNbr = district === 0 ? "At Large" : ordinal(district);
  const stateName = fipsToState[d.properties.stateFips]['Name'];
  d.properties.titleShort = `${d.properties.state} ${titleShortNbr}`;
  d.properties.titleLong = `${stateName}'s ${titleLongNbr} Congressional District`;
});

// use the five-color-map package to assign color numbers to each
// congressional district so that no two touching districts are
// assigned the same color number
districts = fiveColorMap(districts);

// write out data for the next steps
console.log('writing data...');
fs.writeFileSync(outputFile, JSON.stringify(districts, undefined, 1));
console.log('finished processing, ready for tiling');
