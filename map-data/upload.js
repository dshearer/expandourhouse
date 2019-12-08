const fs = require('fs'),
  AWS = require('aws-sdk'),
  mbxClient = require('@mapbox/mapbox-sdk'),
  mbxStyles = require('@mapbox/mapbox-sdk/services/styles'),
  mbxUploads = require('@mapbox/mapbox-sdk/services/uploads'),
  _ = require('lodash');

// parse args
const usage = 'usage: node upload.js (style STYLE_FILE MAPBOX_USER | ' +
  'states STYLE_FILE STATES_TILESET | ' +
  'districts STYLE_FILE DISTRICTS_TILESET)';
const args = process.argv.slice(2);
if (args.length < 1) {
  process.stderr.write(usage + '\n');
  process.exit(1);
}
const cmd = args[0];

// get MapBox access token
const ACCESS_TOKEN_ENV_VAR = 'MAPBOX_WRITE_SCOPE_ACCESS_TOKEN';
const accessToken = process.env[ACCESS_TOKEN_ENV_VAR];
if (accessToken === undefined) {
  process.stderr.write(`Missing env var: ${ACCESS_TOKEN_ENV_VAR}\n`);
  process.exit(1);
}

const baseClient = mbxClient({ accessToken: accessToken }),
  stylesService = mbxStyles(baseClient),
  uploadsService = mbxUploads(baseClient);

function doUploadStyle() {
  // parse args
  if (args.length != 3) {
    process.stderr.write(usage + '\n');
    process.exit(1);
  }
  const styleFile = args[1];
  const user = args[2];

  // get some stuff from the style
  const style = JSON.parse(fs.readFileSync(styleFile, 'utf-8'));
  const styleName = style.name;

  // look for existing style
  return stylesService.listStyles({ownerId: user}).send()
    .then(function(resp) {
      console.log("Uploading style");

      // get ID of style with name styleName
      const styles = resp.body.filter(style => style.name === styleName);
      if (styles.length > 0) {
        // update existing style
        const styleId = styles[0].id;
        return stylesService.updateStyle({styleId: styleId, style: style, ownerId: user}).send()
          .then(() => console.log("Updated existing style"));
      }
      else {
        // create new style
        return stylesService.createStyle({style: style, ownerId: user}).send()
          .then(() => console.log("Created new style"));
      }
    });
}

function doUploadTilesets(statesOrDistricts) {
  // parse args
  if (args.length != 3) {
    process.stderr.write(usage + '\n');
    process.exit(1);
  }
  const styleFile = args[1];
  const tilesetFile = args[2];

  // get some stuff from the style
  const style = JSON.parse(fs.readFileSync(styleFile, 'utf-8'));
  const tilesetId = style.metadata[`dshearer:${statesOrDistricts}-tileset-id`];
  const tilesetName = style.metadata[`dshearer:${statesOrDistricts}-tileset-name`];

  // upload
  return uploadsService.createUploadCredentials().send()
    .then(resp => resp.body)
    .then(creds => {
      // Use aws-sdk to stage the file on Amazon S3
      console.log("Uploading tileset to AWS");
      const s3 = new AWS.S3({
           accessKeyId: creds.accessKeyId,
           secretAccessKey: creds.secretAccessKey,
           sessionToken: creds.sessionToken,
           region: 'us-east-1'
      });
      const req = s3.putObject({
        Bucket: creds.bucket,
        Key: creds.key,
        Body: fs.createReadStream(tilesetFile),
      });
      req.on('httpUploadProgress', progress => {
        var pc = _.floor(100 * progress.loaded / progress.total);
        console.log("Uploaded " + pc + "%");
      });
      return req.promise().then(() => creds);
    })
    .then(creds => {
      // create mapbox upload
      console.log("Sending uploaded tileset to MapBox");
      return uploadsService.createUpload({
        mapId: tilesetId,
        url: creds.url,
        tilesetName: tilesetName,
      }).send();
    });
}

var f;
if (cmd === 'style') {
  f = doUploadStyle;
} else if (cmd === 'states' || cmd === 'districts') {
  f = () => doUploadTilesets(cmd);
} else {
  process.stderr.write(usage + '\n');
  process.exit(1);
}
f().catch(e => {
  console.log(e);
  process.exit(1);
});
