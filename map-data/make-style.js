const fs = require('fs');

const STYLE_TEMPLATE_FILE = 'style-template.json';

// parse args
const args = process.argv.slice(2);
if (args.length != 2) {
  process.stderr.write('usage: node make-style.js CONGRESS_NBR MAPBOX_USER\n');
  process.exit(1);
}
const congressNbr = parseInt(args[0], 10);
const user = args[1];

// make tileset IDs and names
/* ID: max 32 characters, only one period */
const statesTilesetId = `${user}.states-${congressNbr}`;
const districtsTilesetId = `${user}.districts-${congressNbr}`;
/* Name: max 64 characters, no spaces */
const statesTilesetName = `US_States_${congressNbr}`;
const districtsTilesetName = `US_Districts_${congressNbr}`;

// make style
const style = JSON.parse(fs.readFileSync(STYLE_TEMPLATE_FILE, 'utf-8'));
style.name = `Geo for Congress ${congressNbr}`;
style.glyphs = `mapbox://fonts/${user}/{fontstack}/{range}.pbf`;
style.sources.states = {url: `mapbox://${statesTilesetId}`, type: 'vector'};
style.sources.districts = {url: `mapbox://${districtsTilesetId}`, type: 'vector'};
style.metadata['dshearer:states-tileset-id'] = statesTilesetId;
style.metadata['dshearer:districts-tileset-id'] = districtsTilesetId;
style.metadata['dshearer:states-tileset-name'] = statesTilesetName;
style.metadata['dshearer:districts-tileset-name'] = districtsTilesetName;

console.log(JSON.stringify(style, undefined, 0));