/*
This script adds a label for each feature in the given GeoJSON doc where
group == 'boundary'.
*/

const fs = require('fs');
const turf = require('@turf/turf');

const BOUNDARY_GROUP = 'boundary';
const LABEL_GROUP = 'label';

// parse args
const args = process.argv.slice(2);
if (args.length != 1) {
  process.stderr.write('usage: node add-labels.js INPUT_GEOJSON\n');
  process.exit(1);
}
const inputFile = args[0];

// read file
const data = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// organize boundaries by ID
const boundariesById = new Map();
data.features.forEach((f, idx) => {
  if (f.properties.group !== BOUNDARY_GROUP) {
    return;
  }

  if (f.properties.boundaryId === undefined) {
    f.properties.boundaryId = '' + idx;
  }
  boundariesById.set(f.properties.boundaryId, f);
});

// forget boundaries that already have labels
data.features.forEach((f, idx) => {
  if (f.properties.group !== LABEL_GROUP) {
    return;
  }

  boundariesById.delete(f.properties.boundaryId);
});

// make labels for remaining boundaries
boundariesById.forEach((f, id) => {
  if (f.geometry === null) {
    return;
  }
  const labelPoint = turf.centerOfMass(f);
  Object.assign(labelPoint.properties, f.properties);
  labelPoint.properties.group = 'label';
  data.features.push(labelPoint);
});

console.log(JSON.stringify(data, undefined, 1));
