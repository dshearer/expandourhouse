const fs = require('fs');

const CONGRESS_YEARS_FILE = '../backend/loaddata/data/congress-start-years.txt';

// parse args
const args = process.argv.slice(2);
if (args.length != 1) {
  process.stderr.write('usage: node congress-start-year.js CONGRESS_NBR\n');
  process.exit(1);
}
const congressNbr = parseInt(args[0], 10);

// read congress start years
const congressStartYears = new Map();
fs.readFileSync(CONGRESS_YEARS_FILE, 'utf-8')
    .split('\n')
    .forEach((line, idx) => {
        if (line.trim().length == 0) {
            return;
        }
        congressStartYears.set(idx + 1, parseInt(line.trim()));
    });

// get start year
const startYear = congressStartYears.get(congressNbr);
if (startYear === undefined) {
    process.stderr.write("Unknown congress\n");
    process.exit(1);
}
console.log(startYear);
