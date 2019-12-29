const fs = require('fs');

// parse args
const args = process.argv.slice(2);
if (args.length != 2) {
  process.stderr.write('usage: node extract-states-for-year.js INPUT_GEOJSON YYYY\n');
  process.exit(1);
}
const inputFile = args[0];
var forYear = parseInt(args[1], 10);
if (forYear > 1980) {
  forYear = 1980;
}
const forDate = new Date(`${forYear}/01/01`);
process.stderr.write("Getting state boundaries for " + forDate + "\n");

// load file
const data = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// filter boundaries by date
data.features = data.features.filter(f => {
  if (!f.properties.START_DATE || !f.properties.END_DATE) {
      return false;
  }
  const startDate = new Date(f.properties.START_DATE);
  const endDate = new Date(f.properties.END_DATE);
  return startDate <= forDate && forDate <= endDate;
});

// add 'group', 'id', 'titleLong', and 'titleShort' properties
data.features.forEach(f => {
  f.properties.group = 'boundary';
  f.properties.id = f.properties.ID;
  f.properties.ID = undefined;
  f.properties.titleLong = f.properties.FULL_NAME;
  f.properties.titleShort = f.properties.ABBR_NAME;
});

console.log(JSON.stringify(data, undefined, 1));